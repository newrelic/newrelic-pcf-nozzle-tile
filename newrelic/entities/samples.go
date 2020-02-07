// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package entities

import (
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics/samples"
)

// Sample ...
type Sample struct {
	entity *Entity
	sample samples.Sample
}

// SetAttribute ...
func (s *Sample) SetAttribute(name string, value interface{}) *Sample {
	s.sample.SetAttribute(name, value)
	return s
}

// Done ...
func (s *Sample) Done() (string, *metrics.Metric) {
	if metric, found := s.entity.HasMetric(s.sample); found {
		metric.Update(s.sample.Value())
		return "DECISION TAKEN: metric already present\n", metric
	}
	metric := s.sample.NewMetric()
	s.entity.PutMetric(metric)
	return "DECISION TAKEN: new metric\n", metric
}
