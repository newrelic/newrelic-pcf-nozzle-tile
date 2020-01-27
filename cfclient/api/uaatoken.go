// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"

	"github.com/cloudfoundry-incubator/uaago"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
)

// UAATokenRefresher ...
type UAATokenRefresher struct {
	url               string
	clientID          string
	clientSecret      string
	skipSSLValidation bool
	client            *uaago.Client
}

// NewUAATokenRefresher ...
func NewUAATokenRefresher(authEndpoint string,
	clientID string,
	clientSecret string,
	skipSSLValidation bool,
) (*UAATokenRefresher, error) {
	client, err := uaago.NewClient(authEndpoint)
	if err != nil {
		return &UAATokenRefresher{}, err
	}

	return &UAATokenRefresher{
		url:               authEndpoint,
		clientID:          clientID,
		clientSecret:      clientSecret,
		skipSSLValidation: skipSSLValidation,
		client:            client,
	}, nil
}

// RefreshAuthToken ...
func (uaa *UAATokenRefresher) RefreshAuthToken() (string, error) {
	authToken, err := uaa.client.GetAuthToken(uaa.clientID, uaa.clientSecret, uaa.skipSSLValidation)
	if err != nil {
		app.Get().Log.Error(
			fmt.Sprintf(
				"Error getting oauth token: %s. Please check your Client ID and Secret.",
				err.Error(),
			))
		return "", err
	}
	return authToken, nil
}
