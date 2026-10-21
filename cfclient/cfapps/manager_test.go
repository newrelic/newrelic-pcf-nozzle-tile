//go:build cfappsauth

// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Tests in this package require the cfappsauth build tag: cfapp.go reads
// app.Get().Config at package-init time, which calls os.Exit if the NRF_*
// config env vars aren't already set in the process environment. Run with:
//
//	NRF_CF_API_URL=x NRF_CF_API_UAA_URL=x NRF_CF_CLIENT_ID=x NRF_CF_CLIENT_SECRET=x \
//	NRF_CF_API_USERNAME=x NRF_CF_API_PASSWORD=x NRF_NEWRELIC_INSERT_KEY=x NRF_NEWRELIC_ACCOUNT_ID=x \
//	go test -tags=cfappsauth ./cfclient/cfapps/...
//
// (see "make authtest")
package cfapps

import (
	"errors"
	"testing"
)

func TestIsAuthError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"plain 401", errors.New("CF api error Error requesting apps: 401 Unauthorized on GUID abc"), true},
		{
			"uaa signature failure after key rotation",
			errors.New(`CF api error Error requesting apps: Get "https://api.example.com/v2/apps/abc": oauth2: "invalid_token" "Could not verify token signature." on GUID abc`),
			true,
		},
		{
			"unrelated network error",
			errors.New(`CF api error Error requesting apps: Get "https://api.example.com/v2/apps/abc": dial tcp: connection refused on GUID abc`),
			false,
		},
		{"not found", errors.New("404 Not Found"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAuthError(tc.err); got != tc.want {
				t.Errorf("isAuthError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
