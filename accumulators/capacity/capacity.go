// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package capacity

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/fatih/camelcase"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/entities"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/insights"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrpcf"
)

type entityID string
type metricKeyword string

func (k metricKeyword) String() string {
	return string(k)
}

func (k metricKeyword) ToLower() string {
	return strings.ToLower(string(k))
}

type capacityMetrics struct {
	Total     *metrics.Metric
	Remaining *metrics.Metric
}

type capacityMap map[metricKeyword]*capacityMetrics

func (cMap capacityMap) Has(s metricKeyword) (cMetrics *capacityMetrics, found bool) {
	cMetrics, found = cMap[s]
	return
}

// CapacityData ...
type capacityData map[*entities.Entity]capacityMap

// CapacityUpdateTime ...
type capacityUpdateTime map[*entities.Entity]time.Time

// Metrics extends metric.Accumulator for
// Firehose ContainerMetric Envelope Event Types
type Metrics struct {
	accumulators.Accumulator
	capacityData       capacityData
	capacityUpdateTime capacityUpdateTime
	sync               *sync.RWMutex
}

// New satisfies metric.Accumulator
func (m Metrics) New() accumulators.Interface {
	i := Metrics{
		Accumulator: accumulators.NewAccumulator(
			// This isn't a v2 envelope type, but the router will route matching Gauge envelopes here.
			"ValueMetric",
		),
		capacityData:       capacityData{},
		capacityUpdateTime: capacityUpdateTime{},
		sync:               &sync.RWMutex{},
	}
	return i
}

// Update satisfies metric.Accumulator
func (m Metrics) Update(e *loggregator_v2.Envelope) {

	if strings.Contains(m.GetTag(e, "job"), "diego_cell") == false {
		return
	}

	entity := m.GetEntity(e, nrpcf.GetPCFAttributes(e))

	g := e.GetGauge()
	// A single v2 gauge envelope can contain multiple metrics.
	for key, met := range g.Metrics {
		target := strings.ToLower(key)
		if strings.Contains(target, "capacity") == false {
			continue
		}
		if strings.Contains(target, "allocated") == true {
			continue
		}

		splits := camelcase.Split(key)

		metric := entity.
			NewSample(
				key,
				metrics.Types.Gauge,
				met.GetUnit(),
				met.GetValue(),
			).
			Done()

		var cMap capacityMap
		var cMetrics *capacityMetrics
		var found bool

		// Lock before making changes to m.capacityData to avoid race conditions
		m.sync.Lock()
		if cMap, found = m.capacityData[entity]; !found {
			cMap = capacityMap{}
			m.capacityData[entity] = cMap
		}

		keyword := metricKeyword(splits[len(splits)-1])

		if cMetrics, found = cMap.Has(keyword); !found {
			cMetrics = &capacityMetrics{nil, nil}
			cMap[keyword] = cMetrics
		}

		metricType := splits[len(splits)-2]

		switch metricType {
		case "Total":
			cMetrics.Total = metric
		case "Remaining":
			cMetrics.Remaining = metric
		}

		m.capacityUpdateTime[entity] = time.Now()
		// Unlock - done making changes to m.capacityData
		m.sync.Unlock()
	}
}

// Drain overrides Accumulator Drain for deriving metrics here
func (m Metrics) Drain() (c []*entities.Entity) {
	// Lock before making changes to m.capacityData to avoid race conditions
	m.sync.Lock()
	// If new metric data hasn't been received in over the threshold defined in CAPACITY_ENTITY_AGE_MINS, drop this entity and its metrics before draining
	ageThreshold := app.Get().Config.GetDuration("CAPACITY_ENTITY_AGE_MINS")
	for k, v := range m.capacityUpdateTime {
		if time.Since(v) >= ageThreshold*time.Minute {
			delete(m.capacityUpdateTime, k)
			delete(m.capacityData, k)
			app.Get().Log.Debugf("\n##Removing entity data for %v. No update in %v minutes.\n", k, ageThreshold)
		}
	}
	// Copying data into another map to reduce the amount of time the lock is needed.
	myCapacityData := capacityData{}
	for k, v := range m.capacityData {
		myCapacityData[k] = v
	}
	// Unlock - done making changes to m.capacityData
	m.sync.Unlock()

	for entity, cMap := range myCapacityData {

		newEntity := entities.NewEntity(entity.Attributes())

		c = append(c, newEntity)

		for metricKeyword, ms := range cMap {

			if ms.Remaining == nil || ms.Total == nil {
				app.Get().Log.Debugf("\nCapacity metrics do not match for %s\n", newEntity.Signature())
				app.Get().Log.Debugf("\nMetric keyword: %v\n", metricKeyword)
				app.Get().Log.Debugf("\nMS: %v\n", ms)
				app.Get().Log.Debugf("\ncMap: %v", cMap)
				continue
			}

			metric := newEntity.NewSample(
				fmt.Sprintf("%s.used", metricKeyword.ToLower()),
				metrics.Types.Gauge,
				"percent",
				100-((ms.Remaining.LastValue/ms.Total.LastValue)*100),
			)

			metric.SetAttribute(
				"metric.source.unit",
				ms.Total.Unit,
			)
			metric.SetAttribute(
				"metric.source.remaining",
				ms.Remaining.Name,
			)
			metric.SetAttribute(
				"metric.source.remaining.value",
				ms.Remaining.LastValue,
			)
			metric.SetAttribute(
				"metric.source.total",
				ms.Total.Name,
			)
			metric.SetAttribute(
				"metric.source.total.value",
				ms.Total.LastValue,
			)

			metric.Done()

		}
	}
	return c
}

// HarvestMetrics ...
func (m Metrics) HarvestMetrics(
	entity *entities.Entity,
	metric *metrics.Metric,
) {

	metric.SetAttribute(
		"eventType",
		"PCFCapacity",
	)

	metric.SetAttribute("agent.subscription", m.Config().GetString("FIREHOSE_ID"))

	metric.Attributes().AppendAll(entity.Attributes())

	// Get a client with the insert key and RPM account ID from the config.
	client := insights.New().Get(app.Get().Config.GetNewRelicConfig())
	client.EnqueueEvent(metric.Marshal())

}

// GetTag ...
func (m Metrics) GetTag(
	e *loggregator_v2.Envelope,
	ta string,
) string {
	if tv, ok := e.Tags[ta]; ok {
		return tv
	}
	return ""
}
