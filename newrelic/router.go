// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package newrelic

import (
	"reflect"
	"strings"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/firehose"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
)

// Streams ...
type Streams map[string][]accumulators.Interface

// Router object
type Router struct {
	App       *app.Application
	Consumer  *firehose.OneToOneEnvelope
	Streams   Streams
	Collector *Collector
	ErrorChan chan error
	closeChan chan bool
	firehose  *firehose.Firehose
}

// NewRouter with Firehose
func NewRouter(f *firehose.Firehose, c *Collector) *Router {
	router := &Router{
		App:       app.Get(),
		Consumer:  f.Queue,
		Streams:   Streams{},
		Collector: c,
		ErrorChan: make(chan error, 1),
		closeChan: make(chan bool, 1),
		firehose:  f,
	}
	// These are the possible event type names:
	// 		*loggregator_v2.Envelope_Counter
	// 		ContainerMetric
	// 		ValueMetric
	// 		*loggregator_v2.Envelope_Log
	// 		*loggregator_v2.Envelope_Timer
	for _, a := range *router.Collector.accumulators {
		for _, s := range a.Streams() {
			for _, t := range router.App.Config.GetNewEnvelopeTypes() {
				if strings.HasSuffix(strings.ToLower(s), t) {
					router.Streams[s] = append(router.Streams[s], a)
				}
			}
		}
	}
	return router
}

// Close Router
func (r *Router) Close() {
	r.closeChan <- true
}

// Start Router
func (r *Router) Start() {

	// consume from firehose and route synchronously
	go func() {
		r.App.Log.Info("router started")
		// Track the number of consecutive empty diodes
		ed := 0
		for {

			select {

			case <-r.closeChan:
				r.App.Log.Info("closed router")
				return

			case err := <-r.ErrorChan:
				r.App.Log.Errorf("Router error: %e", err.Error())

			default:
				if e, notEmpty := r.Consumer.TryNext(); notEmpty {
					// Reset the emptyDiodes count.  We found an envelope.
					ed = 0

					et := reflect.TypeOf(e.Message).String()
					if et == "*loggregator_v2.Envelope_Gauge" {
						if isContainerMetric(e) {
							et = "ContainerMetric"
						} else {
							et = "ValueMetric"
						}
					}

					for _, a := range r.Streams[et] {
						r.App.Log.Tracer(">")
						a.Update(e)
					}
					continue
				}
				r.App.Log.Tracer("o")
				ed++
				// Have the diodes been empty for > threshold?  Multiplying by 2 due to 500ms sleep
				if ed > (r.App.Config.GetInt("FIREHOSE_RESTART_THRESH_SECS") * 2) {
					// Reset the emptyDiodes counter so we don't constanty restart the nozzle every 500ms
					ed = 0
					r.App.Log.Warnf("Diodes have been empty for > %v seconds.  Restarting firehose nozzle.", r.App.Config.GetInt("FIREHOSE_RESTART_THRESH_SECS"))
					r.firehose.RestartNozzle()
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

}

// isContainerMetric determines if the current v2 Gauge envelope is a v1 ContainerMetric or v1 ValueMetric
func isContainerMetric(e *loggregator_v2.Envelope) bool {
	gauge := e.GetGauge()
	if len(gauge.Metrics) != 5 {
		return false
	}
	required := []string{
		"cpu",
		"memory",
		"disk",
		"memory_quota",
		"disk_quota",
	}

	for _, req := range required {
		if _, found := gauge.Metrics[req]; !found {
			return false
		}
	}
	return true
}
