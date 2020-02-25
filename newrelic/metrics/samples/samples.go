// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package samples

import (
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
)

// Sample ...
type Sample struct {
	name       string
	T          metrics.Type
	unit       string
	value      float64
	attributes *attributes.Attributes
}

// NewSample ...
func NewSample(
	name string,
	t metrics.Type,
	unit string,
	value float64,
	attrs *attributes.Attributes,
) (s Sample) {
	s.name = name
	s.T = t
	s.unit = unit
	s.value = value
	s.attributes = attrs
	return
}

// Signature ...
func (s Sample) Signature() (
	name string,
	t metrics.Type,
	unit string,
	attrs *attributes.Attributes,
) {
	return s.name, s.T, s.unit, s.attributes
}

// NewMetric from Sample
func (s *Sample) NewMetric() *metrics.Metric {
	m := metrics.New(s.name, s.T, s.unit, s.value, s.attributes)
	return m
}

// Name of Metric ...
func (s *Sample) Name(name string) *Sample {
	s.name = name
	return s
}

// Unit ...
func (s *Sample) Unit(unit string) *Sample {
	s.unit = unit
	return s
}

// Value as float64...
func (s Sample) Value() float64 {
	return s.value
}

// SetAttribute ...
func (s *Sample) SetAttribute(name string, value interface{}) *Sample {
	if s.attributes == nil {
		s.attributes = &attributes.Attributes{}
	}
	s.attributes.SetAttribute(name, value)
	return s
}

// Return returns Sample args
func (s *Sample) Return() (
	string,
	metrics.Type,
	string,
	float64,
	*attributes.Attributes,
) {
	return s.name, s.T, s.unit, s.value, s.attributes
}
