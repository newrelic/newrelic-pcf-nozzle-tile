// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"sync"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/logger"
)

var instance *Application
var once sync.Once

// Application Context
type Application struct {
	Config    *config.Config
	Running   chan bool
	CloseChan chan bool
	ErrorChan chan error
	Log       *logger.Logger
	WaitGroup *sync.WaitGroup
}

// Get Application Settings
func Get() *Application {

	once.Do(func() {
		instance = &Application{
			Config:    config.Get(),
			Running:   make(chan bool),
			ErrorChan: make(chan error),
			CloseChan: make(chan bool),
			WaitGroup: &sync.WaitGroup{},
		}
		instance.Log = logger.New(instance.Config)
	})

	return instance

}
