// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package entities

import (
	"sync"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics/samples"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/uid"
)

// Entity ...
type Entity struct {
	attributes *attributes.Attributes
	metrics    *metrics.Map
}

// NewEntity ...
func NewEntity(a *attributes.Attributes) *Entity {
	e := &Entity{attributes: a, metrics: metrics.NewMap()}
	return e
}

// NewMap ...
func NewMap() *Map {
	return &Map{
		collection: Collection{},
		sync:       &sync.RWMutex{},
	}
}

// NewSample ...
func (e *Entity) NewSample(name string, t metrics.Type, unit string, value float64) *Sample {
	return &Sample{sample: samples.NewSample(name, t, unit, value, attributes.NewAttributes()), entity: e}
}

// Attributes returns Attribute struct with methods ...
func (e *Entity) Attributes() *attributes.Attributes {
	return e.attributes
}

// DrainMetrics returns collection of Metrics from collection of Entities
func (e *Entity) DrainMetrics() []*metrics.Metric {
	c := []*metrics.Metric{}
	c = append(c, e.metrics.Drain()...)
	return c
}

// HasMetric helper to metrics.Has()
func (e *Entity) HasMetric(s samples.Sample) (metric *metrics.Metric, found bool) {
	signature := metrics.Signature(s.Signature())
	if metric, found = e.metrics.Has(signature); found {
		return metric, true
	}
	return metric, false
}

// PutMetric ...
func (e *Entity) PutMetric(m *metrics.Metric) {
	e.metrics.Put(m)
}

// AttributeByName ...
func (e *Entity) AttributeByName(name string) *attributes.Attribute {
	return e.Attributes().AttributeByName(name)
}

// Signature of Entity
func (e *Entity) Signature() uid.ID {
	return e.attributes.Signature()
}
