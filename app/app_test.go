// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	os.Setenv("NRF_CF_API_URL", " ")
	os.Setenv("NRF_CF_API_UAA_URL", " ")
	os.Setenv("NRF_CF_CLIENT_ID", " ")
	os.Setenv("NRF_CF_CLIENT_SECRET", " ")
	os.Setenv("NRF_CF_API_USERNAME", " ")
	os.Setenv("NRF_CF_API_PASSWORD", " ")
	os.Setenv("NRF_NEWRELIC_INSERT_KEY", " ")
	os.Setenv("NRF_NEWRELIC_ACCOUNT_ID", " ")
	//Get should return the same object
	assert.Equal(t, Get(), Get())
}
