// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nrevents

import (
	"sync"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/uid"
)

// Nrevent description
type Nrevent struct {
	attributes *attributes.Attributes
	mapSync    *sync.RWMutex
	sender     func(*Nrevent)
}

// New Nrevent
func New(
	attrs *attributes.Attributes,
) (e *Nrevent) {
	e = &Nrevent{
		attributes: attrs,
	}
	return e
}

// Signature of Entity
func (e *Nrevent) Signature() uid.ID {
	id := e.attributes.Signature()
	return id
}

// SetSender ...
func (e *Nrevent) SetSender(fn func(nrevents *Nrevent)) {
	e.sender = fn
}

// Harvest ...
func (e *Nrevent) Harvest() (r *map[string]interface{}) {

	e.Lock()
	r = e.Marshal()
	e.Unlock()

	return
}

// Marshal ...
func (e *Nrevent) Marshal() (r *map[string]interface{}) {
	marsh := e.attributes.Marshal()
	return &marsh
}

// Lock determines if event is part of Map and acts accordingly
func (e *Nrevent) Lock() {
	if e.mapSync == nil {
		return
	}
	e.mapSync.Lock()
}

// Unlock determines if metric is part of Map and acts accordingly
func (e *Nrevent) Unlock() {
	if e.mapSync == nil {
		return
	}
	e.mapSync.Unlock()
}

// Send ...
func (e *Nrevent) Send() {
	e.sender(e)
}

// Attributes ...
func (e *Nrevent) Attributes() *attributes.Attributes {
	return e.attributes
}
