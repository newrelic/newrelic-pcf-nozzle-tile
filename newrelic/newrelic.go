// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package newrelic

import (
	"net/http"
	"os"
	"runtime"
	"time"

	_ "net/http/pprof"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/cfclient/cfapps"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/firehose"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/healthcheck"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/registry"
)

// NewRelic Object
type NewRelic struct {
	App          *app.Application
	CFAppManager *cfapps.CFAppManager
	Firehose     *firehose.Firehose
	Router       *Router
	Streams      Streams
	Harvester    *Harvester
	Harvest      *time.Ticker
	Collector    *Collector
	ErrorChan    chan error
	Running      chan bool
}

// Start New Relic Processing
func Start(interupt <-chan os.Signal) {

	app := app.Get()

	if app.Config.GetBool("PPROF") {
		go func() {
			runtime.SetBlockProfileRate(1)
			app.Log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	nr := &NewRelic{
		App:          app,
		CFAppManager: cfapps.Start(app),
		Harvest:      harvestConfig(app),
		Collector:    NewCollector(registry.Register),
	}

	nr.Firehose = firehose.Start()
	nr.Router = NewRouter(nr.Firehose, nr.Collector)
	nr.Router.Start()
	nr.Harvester = NewHarvester(nr.Collector)
	healthcheck.Start()

	for {
		select {

		case <-interupt:
			app.Log.Info("interupt received, gracefully closing...")
			nr.Firehose.Close()
			nr.Router.Close()
			app.WaitGroup.Wait()
			app.Log.Info("closed New Relic")
			return

		case err := <-app.ErrorChan:
			app.Log.Error(err)

		case <-nr.Harvest.C:
			nr.Firehose.ResetEventCount()
			nr.Harvester.Harvest()
		}

	}
}

func harvestConfig(app *app.Application) *time.Ticker {
	return time.NewTicker(
		app.Config.GetDuration("NEWRELIC_DRAIN_INTERVAL"),
	)
}
