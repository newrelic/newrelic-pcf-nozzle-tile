// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package counter

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
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
			"*loggregator_v2.Envelope_Counter",
		),
	}
	return i
}

// Update satisfies metric.Accumulator
func (m Metrics) Update(e *loggregator_v2.Envelope) {
	m.GetEntity(e, nrpcf.GetPCFAttributes(e)).
		NewSample(
			e.GetCounter().Name,
			metrics.Types.Delta, "delta",
			float64(e.GetCounter().GetDelta()),
		).
		SetAttribute("total.reported",
			e.GetCounter().GetTotal(),
		).
		Done()

}

// HarvestMetrics ...
func (m Metrics) HarvestMetrics(

	entity *entities.Entity,
	metric *metrics.Metric,

) {

	metric.SetAttribute("eventType",
		m.Config().GetString(config.NewRelicEventTypeCounterEvent),
	)

	metric.SetAttribute("agent.subscription", m.Config().GetString("FIREHOSE_ID"))

	metric.Attributes().
		AppendAll(entity.Attributes())

	// Get a client with the insert key and RPM account ID from the config.
	client := insights.New().Get(app.Get().Config.GetNewRelicConfig())
	client.EnqueueEvent(metric.Marshal())
}
