// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"
)

// GetFilter retrieves filter based configuration settings, split on the | character
func (c *Config) GetFilter(filterName string) []string {
	if len(c.GetString(filterName)) == 0 {
		return nil
	}
	// Determine if | or ues provided.  Either is acceptable.
	if strings.Contains(c.GetString(filterName), "|") {
		return strings.Split(c.GetString(filterName), "|")
	}
	return strings.Split(c.GetString(filterName), ",")
}
