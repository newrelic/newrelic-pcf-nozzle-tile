// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nrevents

import (
	"sync"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/uid"
)

// Collection ...
type Collection map[uid.ID]*Nrevent

// Nreventmap ...
type Nreventmap struct {
	collection Collection
	sync       *sync.RWMutex
}

// NewMap ...
func NewMap() *Nreventmap {
	return &Nreventmap{
		collection: Collection{},
		sync:       &sync.RWMutex{},
	}
}

// Drain ...
func (n *Nreventmap) Drain() (c []*Nrevent) {
	n.sync.Lock()
	defer n.sync.Unlock()
	for _, v := range n.collection {
		c = append(c, v)
	}
	n.collection = Collection{}
	return c
}

// ForEach ...
func (n *Nreventmap) ForEach(fn func(event *Nrevent)) int {
	count := 0
	n.sync.Lock()
	for _, v := range n.collection {
		fn(v)
		count++
	}
	n.sync.Unlock()
	return count
}

// Has ...
func (n *Nreventmap) Has(id uid.ID) (event *Nrevent, found bool) {
	n.sync.RLock()
	defer n.sync.RUnlock()
	if event, found = n.collection[id]; found {
		return event, true
	}
	return event, false
}

// Put ...
func (n *Nreventmap) Put(event *Nrevent) {
	event.mapSync = n.sync
	n.sync.Lock()
	n.collection[event.Signature()] = event
	n.sync.Unlock()
}

// Count ...
func (n *Nreventmap) Count() int {
	return len(n.collection)
}
