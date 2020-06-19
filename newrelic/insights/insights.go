// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package insights

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"

	"github.com/newrelic/go-insights/client"
)

var once sync.Once
var instance *InsertManager
var cfg = app.Get().Config

// New ...
func New() *InsertManager {
	once.Do(func() {
		instance = &InsertManager{
			collection: map[string]*client.InsertClient{},
			sync:       &sync.RWMutex{},
		}
	})
	return instance
}

// InsertManager ...
type InsertManager struct {
	collection map[string]*client.InsertClient
	sync       *sync.RWMutex
}

// Has ...
func (im *InsertManager) Has(insertKey string) (c *client.InsertClient, ok bool) {
	im.sync.RLock()
	defer im.sync.RUnlock()
	c, ok = im.collection[insertKey]
	return c, ok
}

// Put ...
func (im *InsertManager) Put(insertKey string, c *client.InsertClient) {
	im.sync.Lock()
	im.collection[insertKey] = c
	im.sync.Unlock()
}

func checkInsightsKey(c *client.InsertClient) error {
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

	if err := c.PostEvent(logEntry.Marshal()); err != nil {
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("invalid insights insert api key: %v", err)
		}
	}
	return nil
}

// New ...
func (im *InsertManager) New(insightsInsertKey string, rpmAccountID string, accountRegion string) *client.InsertClient {
	insertClient := client.NewInsertClient(insightsInsertKey, rpmAccountID)
	insertClient.Logger.Out = os.Stdout
	insertClient.SetCompression(client.Gzip) //always use compression to Insights
	if accountRegion == "EU" {
		//UseCustomURL only sets the host (domain) of the URL
		insertClient.UseCustomURL(cfg.GetString("NEWRELIC_EU_BASE_URL"))
	}
	if cfg.GetString("NEWRELIC_CUSTOM_URL") != "" {
		insertClient.UseCustomURL(cfg.GetString("NEWRELIC_CUSTOM_URL"))
	}
	// a regular check on the insight licence is implemented. If an error related with the key is
	// returned from insights the nozzle will be stopped.
	go func(rpm string) {
		for {
			if err := checkInsightsKey(insertClient); err != nil {
				app.Get().Log.Fatalf("fail insights insert client for rpm %s:  %s", rpm, err.Error())
			}
			app.Get().Log.Debugf("insights key successfully checked for rpm:%s", rpm)
			time.Sleep(10 * time.Minute)
		}
	}(rpmAccountID)

	insertClient.Start()
	im.sync.Lock()
	im.collection[insightsInsertKey] = insertClient
	im.sync.Unlock()
	return insertClient
}

// Get ...
func (im *InsertManager) Get(insightsInsertKey string, rpmAccountID string, accountRegion string) *client.InsertClient {
	if c, ok := im.Has(insightsInsertKey); ok {
		return c
	}
	return im.New(insightsInsertKey, rpmAccountID, accountRegion)
}

// FlushAll clients
func (im *InsertManager) FlushAll() {
	for _, c := range im.collection {
		c.Flush()
	}
}
