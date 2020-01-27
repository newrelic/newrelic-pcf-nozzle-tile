// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic"
)

// Version uses -ldflags "-X main.Version=$(git describe)"
var Version string

func main() {

	version()
	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	newrelic.Start(interupt)

}

func inArgs(a string) bool {
	args := strings.Join(os.Args, ",")
	return strings.Contains(args, a)
}

func version() {
	if len(Version) > 0 {
		config.Get().Set("Version", Version)
	}

	if inArgs("version") {
		fmt.Printf("\nNew Relic PCF Firehose Agent: %s\n\n", config.Get().GetString("Version"))
		os.Exit(0)
	}
}
