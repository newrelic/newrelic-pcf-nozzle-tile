// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"reflect"
	"sync"

	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/uid"
)

// Metric is a univeral derivative from events
type Metric struct {
	Name       string  `json:"metric.name"`
	T          Type    `json:"metric.type"`
	Unit       string  `json:"metric.unit"`
	Min        float64 `json:"metric.min"`
	Max        float64 `json:"metric.max"`
	Sum        float64 `json:"metric.sum"`
	LastValue  float64 `json:"metric.sample.last.value"`
	Samples    int     `json:"metric.samples.count"`
	attributes *attributes.Attributes
	Aliases    *attributes.Attributes
	mapSync    *sync.RWMutex
	sender     func(*Metric)
}

// New ...
func New(name string, t Type, unit string, value float64, attrs *attributes.Attributes) *Metric {
	return &Metric{
		Name:       name,
		T:          t,
		Unit:       unit,
		Min:        value,
		Max:        value,
		Sum:        value,
		LastValue:  value,
		attributes: attrs,
		Samples:    1,
		Aliases:    attributes.NewAttributes(),
	}
}

// Update sets last sample to Metric
func (m *Metric) Update(v float64) *Metric {
	m.Lock()
	m.LastValue = v
	m.Sum += v
	if v < m.Min {
		m.Min = v
	}
	if v > m.Max {
		m.Max = v
	}
	m.Samples++
	m.Unlock()
	return m
}

// SetAttribute ...
func (m *Metric) SetAttribute(name string, value interface{}) *Metric {
	m.Lock()
	m.attributes.SetAttribute(name, value)
	m.Unlock()
	return m
}

// Signature of Entity
func (m *Metric) Signature() uid.ID {
	id := m.attributes.Signature()
	id.Concat(m.Name, m.T.String(), m.Unit)
	return id
}

// Attributes ...
func (m *Metric) Attributes() *attributes.Attributes {
	return m.attributes
}

// Signature ...
func Signature(name string, t Type, unit string, attrs *attributes.Attributes) uid.ID {
	id := attrs.Signature()
	id.Concat(name, t, unit)
	return id
}

// Lock determines if metric is part of Map and acts accordingly
func (m *Metric) Lock() {
	if m.mapSync == nil {
		return
	}
	m.mapSync.Lock()
}

// Unlock determines if metric is part of Map and acts accordingly
func (m *Metric) Unlock() {
	if m.mapSync == nil {
		return
	}
	m.mapSync.Unlock()
}

// Marshal Metric ...
func (m *Metric) Marshal() (r *map[string]interface{}) {

	payload := map[string]interface{}{}
	fields := reflect.TypeOf(m).Elem()
	values := reflect.ValueOf(m).Elem()

	for i := 0; i < fields.NumField(); i++ {
		if values.Field(i).Kind() == reflect.Ptr {
			continue
		}
		if name, ok := fields.Field(i).Tag.Lookup("json"); ok {

			value := values.Field(i).Interface()

			switch value.(type) {
			case Type:
				value = value.(Type).String()
			}

			switch value.(type) {
			case string, float64, int:
				if attr := m.Aliases.Has(name); attr != nil {
					payload[attr.Value().(string)] = value
				} else {
					payload[name] = value
				}
			}
		}
	}

	for k, v := range m.Attributes().Marshal() {
		payload[k] = v
	}

	return &payload

}
