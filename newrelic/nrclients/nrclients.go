// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nrclients

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"

	clientConfig "github.com/newrelic/newrelic-client-go/pkg/config"
	"github.com/newrelic/newrelic-client-go/pkg/events"
	"github.com/newrelic/newrelic-client-go/pkg/logs"
	"github.com/newrelic/newrelic-client-go/pkg/region"
)

var once sync.Once
var instance *ClientManager
var cfg = app.Get().Config

// New ...
func New() *ClientManager {
	once.Do(func() {
		instance = &ClientManager{
			eCollection: map[string]*events.Events{},
			lCollection: map[string]*logs.Logs{},
			sync:        &sync.RWMutex{},
		}
	})
	return instance
}

// ClientManager ...
type ClientManager struct {
	eCollection map[string]*events.Events
	lCollection map[string]*logs.Logs
	sync        *sync.RWMutex
}

// HasEventClient ...
func (cm *ClientManager) HasEventClient(insertKey string) (c *events.Events, ok bool) {
	cm.sync.RLock()
	defer cm.sync.RUnlock()
	c, ok = cm.eCollection[insertKey]
	return c, ok
}

// HasLogClient ...
func (cm *ClientManager) HasLogClient(insertKey string) (c *logs.Logs, ok bool) {
	cm.sync.RLock()
	defer cm.sync.RUnlock()
	c, ok = cm.lCollection[insertKey]
	return c, ok
}

// PutEventClient ...
func (cm *ClientManager) PutEventClient(insertKey string, c *events.Events) {
	cm.sync.Lock()
	cm.eCollection[insertKey] = c
	cm.sync.Unlock()
}

// PutLogClient ...
func (cm *ClientManager) PutLogClient(insertKey string, c *logs.Logs) {
	cm.sync.Lock()
	cm.lCollection[insertKey] = c
	cm.sync.Unlock()
}

func checkInsightsKeyEvents(c *events.Events, rpm int) error {
	logEntry := attributes.NewAttributes()
	u, _ := url.Parse(cfg.GetString(config.EnvCFAPIRUL))
	logEntry.SetAttribute("pcf.domain", u.Hostname())
	logEntry.SetAttribute("agent.version", cfg.GetString("Version"))
	logEntry.SetAttribute("agent.instance", cfg.GetInt("CF_INSTANCE_INDEX"))
	logEntry.SetAttribute("agent.ip", cfg.GetString("CF_INSTANCE_IP"))
	logEntry.SetAttribute("eventType", cfg.GetString(config.NewRelicEventTypeLogMessage))
	logEntry.SetAttribute("log.timestamp", time.Now().Unix())
	logEntry.SetAttribute("agent.subscription", cfg.GetString("FIREHOSE_ID"))
	logEntry.SetAttribute("log.message", "insights heartbeat")

	if err := c.CreateEvent(rpm, logEntry.Marshal()); err != nil {
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("invalid insights insert api key: %v", err)
		}
	}
	return nil
}

// Would it be possible to combine this into a single function?
// Maybe if we just iterate over the collections on a scheduled basis, sending a test message for each.
func checkInsightsKeyLogs(c *logs.Logs, rpm int) error {
	logEntry := attributes.NewAttributes()
	u, _ := url.Parse(cfg.GetString(config.EnvCFAPIRUL))
	logEntry.SetAttribute("pcf.domain", u.Hostname())
	logEntry.SetAttribute("agent.version", cfg.GetString("Version"))
	logEntry.SetAttribute("agent.instance", cfg.GetInt("CF_INSTANCE_INDEX"))
	logEntry.SetAttribute("agent.ip", cfg.GetString("CF_INSTANCE_IP"))
	logEntry.SetAttribute("timestamp", time.Now().Unix())
	logEntry.SetAttribute("agent.subscription", cfg.GetString("FIREHOSE_ID"))
	logEntry.SetAttribute("message", "insights heartbeat")

	if err := c.CreateLogEntry(logEntry.Marshal()); err != nil {
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("invalid insights insert api key: %v", err)
		}
	}
	return nil
}

// NewEventClient ...
func (cm *ClientManager) NewEventClient(insightsInsertKey string, rpmAccountID string, accountRegion string) *events.Events {
	clCfg := clientConfig.New()
	clCfg.InsightsInsertKey = insightsInsertKey
	clCfg.Compression = clientConfig.Compression.Gzip
	regName, err := region.Parse(accountRegion)
	if err != nil {
		app.Get().Log.Fatalf("fail parsing region while creating insert client")
	}
	reg, err := region.Get(regName)
	if err != nil {
		app.Get().Log.Fatalf("fail getting region while creating insert client")
	}
	if cfg.GetString("NEWRELIC_CUSTOM_URL") != "" {
		reg.SetInsightsBaseURL(cfg.GetString("NEWRELIC_CUSTOM_URL"))
	}
	err = clCfg.SetRegion(reg)
	if err != nil {
		app.Get().Log.Fatalf("fail setting region while creating insert client")
	}

	insertClient := events.New(clCfg)

	rpmID, err := strconv.Atoi(rpmAccountID)
	if err != nil {
		app.Get().Log.Fatalf("Error converting account ID to int: %v", rpmAccountID)
	}

	if err := insertClient.BatchMode(context.Background(), rpmID); err != nil {
		app.Get().Log.Fatalf("error starting batch mode (events): %s", err.Error())
	}

	// a regular check on the insight license is implemented. If an error related with the key is
	// returned from insights the nozzle will be stopped.
	go func(rpm int) {
		for {
			if err := checkInsightsKeyEvents(&insertClient, rpm); err != nil {
				app.Get().Log.Fatalf("fail insert client (events) for rpm %d: %s", rpm, err.Error())
			}
			app.Get().Log.Debugf("insert key (events) successfully checked for rpm: %d", rpm)
			time.Sleep(10 * time.Minute)
		}
	}(rpmID)

	cm.sync.Lock()
	cm.eCollection[insightsInsertKey] = &insertClient
	cm.sync.Unlock()
	return &insertClient
}

// NewLogClient ...
func (cm *ClientManager) NewLogClient(insightsInsertKey string, rpmAccountID string, accountRegion string) *logs.Logs {
	clCfg := clientConfig.New()
	clCfg.InsightsInsertKey = insightsInsertKey
	clCfg.Compression = clientConfig.Compression.Gzip
	regName, err := region.Parse(accountRegion)
	if err != nil {
		app.Get().Log.Fatalf("fail parsing region while creating insert client")
	}
	reg, err := region.Get(regName)
	if err != nil {
		app.Get().Log.Fatalf("fail getting region while creating insert client")
	}
	if cfg.GetString("NEWRELIC_CUSTOM_URL") != "" {
		reg.SetInsightsBaseURL(cfg.GetString("NEWRELIC_CUSTOM_URL"))
	}
	err = clCfg.SetRegion(reg)
	if err != nil {
		app.Get().Log.Fatalf("fail setting region while creating insert client")
	}

	insertClient := logs.New(clCfg)

	rpmID, err := strconv.Atoi(rpmAccountID)
	if err != nil {
		app.Get().Log.Fatalf("Error converting account ID to int: %v", rpmAccountID)
	}

	if err := insertClient.BatchMode(context.Background(), rpmID); err != nil {
		app.Get().Log.Fatalf("error starting batch mode (logs): %s", err.Error())
	}

	// a regular check on the insight license is implemented. If an error related with the key is
	// returned from insights the nozzle will be stopped.
	go func(rpm int) {
		for {
			if err := checkInsightsKeyLogs(&insertClient, rpm); err != nil {
				app.Get().Log.Fatalf("fail insert client (logs) for rpm %d: %s", rpm, err.Error())
			}
			app.Get().Log.Debugf("insert key (logs) successfully checked for rpm: %d", rpm)
			time.Sleep(10 * time.Minute)
		}
	}(rpmID)

	cm.sync.Lock()
	cm.lCollection[insightsInsertKey] = &insertClient
	cm.sync.Unlock()
	return &insertClient
}

// GetEventClient ...
func (cm *ClientManager) GetEventClient(insightsInsertKey string, rpmAccountID string, accountRegion string) *events.Events {
	if c, ok := cm.HasEventClient(insightsInsertKey); ok {
		return c
	}
	return cm.NewEventClient(insightsInsertKey, rpmAccountID, accountRegion)
}

// GetLogClient ...
func (cm *ClientManager) GetLogClient(insightsInsertKey string, rpmAccountID string, accountRegion string) *logs.Logs {
	if c, ok := cm.HasLogClient(insightsInsertKey); ok {
		return c
	}
	return cm.NewLogClient(insightsInsertKey, rpmAccountID, accountRegion)
}

// FlushAll clients
func (cm *ClientManager) FlushAll() {
	for _, c := range cm.eCollection {
		if err := c.Flush(); err != nil {
			app.Get().Log.Errorf("Unable to flush events: %v", err)
		}
	}

	for _, c := range cm.lCollection {
		if err := c.Flush(); err != nil {
			app.Get().Log.Errorf("Unable to flush logs: %v", err)
		}
	}
}
