// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package httpfirehose

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/cfclient/api"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
)

// HttpFirehose Object
type HttpFirehose struct {
	apiClient  *api.Client
	httpClient *http.Client
}

// NewHttpFirehose creates a new object with the correct TLS configuration
func NewHttpFirehose(c *api.Client, conf *config.Config) *HttpFirehose {
	return &HttpFirehose{
		apiClient: c,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: conf.GetBool("CF_SKIP_SSL"),
				},
			},
			Timeout: time.Duration(conf.GetInt("FIREHOSE_HTTP_TIMEOUT_MINS")) * time.Minute,
		},
	}
}

// Do will add the token as an authorization header on all HTTP requests from FirehoseHttp
func (h *HttpFirehose) Do(req *http.Request) (*http.Response, error) {
	token, err := h.apiClient.Client.GetToken()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", token)
	// Connection should stream for up to 14 minutes.
	app.Get().Log.Debugln("Issuing new HTTP firehose request")
	return h.httpClient.Do(req)
}
