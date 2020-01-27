// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
)

// Client is the PCF API Client
type Client struct {
	Client       *cfclient.Client
	UAARefresher *UAATokenRefresher
}

// New API Client
func New() (c *Client, err error) {

	config := app.Get().Config

	c = &Client{}

	c.Client, err = cfclient.NewClient(&cfclient.Config{
		ApiAddress:        config.GetString("CF_API_URL"),
		ClientID:          config.GetString("CF_CLIENT_ID"),
		ClientSecret:      config.GetString("CF_CLIENT_SECRET"),
		SkipSslValidation: config.GetBool("CF_SKIP_SSL"),
	})

	if err != nil {
		app.Get().Log.Errorf("failed to connect to PCF client: %s", err.Error())
	}

	c.UAARefresher, err = NewUAATokenRefresher(
		config.GetString("CF_API_UAA_URL"),
		config.GetString("CF_CLIENT_ID"),
		config.GetString("CF_CLIENT_SECRET"),
		config.GetBool("CF_SKIP_SSL"),
	)

	return c, err

}
