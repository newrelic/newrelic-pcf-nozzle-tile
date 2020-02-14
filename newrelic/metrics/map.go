// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"sync"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/uid"
)

// Collection ...
type Collection map[uid.ID]*Metric

// Map ...
type Map struct {
	collection Collection
	sync       *sync.RWMutex
}

// NewMap ...
func NewMap() *Map {
	return &Map{
		collection: Collection{},
		sync:       &sync.RWMutex{},
	}
}

// Drain ...
func (m *Map) Drain() (c []*Metric) {
	m.sync.Lock()
	defer m.sync.Unlock()
	for _, v := range m.collection {
		c = append(c, v)
	}
	m.collection = Collection{}
	return c
}

// ForEach ...
func (m *Map) ForEach(fn func(metric *Metric)) int {
	count := 0
	m.sync.Lock()
	for _, v := range m.collection {
		fn(v)
		count++
	}
	m.sync.Unlock()
	return count
}

// Has ...
func (m *Map) Has(id uid.ID) (metric *Metric, found bool) {
	m.sync.RLock()
	defer m.sync.RUnlock()
	if metric, found = m.collection[id]; found {
		return metric, true
	}
	return metric, false
}

// Put ...
func (m *Map) Put(metric *Metric) {
	metric.mapSync = m.sync
	m.sync.Lock()
	m.collection[metric.Signature()] = metric
	m.sync.Unlock()
}

// Count ...
func (m *Map) Count() int {
	return len(m.collection)
}
