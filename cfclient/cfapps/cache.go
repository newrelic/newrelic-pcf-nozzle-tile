// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cfapps

import (
	"sync"
	"time"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
)

// Cache ...
type Cache struct {
	Collection map[string]*CFApp
	sync       *sync.RWMutex
}

// NewCache ...
func NewCache() *Cache {
	cache := &Cache{
		Collection: map[string]*CFApp{},
		sync:       &sync.RWMutex{},
	}
	cache.Start()
	return cache
}

// Start ...
func (c *Cache) Start() {
	go func() {
		cacheDuration := app.Get().Config.GetDuration("FIREHOSE_CACHE_DURATION_MINS")
		update := time.NewTicker(30 * time.Second).C
		timeoutCache := time.NewTicker(cacheDuration * time.Minute).C

		for {
			select {

			case <-timeoutCache:
				GetInstance().app.Log.Debug("Cleaning Cache")
				GetInstance().app.Log.Debug("Cache length before cleaning: ", len(c.Collection))
				Collection := map[string]*CFApp{}
				c.sync.Lock()
				for k, v := range c.Collection {
					if time.Since(v.LastPull) < cacheDuration*time.Minute {
						Collection[k] = v
					}
					c.Collection = Collection
				}
				GetInstance().app.Log.Debug("Cache length after cleaning: ", len(c.Collection))
				c.sync.Unlock()

			case <-update:
				GetInstance().app.Log.Debug("Updating status of applications")
				c.sync.RLock()
				for _, v := range c.Collection {
					v.UpdateInstances()
				}
				c.sync.RUnlock()
			}
		}
	}()
}

// Get ...
func (c *Cache) Get(id string) (app *CFApp, found bool) {
	c.sync.RLock()
	defer c.sync.RUnlock()

	if app, found = c.Collection[id]; found {
		return app, true
	}
	return app, false
}

// Put ...
func (c *Cache) Put(id string) *CFApp {
	c.sync.Lock()
	defer c.sync.Unlock()

	if app, found := c.Collection[id]; found {
		return app
	}
	GetInstance().app.Log.Debug("Adding new app: ", id)
	app := NewCFApp(id)
	c.Collection[app.GUID] = app
	GetInstance().UpdateAppAsync(app)

	return app
}
