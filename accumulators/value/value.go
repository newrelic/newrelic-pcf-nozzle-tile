// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package value

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"encoding/json"
	"fmt"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/entities"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/insights"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrpcf"
	"os"
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

	var update string
	var debug *metrics.Metric
	jsonP, _ := json.Marshal(e)
	f, err := os.OpenFile("../tmpdat1", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.Write([]byte("\n\nNEW METRIC\n\n")); err != nil {
		panic(err)
	}

	if _, err = f.Write(jsonP); err != nil {
		panic(err)
	}
	if _, err = f.Write([]byte("\n")); err != nil {
		panic(err)
	}
	ent := m.GetEntity(e, nrpcf.GetPCFAttributes(e))
	g := e.GetGauge()
	// A single v2 envelope can contain multiple metrics.
	for key, met := range g.Metrics {
		update, debug = ent.NewSample(key, metrics.Types.Gauge, met.GetUnit(), met.GetValue()).Done()
		jsonP, _ = json.Marshal(debug.Marshal())
	}

	if _, err = f.Write([]byte(update + " " + fmt.Sprintf("%d", m.Accumulator.Entities.Count()))); err != nil {
		panic(err)
	}
	if _, err = f.Write(jsonP); err != nil {
		panic(err)
	}

}

// HarvestMetrics ...
func (m Metrics) HarvestMetrics(entity *entities.Entity, metric *metrics.Metric) {

	f, err := os.OpenFile("../tmpdat1", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	jsonP, _ := json.Marshal(metric.Marshal())
	if _, err = f.Write([]byte("\n\nHARVEST\n\n")); err != nil {
		panic(err)
	}
	if _, err = f.Write(jsonP); err != nil {
		panic(err)
	}

	metric.SetAttribute(
		"eventType",
		m.Config().GetString(config.NewRelicEventTypeValueMetric),
	)

	metric.SetAttribute("agent.subscription", m.Config().GetString("FIREHOSE_ID"))

	metric.Attributes().AppendAll(entity.Attributes())

	// Get a client with the insert key and RPM account ID from the config.
	client := insights.New().Get(app.Get().Config.GetNewRelicConfig())
	client.EnqueueEvent(metric.Marshal())

	jsonP, _ = json.Marshal(metric.Marshal())
	if _, err = f.Write(jsonP); err != nil {
		panic(err)
	}

}
