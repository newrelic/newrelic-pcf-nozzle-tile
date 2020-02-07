// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"os"
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

	f, err := os.OpenFile("../tmpdat1", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.Write([]byte("\n\nChecking if id present in the metrics already collected: " + id.String())); err != nil {
		panic(err)
	}

	for a, b := range m.collection {
		if _, err = f.Write([]byte(a.String() + "------" + b.Name)); err != nil {
			panic(err)
		}
	}

	if metric, found = m.collection[id]; found {
		return metric, true
	}
	return metric, false
}

// Put ...
func (m *Map) Put(metric *Metric) {
	metric.mapSync = m.sync
	go func() {
		f, err := os.OpenFile("../tmpdat1", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if _, err = f.Write([]byte("\nstartWaiting on metric lock\n")); err != nil {
			panic(err)
		}

		m.sync.Lock()
		if _, err = f.Write([]byte("\nfiniscedWaiting on metric lock\n")); err != nil {
			panic(err)
		}
		m.collection[metric.Signature()] = metric
		m.sync.Unlock()
	}()
}

// Count ...
func (m *Map) Count() int {
	return len(m.collection)
}
