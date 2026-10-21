//go:build cfappsauth

// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// See manager_test.go for why this package's tests need the cfappsauth build tag
// and NRF_* placeholder env vars.
package cfapps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
)

const selfHealTestGUID = "test-guid"

// newMockUAA simulates UAA issuing a short-lived access token on password grant, then
// rejecting the subsequent refresh-token exchange with the same signature-verification
// error UAA returns after a signing-key rotation (newrelic-pcf-nozzle-tile#81):
// `oauth2: "invalid_token" "Could not verify token signature."`
func newMockUAA(tokenTTLSeconds int) *httptest.Server {
	var passwordGrants int32
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		if r.FormValue("grant_type") == "refresh_token" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":             "invalid_token",
				"error_description": "Could not verify token signature.",
			})
			return
		}
		// password grant: used both for the initial login and for UpdateClient's relogin.
		n := atomic.AddInt32(&passwordGrants, 1)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  fmt.Sprintf("token-%d", n),
			"token_type":    "bearer",
			"refresh_token": fmt.Sprintf("refresh-%d", n),
			"expires_in":    tokenTTLSeconds,
		})
	})
	return httptest.NewServer(mux)
}

// newMockCFAPI serves just enough of the CF API for a client to bootstrap
// (/v2/info) and fetch one known app by GUID. Only an exact match is registered
// for the app GUID path so that sibling endpoints the real nozzle also calls in the
// background (instances, env) fall through to the mux's default 404 instead of
// matching as a subtree and returning a response those callers don't expect.
func newMockCFAPI(uaaURL string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"authorization_endpoint": uaaURL,
			"token_endpoint":         uaaURL,
		})
	})
	mux.HandleFunc("/v2/apps/"+selfHealTestGUID, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	})
	return httptest.NewServer(mux)
}

// TestFetchAppSelfHealsAfterTokenSignatureError reproduces the production failure mode
// from newrelic-pcf-nozzle-tile#81: the cached token stops verifying partway through its
// validity window (as happens when UAA rotates its signing key), and proves FetchApp
// detects it via isAuthError and re-authenticates automatically instead of requiring a
// restart.
func TestFetchAppSelfHealsAfterTokenSignatureError(t *testing.T) {
	// x/oauth2 treats a token as expired 10s before its actual Expiry (defaultExpiryDelta),
	// so the TTL needs enough headroom above that for the first fetch to see a token that's
	// still considered valid.
	const tokenTTLSeconds = 15
	uaa := newMockUAA(tokenTTLSeconds)
	defer uaa.Close()
	cf := newMockCFAPI(uaa.URL)
	defer cf.Close()

	a := app.Get()
	a.Config.Set("CF_API_URL", cf.URL)
	a.Config.Set("CF_API_UAA_URL", uaa.URL)
	a.Config.Set("CF_API_USERNAME", "admin")
	a.Config.Set("CF_API_PASSWORD", "admin")
	a.Config.Set("CF_SKIP_SSL", true)

	manager := Start(a)
	cfApp := NewCFApp(selfHealTestGUID)

	if err := manager.FetchApp(cfApp); err != nil {
		t.Fatalf("expected initial fetch to succeed, got: %v", err)
	}

	// Cross the (TTL - defaultExpiryDelta) mark so the token source considers the cached
	// token expired and attempts a refresh.
	time.Sleep((tokenTTLSeconds - 10 + 1) * time.Second)

	err := manager.FetchApp(cfApp)
	if err == nil || !isAuthError(err) {
		t.Fatalf("expected a signature/auth error once the token expired, got: %v", err)
	}

	// UpdateClient() runs in a goroutine (triggered by the error above); poll until
	// the fresh login it performs completes.
	deadline := time.Now().Add(3 * time.Second)
	for {
		err = manager.FetchApp(cfApp)
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("nozzle did not self-heal within the deadline, last error: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
