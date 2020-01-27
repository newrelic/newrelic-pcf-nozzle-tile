// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package firehose

import (
	"context"
	"log"
	"os"
	"sync/atomic"

	"github.com/cloudfoundry/go-loggregator"

	"code.cloudfoundry.org/go-diodes"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/cfclient/api"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/firehose/httpfirehose"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/logger"
)

// Firehose Object...
type Firehose struct {
	log        *logger.Logger
	config     *config.Config
	nozzle     *loggregator.RLPGatewayClient
	eventsChan chan *loggregator_v2.Envelope
	closeChan  chan bool
	Queue      *OneToOneEnvelope
	EventCount int64
	cancel     context.CancelFunc
}

// Close Firehose
func (f *Firehose) Close() {
	f.cancel()
	f.closeChan <- true
	f.log.Info("closed firehose consumer")
}

// GetEventCount ...
func (f *Firehose) GetEventCount() int64 {
	return atomic.LoadInt64(&f.EventCount)
}

// ResetEventCount ...
func (f *Firehose) ResetEventCount() {
	atomic.StoreInt64(&f.EventCount, 0)
}

// Start New Firehose
func Start() *Firehose {

	var errorChan <-chan error

	f := &Firehose{
		log:        app.Get().Log,
		config:     app.Get().Config,
		eventsChan: make(chan *loggregator_v2.Envelope, 1024),
		closeChan:  make(chan bool),
	}

	f.log.Info("starting firehose")

	pcf, err := api.New()

	if err != nil {
		f.log.Fatalf("failed to start PCF Firehose: %s", err.Error())
	}

	f.ResetEventCount()

	// Create a HTTP client which will be used to interact with the RLP Gateway
	fh := httpfirehose.NewHttpFirehose(pcf, f.config)

	// We will only pass a logger to the RLPGatewayClient if Debug level logging is enabled.
	if f.config.GetString("LOG_LEVEL") == "DEBUG" {
		log := log.New(os.Stdout, "RLP: ", log.Ldate|log.Ltime|log.Lshortfile)
		// Create a RLP Gateway Client using the HTTP client created above (with logger).
		f.nozzle = loggregator.NewRLPGatewayClient(f.config.GetString("CF_API_RLPG_URL"), loggregator.WithRLPGatewayClientLogger(log), loggregator.WithRLPGatewayHTTPClient(fh))
	} else {
		// Create a RLP Gateway Client using the HTTP client created above (no logger).
		f.nozzle = loggregator.NewRLPGatewayClient(f.config.GetString("CF_API_RLPG_URL"), loggregator.WithRLPGatewayHTTPClient(fh))
	}

	f.startNozzle()

	f.Queue = NewOneToOneEnvelope(
		f.config.GetInt("FIREHOSE_DIODE_BUFFER"),
		diodes.AlertFunc(func(missed int) {
			f.log.Warnf("Firehose diode dropped %d messages", missed)
		}))

	f.log.Info("firehose started")

	// Firehouse non-blocking event queuing via PCF diodes
	go func() {

		for {
			select {

			case err := <-errorChan:
				app.Get().ErrorChan <- err

			case <-f.closeChan:
				f.log.Info("closed firehose")
				return

			case event := <-f.eventsChan:
				f.Queue.Set(event)
				atomic.AddInt64(&f.EventCount, 1)
				f.log.Tracer("<")

			}
		}

	}()

	return f

}

// StartNozzle creates a context, connects to the RLP Gateway, and places envelopes on the eventsChan.
func (f *Firehose) startNozzle() {
	ctx, cancel := context.WithCancel(context.Background())
	f.cancel = cancel

	es := f.nozzle.Stream(ctx, &loggregator_v2.EgressBatchRequest{
		ShardId:   f.config.GetString("FIREHOSE_ID"),
		Selectors: f.config.GetSelectors(),
	})

	go func() {
		for ctx.Err() == nil {
			for _, e := range es() {
				f.eventsChan <- e
			}
		}
	}()
}

// RestartNozzle calls the context cancel function, then starts the nozzle again.
func (f *Firehose) RestartNozzle() {
	// Cancel the context, which will stop the current HTTP requests.
	f.cancel()
	// Restart the nozzle RLP Gateway connection
	f.startNozzle()
}
