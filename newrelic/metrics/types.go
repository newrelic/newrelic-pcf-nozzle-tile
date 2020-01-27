// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metrics

// Type ...
type Type int

func (m Type) String() string {
	switch m {
	case gauge:
		return "Gauge"
	case count:
		return "Count"
	case counter:
		return "Counter"
	case delta:
		return "Delta"
	}
	return "unset"
}

const (
	gauge Type = iota
	count
	counter
	delta
)

// MetricTypes ...
type metricTypes struct {
	Count   Type
	Counter Type
	Gauge   Type
	Delta   Type
}

// Types ...
var Types = metricTypes{
	Count:   count,
	Counter: counter,
	Gauge:   gauge,
	Delta:   delta,
}
