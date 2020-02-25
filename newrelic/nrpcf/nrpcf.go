// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nrpcf

import (
	"net/url"
	"reflect"
	"strings"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/go-insights/client"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/cfclient/cfapps"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/entities"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/insights"
)

// GetPCFAttributes ...
func GetPCFAttributes(e *loggregator_v2.Envelope) *attributes.Attributes {
	attrs := EntityAttributes(e)
	et := reflect.TypeOf(e.Message).String()
	if et == "*loggregator_v2.Envelope_Gauge" {
		if IsContainerMetric(e) {
			et = "ContainerMetric"
		} else {
			et = "ValueMetric"
		}
	}
	if strings.Contains(et, "Envelope_Log") {
		for _, a := range EntityLogAttributes(e).Get() {
			attrs.Append(a)
		}
	}
	if strings.Contains(et, "ContainerMetric") {
		for _, a := range EntityContainerAttributes(e).Get() {
			attrs.Append(a)
		}
	}
	cfg := config.Get()
	attrs.SetAttribute(cfg.AttributeName(config.EnvDomain), PCFDomain())
	attrs.SetAttribute(cfg.GetString(config.EnvDomainAlias), PCFDomain())
	attrs.SetAttribute("agent.version", cfg.GetString("Version"))
	attrs.SetAttribute("agent.instance", cfg.GetInt("CF_INSTANCE_INDEX"))
	attrs.SetAttribute("agent.ip", cfg.GetString("CF_INSTANCE_IP"))
	return attrs
}

// PCFDomain ...
func PCFDomain() string {
	u, _ := url.Parse(config.Get().GetString(config.EnvCFAPIRUL))
	return u.Hostname()
}

// EntityAttributes ...
func EntityAttributes(e *loggregator_v2.Envelope) *attributes.Attributes {
	et := strings.Split(reflect.TypeOf(e.Message).String(), "_")
	cfg := config.Get()
	return attributes.NewAttributes(
		attributes.New(cfg.AttributeName(config.EnvEnvelopeType), et[len(et)-1]),
		attributes.New(cfg.AttributeName(config.EnvOrigin), e.Tags["origin"]),
		attributes.New(cfg.AttributeName(config.EnvDeployment), e.Tags["deployment"]),
		attributes.New(cfg.AttributeName(config.EnvJob), e.Tags["job"]),
		attributes.New(cfg.AttributeName(config.EnvIndex), e.Tags["index"]),
		attributes.New(cfg.AttributeName(config.EnvIP), e.Tags["ip"]),
	)
}

// EntityContainerAttributes ...
func EntityContainerAttributes(e *loggregator_v2.Envelope) *attributes.Attributes {
	cfg := config.Get()
	return attributes.NewAttributes(
		attributes.New(cfg.AttributeName(config.EnvAppID), e.GetSourceId()),
		attributes.New(cfg.AttributeName(config.EnvAppInstanceIndex), e.GetInstanceId()),
	)
}

// EntityLogAttributes ...
func EntityLogAttributes(e *loggregator_v2.Envelope) *attributes.Attributes {
	cfg := config.Get()
	return attributes.NewAttributes(
		attributes.New(cfg.AttributeName(config.EnvAppID), e.GetSourceId()),
		attributes.New(cfg.AttributeName(config.EnvAppInstanceIndex), e.GetInstanceId()),
	)
}

// GetRpmId from the credentials map
func GetRpmId(credentials map[string]interface{}) (string, bool) {
	rpmId, found := credentials["rpmAccountId"].(string)
	if found && len(rpmId) == 0 {
		found = false
	}
	return rpmId, found
}

// GetInsertKey from the credentials map
func GetInsertKey(credentials map[string]interface{}) (string, bool) {
	insertKey, found := credentials["insightsInsertKey"].(string)
	if found && len(insertKey) == 0 {
		found = false
	}
	return insertKey, found
}

// GetLicenseKey from the credentials map
func GetLicenseKey(credentials map[string]interface{}) (string, bool) {
	licenseKey, found := credentials["licenseKey"].(string)
	if found && len(licenseKey) == 0 {
		found = false
	}
	return licenseKey, found
}

// GetInsertClientForApp checks app for newrelic plan sub-account insert creds
// and return insight client from insert manager/cache or new.
// If app does not have a plan, this returns the main account credentials (from the config file)
func GetInsertClientForApp(e *entities.Entity) (c *client.InsertClient) {

	guid := e.AttributeByName(config.Get().AttributeName(config.EnvAppID)).Value()
	cfapp := cfapps.GetInstance().GetApp(guid.(string))
	im := insights.New()

	cfapp.Lock.RLock()
	vcap := cfapp.VcapServices
	cfapp.Lock.RUnlock()

	if vcap == nil {
		defaultClient := im.Get(config.Get().GetNewRelicConfig())
		return defaultClient
	}

	//Can do this if newrelic isn't found, but also need to check for rpmAccountId and insightsInsertKey values
	if _, found := vcap["newrelic"]; !found {
		defaultClient := im.Get(config.Get().GetNewRelicConfig())
		return defaultClient
	}

	newrelicSlice := vcap["newrelic"].([]interface{})
	newrelic := newrelicSlice[0].(map[string]interface{})

	// Get the credentials map from inside of the newrelic map, if it exists.
	if _, found := newrelic["credentials"].(map[string]interface{}); !found {
		defaultClient := im.Get(config.Get().GetNewRelicConfig())
		return defaultClient
	}
	credentials := newrelic["credentials"].(map[string]interface{})

	// Call GetInsertKey
	insertKey, found := GetInsertKey(credentials)
	if !found {
		defaultClient := im.Get(config.Get().GetNewRelicConfig())
		return defaultClient
	}
	// Call GetRpmId
	rpmId, found := GetRpmId(credentials)
	if !found {
		defaultClient := im.Get(config.Get().GetNewRelicConfig())
		return defaultClient
	}

	// Call GetLicenseKey
	licenseKey, found := GetLicenseKey(credentials)
	if !found {
		defaultClient := im.Get(config.Get().GetNewRelicConfig())
		return defaultClient
	}

	isEU := strings.HasPrefix(licenseKey, "eu01x")
	var accountRegion string
	if isEU {
		accountRegion = "EU"
	} else {
		accountRegion = "US"
	}

	// Call Get from Insights manager to get a client with this configuration.
	c = im.Get(insertKey, rpmId, accountRegion)

	return c

}

// IsContainerMetric determines if the current v2 Gauge envelope is a v1 ContainerMetric or v1 ValueMetric
func IsContainerMetric(e *loggregator_v2.Envelope) bool {
	gauge := e.GetGauge()
	if len(gauge.Metrics) != 5 {
		return false
	}
	required := []string{
		"cpu",
		"memory",
		"disk",
		"memory_quota",
		"disk_quota",
	}

	for _, req := range required {
		if _, found := gauge.Metrics[req]; !found {
			return false
		}
	}
	return true
}
