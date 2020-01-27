// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package entities

import (
	"sync"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics/samples"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrevents"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/uid"
)

// Entity ...
type Entity struct {
	uid        uid.ID
	attributes *attributes.Attributes
	metrics    *metrics.Map
	nrevents   *nrevents.Nreventmap
}

// NewEntity ...
func NewEntity(
	a *attributes.Attributes,
) *Entity {
	e := &Entity{
		attributes: a,
		metrics:    metrics.NewMap(),
		nrevents:   nrevents.NewMap(),
	}
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
func (e *Entity) NewSample(
	name string,
	t metrics.Type,
	unit string,
	value float64,
) *Sample {
	return &Sample{
		sample: samples.NewSample(
			name,
			t,
			unit,
			value,
			attributes.NewAttributes()),
		entity: e,
	}
}

// Attributes returns Attribute struct with methods ...
func (e *Entity) Attributes() *attributes.Attributes {
	return e.attributes
}

// SetAttribute ...
func (e *Entity) SetAttribute(
	name string,
	value interface{},
) *attributes.Attribute {
	return e.Attributes().
		SetAttribute(name, value)
}

// ForEachMetric ...
func (e *Entity) ForEachMetric(fn func(*metrics.Metric)) int {
	return e.metrics.ForEach(fn)
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

// AddAttributesToMap ...
func (e *Entity) AddAttributesToMap(
	m *map[string]interface{},
) *map[string]interface{} {
	for _, a := range e.Attributes().Get() {
		(*m)[a.Name()] = a.Value
	}
	return m
}

// AddAttributesToMetric ...
func (e *Entity) AddAttributesToMetric(m *metrics.Metric) {
	for _, a := range e.Attributes().Get() {
		m.Attributes().Append(a)
	}
}

// AttributeByName ...
func (e *Entity) AttributeByName(name string) *attributes.Attribute {
	return e.Attributes().AttributeByName(name)
}

// MetricCount ...
func (e *Entity) MetricCount() int {
	return e.metrics.Count()
}

// Signature of Entity
func (e *Entity) Signature() uid.ID {
	return e.attributes.Signature()
}
