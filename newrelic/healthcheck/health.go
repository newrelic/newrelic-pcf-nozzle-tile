// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package healthcheck

import (
	"fmt"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"net/http"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
)

// Start creates a HTTP server that listens and responds to /health requests
func Start() {
	go func() {
		http.HandleFunc("/health", healthCheckHandler)
		app.Get().Log.Fatal(http.ListenAndServe(":"+config.Get().GetString("HEALTH_PORT"), nil))
	}()
}

// healthCheckHandler defines the response for requests to /health endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I'm alive and well!")
}
