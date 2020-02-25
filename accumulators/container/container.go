// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/cfclient/cfapps"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/entities"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrpcf"
)

// Metrics extends metric.Accumulator for
// Firehose ContainerMetric Envelope Event Types
type Metrics struct {
	accumulators.Accumulator
	CFAppManager *cfapps.CFAppManager
}

// New satisfies metric.Accumulator
func (m Metrics) New() accumulators.Interface {
	i := Metrics{
		Accumulator: accumulators.NewAccumulator(
			// This isn't a v2 envelope type, but the router will route matching Gauge envelopes here.
			"ContainerMetric",
		),
		CFAppManager: cfapps.GetInstance(),
	}
	return i
}

// Update ...
func (m Metrics) Update(e *loggregator_v2.Envelope) {
	entity := m.GetEntity(e, nrpcf.GetPCFAttributes(e))

	attrs := m.CFAppManager.GetAppInstanceAttributes(
		e.GetSourceId(),
		m.ConvertSourceInstance(e.GetInstanceId()),
	)

	entity.Attributes().AppendAll(attrs)

	entity.NewSample(
		"app.cpu",
		metrics.Types.Gauge,
		"percent",
		e.GetGauge().Metrics["cpu"].GetValue(),
	).Done()

	entity.NewSample(
		"app.disk",
		metrics.Types.Gauge,
		"bytes",
		e.GetGauge().Metrics["disk"].GetValue(),
	).SetAttribute(
		"app.disk.quota",
		e.GetGauge().Metrics["disk_quota"].GetValue(),
	).Done()

	entity.NewSample(
		"app.memory",
		metrics.Types.Gauge,
		"bytes",
		e.GetGauge().Metrics["memory"].GetValue(),
	).SetAttribute(
		"app.memory.quota",
		e.GetGauge().Metrics["memory_quota"].GetValue(),
	).Done()

}

// HarvestMetrics ...
func (m Metrics) HarvestMetrics(

	entity *entities.Entity,
	metric *metrics.Metric,

) {

	if metric.Name != "app.cpu" {
		percentUsedAttributeName := fmt.Sprintf("%s.used", metric.Name)
		metric.
			SetAttribute(
				percentUsedAttributeName,
				calculateUsed(metric),
			)
	}

	metric.SetAttribute(
		"eventType",
		config.Get().GetString(config.NewRelicEventTypeContainer),
	)

	metric.SetAttribute("agent.subscription", config.Get().GetString("FIREHOSE_ID"))

	metric.Attributes().
		AppendAll(entity.Attributes())

	// Get a client for this metric - checking for insert key and account ID info in the application
	// We will default to what is in the configuration file	if application specific info isn't found
	client := nrpcf.GetInsertClientForApp(entity)
	client.EnqueueEvent(metric.Marshal())

}

func calculateUsed(metric *metrics.Metric) float64 {
	quotaAttributeName := fmt.Sprintf("%s.quota", metric.Name)
	bytesUsed := metric.LastValue
	bytesQuota := metric.Attributes().FloatValueOf(quotaAttributeName)
	return (bytesUsed / bytesQuota) * 100
}

// ConvertSourceInstance from a string to int32
func (m Metrics) ConvertSourceInstance(i string) int32 {
	if num, err := strconv.ParseInt(i, 10, 32); err == nil {
		return int32(num)
	}
	return 0
}
