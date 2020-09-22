// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package cfapps

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
)

// Cache ...
type Cache struct {
	Collection  map[string]*CFApp
	WriteBuffer chan *CFApp
	sync        *sync.RWMutex
	isUpdating  bool
}

// NewCache ...
func NewCache() *Cache {
	cache := &Cache{
		Collection:  map[string]*CFApp{},
		WriteBuffer: make(chan *CFApp, app.Get().Config.GetDuration("FIREHOSE_CACHE_WRITE_BUFFER_SIZE")),
		sync:        &sync.RWMutex{},
	}
	cache.Start()
	return cache
}

// Start ...
func (c *Cache) Start() {
	go func() {
		// Staggering when applications may be removed from the cache by instance ID.
		instanceId := os.Getenv("CF_INSTANCE_INDEX")
		instanceIdInt, err := strconv.Atoi(instanceId)
		if err != nil {
			instanceIdInt = 0
		}
		cacheDuration := app.Get().Config.GetDuration("FIREHOSE_CACHE_DURATION_MINS")
		cacheUpdate := app.Get().Config.GetDuration("FIREHOSE_CACHE_UPDATE_INTERVAL_SECS")
		update := time.NewTicker(cacheUpdate * time.Second).C
		timeoutCache := time.NewTicker((cacheDuration + time.Duration(instanceIdInt*2)) * time.Minute).C

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
				if !c.isUpdating {
					go c.updateInstances()
				}

			case app := <-c.WriteBuffer:
				c.sync.Lock()
				c.Collection[app.GUID] = app
				c.sync.Unlock()
			}
		}
	}()
}

func (c *Cache) updateInstances() {
	c.isUpdating = true
	GetInstance().app.Log.Debug("Updating status of applications")
	now := time.Now()

	var apps []*CFApp
	c.sync.RLock()
	for _, v := range c.Collection {
		apps = append(apps, v)
	}
	c.sync.RUnlock()

	for _, a := range apps {
		a.UpdateInstances()
	}

	c.isUpdating = false
	GetInstance().app.Log.Debugf("Finish cache updating %s", time.Since(now))
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
