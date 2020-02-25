// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package value

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/entities"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/insights"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrpcf"
)

// Metrics extends metric.Accumulator for
// Firehose ContainerMetric Envelope Event Types
type Metrics struct {
	accumulators.Accumulator
}

// New satisfies metric.Accumulator
func (m Metrics) New() accumulators.Interface {
	i := Metrics{
		Accumulator: accumulators.NewAccumulator(
			// Does not match a v2 envelope type, but router will send appropriate envelopes here.
			"ValueMetric",
		),
	}
	return i
}

// Update satisfies metric.Accumulator
func (m Metrics) Update(e *loggregator_v2.Envelope) {

	ent := m.GetEntity(e, nrpcf.GetPCFAttributes(e))
	g := e.GetGauge()
	// A single v2 envelope can contain multiple metrics.
	for key, met := range g.Metrics {
		ent.
			NewSample(
				key,
				metrics.Types.Gauge,
				met.GetUnit(),
				met.GetValue(),
			).
			Done()
	}
}

// HarvestMetrics ...
func (m Metrics) HarvestMetrics(entity *entities.Entity, metric *metrics.Metric) {

	metric.SetAttribute(
		"eventType",
		config.Get().GetString(config.NewRelicEventTypeValueMetric),
	)

	metric.SetAttribute("agent.subscription", config.Get().GetString("FIREHOSE_ID"))

	metric.Attributes().AppendAll(entity.Attributes())

	// Get a client with the insert key and RPM account ID from the config.
	client := insights.New().Get(config.Get().GetNewRelicConfig())
	client.EnqueueEvent(metric.Marshal())

}
