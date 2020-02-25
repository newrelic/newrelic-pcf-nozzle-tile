// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package insights

import (
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"os"
	"sync"

	"github.com/newrelic/go-insights/client"
)

var once sync.Once
var instance *InsertManager

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

// New ...
func (im *InsertManager) New(insightsInsertKey string, rpmAccountID string, accountRegion string) *client.InsertClient {
	insertClient := client.NewInsertClient(insightsInsertKey, rpmAccountID)
	insertClient.Logger.Out = os.Stdout
	insertClient.SetCompression(client.Gzip) //always use compression to Insights
	if accountRegion == "EU" {
		//UseCustomURL only sets the host (domain) of the URL
		insertClient.UseCustomURL(config.Get().GetString("NEWRELIC_EU_BASE_URL"))
	}
	if config.Get().GetString("NEWRELIC_CUSTOM_URL") != "" {
		insertClient.UseCustomURL(config.Get().GetString("NEWRELIC_CUSTOM_URL"))
	}
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
