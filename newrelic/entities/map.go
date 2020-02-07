// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package entities

import (
	"os"
	"sync"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/uid"
)

// Collection ...
type Collection map[uid.ID]*Entity

// Map ...
type Map struct {
	collection Collection
	sync       *sync.RWMutex
}

// Drain returns collection of Entities ...
func (m *Map) Drain() (c []*Entity) {
	m.sync.Lock()
	defer m.sync.Unlock()
	for _, v := range m.collection {
		c = append(c, v)
	}
	m.collection = Collection{}
	return c
}

// ForEach ...
func (m *Map) ForEach(fn func(entity *Entity)) int {
	count := 0
	m.sync.Lock()
	for _, v := range m.collection {
		fn(v)
		count++
	}
	m.sync.Unlock()
	return count
}

// FindAllMetrics stub
func (m *Map) FindAllMetrics(name string) (ms *Map) {
	return ms
}

// Has ...
func (m *Map) Has(id uid.ID) (entity *Entity, found bool) {
	m.sync.RLock()
	defer m.sync.RUnlock()
	if entity, found = m.collection[id]; found {
		return entity, true
	}
	return entity, false
}

// Put ...
func (m *Map) Put(entity *Entity) {
	go func() {
		f, err := os.OpenFile("../tmpdat1", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if _, err = f.Write([]byte("\nstartWaiting on entity lock\n")); err != nil {
			panic(err)
		}

		m.sync.Lock()
		if _, err = f.Write([]byte("\nfiniscedWaiting on entity lock\n")); err != nil {
			panic(err)
		}
		m.collection[entity.Signature()] = entity
		m.sync.Unlock()
	}()
}

// Count ...
func (m *Map) Count() int {
	return len(m.collection)
}
