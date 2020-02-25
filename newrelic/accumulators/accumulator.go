// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package accumulators ...
// count = Total = last sample value
// gauge = Total = last sample value
// counter = Total = sum of sample values
package accumulators

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/entities"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
)

// EnvelopeTypes ..
type EnvelopeTypes []string

// Interface ensures Firehose Envelope consumption
type Interface interface {
	New() Interface
	Update(*loggregator_v2.Envelope)
	Streams() []string
	HarvestMetrics(*entities.Entity, *metrics.Metric)
	Drain() []*entities.Entity
}

// Accumulator Universal handler for Firehose Envelopes
type Accumulator struct {
	Entities      *entities.Map
	EnvelopeTypes []string
	ctx           *app.Application
}

// NewAccumulator is generic and requires .Interface to be set
// func NewAccumulator(t ...events.Envelope_EventType) Accumulator {
func NewAccumulator(t ...string) Accumulator {
	types := make(EnvelopeTypes, len(t))
	for _, envelopType := range t {
		types = append(types, envelopType)
	}
	return Accumulator{
		Entities:      entities.NewMap(),
		EnvelopeTypes: types,
		ctx:           app.Get(),
	}
}

// GetEntity ...
func (a *Accumulator) GetEntity(e *loggregator_v2.Envelope, attrs *attributes.Attributes) *entities.Entity {
	if e, found := a.Entities.Has(attrs.Signature()); found {
		return e
	}
	entity := entities.NewEntity(attrs)
	a.Entities.Put(entity)
	return entity
}

// Drain ...
func (a Accumulator) Drain() []*entities.Entity {
	return a.Entities.Drain()
}

// Streams ...
func (a Accumulator) Streams() []string {
	return a.EnvelopeTypes
}
