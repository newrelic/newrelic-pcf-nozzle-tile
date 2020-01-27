// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package metrics

import "reflect"

// JSONDefaults for JSONMap
var JSONDefaults JSONMap

func init() {
	JSONDefaults = JSONMap{
		"EnvelopeType":     "eventType",
		"Origin":           "pcf.origin",
		"Deployement":      "pcf.deployment",
		"Job":              "pcf.job",
		"Index":            "pcf.index",
		"IP":               "pcf.ip",
		"AppGUID":          "pcf.app.guid",
		"AppName":          "pcf.app.name",
		"AppSpaceName":     "pcf.app.space.name",
		"AppOrgName":       "pcf.app.org.name",
		"AppInstanceIndex": "pcf.app.instance.index",
		"AppInstanceState": "pcf.app.instance.state",
		"Name":             "pcf.metric",
		"Unit":             "pcf.metric.unit",
		"Min":              "pcf.metric.min",
		"Max":              "pcf.metric.max",
		"Total":            "pcf.metric.total",
		"Quota":            "pcf.metric.quota",
		"Count":            "pcf.metric.samples",
		"Average":          "pcf.metric.average",
		"QuotaUsed":        "pcf.metric.quota.used",
		"Drift":            "pcf.metric.drift",
	}
}

// JSONMap Insights Attributes
type JSONMap map[string]interface{}

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
