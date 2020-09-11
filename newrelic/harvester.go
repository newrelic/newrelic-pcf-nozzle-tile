// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package newrelic

import (
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrclients"
)

// Harvester ...
type Harvester struct {
	collector *Collector
}

// NewHarvester ...
func NewHarvester(c *Collector) *Harvester {
	i := &Harvester{
		collector: c,
	}
	return i
}

// Accumulators return iterable list of registered Accumulators
func (h *Harvester) Accumulators() []accumulators.Interface {
	return *h.collector.accumulators
}

// Harvest queues processed metrics
func (h *Harvester) Harvest() {
	app.Get().Log.Debug("\nHarvest...")
	for _, accumulator := range h.Accumulators() {
		for _, entity := range accumulator.Drain() {
			for _, metric := range entity.DrainMetrics() {
				accumulator.HarvestMetrics(entity, metric)
			}
		}
	}
	app.Get().Log.Debug("Harvest COMPLETE")
	// Tell the ClientManager to flush all clients.
	nrclients.New().FlushAll()
}
