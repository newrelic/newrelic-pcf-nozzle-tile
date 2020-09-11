// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"strconv"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/entities"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrclients"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrpcf"
)

// Nrevents extends event.Accumulator for
// Firehose HttpStartStop Envelope Event Types
type Nrevents struct {
	accumulators.Accumulator
	logsEnabled bool
}

// New satisfies event.Accumulator
func (n Nrevents) New() accumulators.Interface {
	i := Nrevents{
		Accumulator: accumulators.NewAccumulator(
			"*loggregator_v2.Envelope_Timer",
		),
	}
	i.logsEnabled = i.Config().GetBool("LOGS_HTTP")
	return i
}

// Update satisfies event.Accumulator
// func (n Nrevents) Update(e *events.Envelope) {
func (n Nrevents) Update(e *loggregator_v2.Envelope) {
	entity := n.GetEntity(e, nrpcf.GetPCFAttributes(e))
	s := attributes.NewAttributes()
	s.SetAttribute("http.duration", float64(n.GetDuration(e)))
	cl, clErr := strconv.ParseInt(n.GetTag(e, "content_length"), 10, 0)
	if clErr == nil {
		s.SetAttribute("http.content.length", cl)
	}
	sc, scErr := strconv.ParseInt(n.GetTag(e, "status_code"), 10, 0)
	if scErr == nil {
		s.SetAttribute("http.status", sc)
	}
	s.SetAttribute("http.uri", n.GetTag(e, "uri"))
	s.SetAttribute("http.method", n.GetTag(e, "method"))
	s.SetAttribute("http.peer.type", n.GetTag(e, "peer_type"))
	s.SetAttribute("http.start.timestamp", e.GetTimer().GetStart())
	s.SetAttribute("http.stop.timestamp", e.GetTimer().GetStop())
	s.SetAttribute("http.remote.address", n.GetTag(e, "remote_address"))
	s.SetAttribute("http.user.agent", n.GetTag(e, "user_agent"))
	s.SetAttribute("http.request.id", n.GetTag(e, "request_id"))

	s.SetAttribute("agent.subscription", n.Config().GetString("FIREHOSE_ID"))

	s.AppendAll(entity.Attributes())

	if n.logsEnabled {
		client := nrclients.New().GetLogClient(app.Get().Config.GetNewRelicConfig())
		client.EnqueueLogEntry(context.Background(), s.Marshal())
		return
	}
	s.SetAttribute(
		"eventType",
		n.Config().GetString(config.NewRelicEventTypeHTTPStartStop),
	)
	// Get an insert client and enqueue the event.
	client := nrclients.New().GetEventClient(app.Get().Config.GetNewRelicConfig())
	client.EnqueueEvent(context.Background(), s.Marshal())
}

// HarvestMetrics (stub for HttpStartStop)...
func (n Nrevents) HarvestMetrics(
	entity *entities.Entity,
	metric *metrics.Metric,
) {
}

// GetDuration ...
func (n Nrevents) GetDuration(
	e *loggregator_v2.Envelope,
) float64 {
	return float64(time.Unix(0, e.GetTimer().GetStop()).Sub(time.Unix(0, e.GetTimer().GetStart()))) / float64(time.Millisecond)
}

// GetTag ...
func (n Nrevents) GetTag(
	e *loggregator_v2.Envelope,
	ta string,
) string {
	if tv, ok := e.Tags[ta]; ok {
		return tv
	}
	return ""
}
