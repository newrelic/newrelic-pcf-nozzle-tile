// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nrpcf

import (
	"net/url"
	"reflect"
	"strings"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/newrelic-client-go/pkg/events"
	"github.com/newrelic/newrelic-client-go/pkg/logs"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/app"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/cfclient/cfapps"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/entities"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrclients"
)

var cfg = config.Get()

var (
	envelopeType     = cfg.AttributeName(config.EnvEnvelopeType)
	domain           = cfg.AttributeName(config.EnvDomain)
	origin           = cfg.AttributeName(config.EnvOrigin)
	deployment       = cfg.AttributeName(config.EnvDeployment)
	job              = cfg.AttributeName(config.EnvJob)
	index            = cfg.AttributeName(config.EnvIndex)
	ip               = cfg.AttributeName(config.EnvIP)
	appID            = cfg.AttributeName(config.EnvAppID)
	appName          = cfg.AttributeName(config.EnvAppName)
	appSpaceName     = cfg.AttributeName(config.EnvAppSpaceName)
	appOrgName       = cfg.AttributeName(config.EnvAppOrgName)
	appInstanceIndex = cfg.AttributeName(config.EnvAppInstanceIndex)
	appInstanceState = cfg.AttributeName(config.EnvAppInstanceState)
	rabbitMqTags     = cfg.AttributeName(config.EnvRabbitMQTags)
)

func inArray(searchStr string, array []string) bool {
	for _, v := range array {
		if v == searchStr { // item found in array of strings
			return true
		}
	}
	return false
}

// GetPCFAttributes ...
func GetPCFAttributes(e *loggregator_v2.Envelope) *attributes.Attributes {
	attrs := EntityAttributes(e)

	if app.Get().Config.GetBool(config.EnvRabbitMQTags) {
		if e.Tags["origin"] == "p.rabbitmq" {
			tags := e.GetTags()
			for name, val := range tags {
				if !inArray(name, []string{"Gauge", "Count", "Counter", "Delta", "origin", "deployment", "job", "index", "ip"}) {
					attrs.SetAttribute("tags."+name, val)
				}
			}
		}
	}

	et := reflect.TypeOf(e.Message).String()
	if et == "*loggregator_v2.Envelope_Gauge" {
		if isContainerMetric(e) {
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
	attrs.SetAttribute(domain, PCFDomain())
	attrs.SetAttribute(cfg.GetString(config.EnvDomainAlias), PCFDomain())
	attrs.SetAttribute("agent.version", cfg.GetString("Version"))
	attrs.SetAttribute("agent.instance", cfg.GetInt("CF_INSTANCE_INDEX"))
	attrs.SetAttribute("agent.ip", cfg.GetString("CF_INSTANCE_IP"))
	return attrs
}

// PCFDomain ...
func PCFDomain() string {
	u, _ := url.Parse(app.Get().Config.GetString(config.EnvCFAPIRUL))
	return u.Hostname()
}

// EntityAttributes ...
func EntityAttributes(e *loggregator_v2.Envelope) *attributes.Attributes {
	et := strings.Split(reflect.TypeOf(e.Message).String(), "_")
	return attributes.NewAttributes(
		attributes.New(envelopeType, et[len(et)-1]),
		attributes.New(origin, e.Tags["origin"]),
		attributes.New(deployment, e.Tags["deployment"]),
		attributes.New(job, e.Tags["job"]),
		attributes.New(index, e.Tags["index"]),
		attributes.New(ip, e.Tags["ip"]),
	)
}

// EntityContainerAttributes ...
func EntityContainerAttributes(e *loggregator_v2.Envelope) *attributes.Attributes {
	return attributes.NewAttributes(
		attributes.New(appID, e.GetSourceId()),
		attributes.New(appInstanceIndex, e.GetInstanceId()),
	)
}

// EntityLogAttributes ...
func EntityLogAttributes(e *loggregator_v2.Envelope) *attributes.Attributes {
	return attributes.NewAttributes(
		attributes.New(appID, e.GetSourceId()),
		attributes.New(appInstanceIndex, e.GetInstanceId()),
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
func GetInsertClientForApp(e *entities.Entity) (c *events.Events) {

	guid := e.AttributeByName(config.Get().AttributeName(config.EnvAppID)).Value()
	cfapp := cfapps.GetInstance().GetApp(guid.(string))
	cm := nrclients.New()

	cfapp.Lock.RLock()
	vcap := cfapp.VcapServices
	cfapp.Lock.RUnlock()

	if vcap == nil {
		defaultClient := cm.GetEventClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}

	//Can do this if newrelic isn't found, but also need to check for rpmAccountId and insightsInsertKey values
	if _, found := vcap["newrelic"]; !found {
		defaultClient := cm.GetEventClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}

	newrelicSlice := vcap["newrelic"].([]interface{})
	newrelic := newrelicSlice[0].(map[string]interface{})

	// Get the credentials map from inside of the newrelic map, if it exists.
	if _, found := newrelic["credentials"].(map[string]interface{}); !found {
		defaultClient := cm.GetEventClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}
	credentials := newrelic["credentials"].(map[string]interface{})

	// Call GetInsertKey
	insertKey, found := GetInsertKey(credentials)
	if !found {
		defaultClient := cm.GetEventClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}
	// Call GetRpmId
	rpmId, found := GetRpmId(credentials)
	if !found {
		defaultClient := cm.GetEventClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}

	// Call GetLicenseKey
	licenseKey, found := GetLicenseKey(credentials)
	if !found {
		defaultClient := cm.GetEventClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}

	isEU := strings.HasPrefix(licenseKey, "eu01x")
	var accountRegion string
	if isEU {
		accountRegion = "EU"
	} else {
		accountRegion = "US"
	}

	// Call Get from NR clients manager to get a client with this configuration.
	c = cm.GetEventClient(insertKey, rpmId, accountRegion)

	return c

}

// GetLogClientForApp checks app for newrelic plan sub-account insert creds
// and return insight client from insert manager/cache or new.
// If app does not have a plan, this returns the main account credentials (from the config file)
func GetLogClientForApp(e *entities.Entity) (c *logs.Logs) {

	guid := e.AttributeByName(config.Get().AttributeName(config.EnvAppID)).Value()
	cfapp := cfapps.GetInstance().GetApp(guid.(string))
	cm := nrclients.New()

	cfapp.Lock.RLock()
	vcap := cfapp.VcapServices
	cfapp.Lock.RUnlock()

	if vcap == nil {
		defaultClient := cm.GetLogClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}

	//Can do this if newrelic isn't found, but also need to check for rpmAccountId and insightsInsertKey values
	if _, found := vcap["newrelic"]; !found {
		defaultClient := cm.GetLogClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}

	newrelicSlice := vcap["newrelic"].([]interface{})
	newrelic := newrelicSlice[0].(map[string]interface{})

	// Get the credentials map from inside of the newrelic map, if it exists.
	if _, found := newrelic["credentials"].(map[string]interface{}); !found {
		defaultClient := cm.GetLogClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}
	credentials := newrelic["credentials"].(map[string]interface{})

	// Call GetInsertKey
	insertKey, found := GetInsertKey(credentials)
	if !found {
		defaultClient := cm.GetLogClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}
	// Call GetRpmId
	rpmId, found := GetRpmId(credentials)
	if !found {
		defaultClient := cm.GetLogClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}

	// Call GetLicenseKey
	licenseKey, found := GetLicenseKey(credentials)
	if !found {
		defaultClient := cm.GetLogClient(app.Get().Config.GetNewRelicConfig())
		return defaultClient
	}

	isEU := strings.HasPrefix(licenseKey, "eu01x")
	var accountRegion string
	if isEU {
		accountRegion = "EU"
	} else {
		accountRegion = "US"
	}

	// Call Get from NR clients manager to get a client with this configuration.
	c = cm.GetLogClient(insertKey, rpmId, accountRegion)

	return c

}

// isContainerMetric determines if the current v2 Gauge envelope is a v1 ContainerMetric or v1 ValueMetric
func isContainerMetric(e *loggregator_v2.Envelope) bool {
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
