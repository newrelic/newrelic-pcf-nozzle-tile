// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"github.com/newrelic/newrelic-pcf-nozzle-tile/accumulators/container"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/accumulators/counter"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/accumulators/http"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/accumulators/logmessage"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/accumulators/value"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
)

// Accumulators ...
type Accumulators []accumulators.Interface

// Register ...
var Register = &Accumulators{
	counter.Metrics{},
	container.Metrics{},
	value.Metrics{},
	logmessage.Nrevents{},
	http.Nrevents{},
}
