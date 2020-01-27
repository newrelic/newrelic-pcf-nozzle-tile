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
	Collection  map[string]*CFApp
	WriteBuffer chan *CFApp
	sync        *sync.RWMutex
}

// NewCache ...
func NewCache() *Cache {
	cache := &Cache{
		Collection:  map[string]*CFApp{},
		WriteBuffer: make(chan *CFApp, 1024),
		sync:        &sync.RWMutex{},
	}
	cache.Start()
	return cache
}

// Start ...
func (c *Cache) Start() {
	go func() {
		cacheDuration := app.Get().Config.GetDuration("FIREHOSE_CACHE_DURATION_MINS")
		for {
			select {

			case <-time.NewTicker(cacheDuration * time.Minute).C:
				Collection := map[string]*CFApp{}
				c.sync.Lock()
				for k, v := range c.Collection {
					if time.Since(v.LastPull) < cacheDuration*time.Minute {
						Collection[k] = v
					}
					c.Collection = Collection
				}
				c.sync.Unlock()

			case <-time.NewTicker(30 * time.Second).C:
				for _, v := range c.Collection {
					v.UpdateInstances()
				}

			case app := <-c.WriteBuffer:
				c.sync.Lock()
				c.Collection[app.GUID] = app
				c.sync.Unlock()

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
func (c *Cache) Put(app *CFApp) {
	c.WriteBuffer <- app
}

// Drain ...
func (c *Cache) Drain() map[string]*CFApp {
	c.sync.Lock()
	defer c.sync.Unlock()
	result := map[string]*CFApp{}
	for k, v := range c.Collection {
		result[k] = v
	}
	c.Collection = map[string]*CFApp{}
	return result
}
