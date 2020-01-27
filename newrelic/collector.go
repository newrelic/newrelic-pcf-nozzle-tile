// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package newrelic

import (
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/registry"
)

// Collector Metrics In-Memory Store
type Collector struct {
	accumulators *registry.Accumulators
}

// NewCollector ...
func NewCollector(r *registry.Accumulators) *Collector {
	collector := Collector{
		accumulators: &registry.Accumulators{},
	}
	for _, i := range *r {
		collector.Append(i.New())
	}
	return &collector
}

// Append ...
func (c *Collector) Append(i accumulators.Interface) {
	*c.accumulators = append(*c.accumulators, i)
}

// Accumulators returns pointer to slice of registered Accumulators
func (c *Collector) Accumulators() *registry.Accumulators {
	return c.accumulators
}

// Length ...
func (c *Collector) Length() int {
	return len(*c.accumulators)
}
