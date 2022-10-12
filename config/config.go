// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"strings"
	"sync"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const envPrefix = "NRF"

var instance *Config
var once = &sync.Once{}

// Config extending Viper...
type Config struct {
	*viper.Viper
}

// Get Config with defaults
func Get() *Config {
	once.Do(func() {
		instance = set()
	})
	return instance
}

func set() *Config {

	v := viper.New()

	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	// Required environment variables
	for _, s := range []string{
		"CF_API_URL",
		"CF_API_UAA_URL",
		"CF_CLIENT_ID",
		"CF_CLIENT_SECRET",
		"CF_API_USERNAME",
		"CF_API_PASSWORD",
		"NEWRELIC_INSERT_KEY",
		"NEWRELIC_ACCOUNT_ID",
	} {
		if v.GetString(s) == "" {
			logrus.Fatalf("missing required env variable %s_%s", envPrefix, s)
		} else {
			v.BindEnv(s)
		}
	}

	v.BindEnv(EnvCFAPIRUL)
	v.BindEnv("CF_API_UAA_URL")
	rlpgURL := strings.Replace(v.GetString("CF_API_UAA_URL"), "uaa", "log-stream", 1)
	v.SetDefault("CF_API_RLPG_URL", rlpgURL)
	v.BindEnv("CF_API_CLIENT_ID")
	v.BindEnv("CF_API_CLIENT_SECRET")
	// Passing in two strings to override config prefix.
	v.BindEnv("CF_INSTANCE_INDEX", "CF_INSTANCE_INDEX")
	v.BindEnv("CF_INSTANCE_IP", "CF_INSTANCE_IP")

	v.SetDefault("Version", "dev")

	v.SetDefault("CF_SKIP_SSL", true)

	v.SetDefault("HEALTH_PORT", 8080)

	// Cache purge threshold in minutes
	v.SetDefault("FIREHOSE_CACHE_DURATION_MINS", 30)
	// Cache instance update in seconds
	v.SetDefault("FIREHOSE_CACHE_UPDATE_INTERVAL_SECS", 60)
	// Cache instance update in seconds
	v.SetDefault("FIREHOSE_CACHE_WRITE_BUFFER_SIZE", 2048)
	// Rate limiter burst limit
	v.SetDefault("FIREHOSE_RATE_BURST", 5)
	// Rate limiter timeout in seconds.
	v.SetDefault("FIREHOSE_RATE_TIMEOUT_SECS", 60)

	v.BindEnv("NEWRELIC_INSERT_KEY")
	v.BindEnv("NEWRELIC_ACCOUNT_ID")

	v.SetDefault("LOG_LEVEL", "INFO")
	v.SetDefault("TRACER", false)
	v.SetDefault(EnvRabbitMQTags, true)

	v.SetDefault(EnvFirehoseID, "newrelic-firehose")
	v.SetDefault("FIREHOSE_DIODE_BUFFER", 8192)
	v.SetDefault("FIREHOSE_HTTP_TIMEOUT_MINS", 20)
	v.SetDefault("FIREHOSE_RESTART_THRESH_SECS", 15)
	v.SetDefault("NEWRELIC_DRAIN_INTERVAL", "59s")
	v.SetDefault("NEWRELIC_ENQUEUE_TIMEOUT", "1s")

	v.SetDefault(NewRelicEventTypeContainer, "PCFContainerMetric")
	v.SetDefault(NewRelicEventTypeValueMetric, "PCFValueMetric")
	v.SetDefault(NewRelicEventTypeCounterEvent, "PCFCounterEvent")
	v.SetDefault(NewRelicEventTypeLogMessage, "PCFLogMessage")
	v.SetDefault(NewRelicEventTypeHTTPStartStop, "PCFHttpStartStop")

	v.SetDefault("ATTR_PREFIX", "pcf")
	v.SetDefault(EnvEnvelopeType, "envelope.type")
	v.SetDefault(EnvDomain, "domain")
	v.SetDefault(EnvDomainAlias, "bosh.domain")
	v.SetDefault(EnvOrigin, "origin")
	v.SetDefault(EnvDeployment, "deployment")
	v.SetDefault(EnvJob, "job")
	v.SetDefault(EnvIndex, "index")
	v.SetDefault(EnvIP, "IP")
	v.SetDefault(EnvAppID, "app.id")
	v.SetDefault(EnvAppName, "app.name")
	v.SetDefault(EnvAppSpaceName, "app.space.name")
	v.SetDefault(EnvAppOrgName, "app.org.name")
	v.SetDefault(EnvAppInstanceIndex, "app.instance.index")
	v.SetDefault(EnvAppInstanceState, "app.instance.state")
	v.SetDefault(EnvAppInstanceUID, "app.instance.uid")
	v.SetDefault(EnvAppInstancesDesired, "app.instances.desired")
	v.SetDefault(EnvAppRpmId, "app.rpm.id")
	v.SetDefault(EnvAppInsertKey, "app.insert.key")

	// Filtering capabilities for log message events - , or | separated values
	v.SetDefault("LOGMESSAGE_SOURCE_INCLUDE", "")
	v.SetDefault("LOGMESSAGE_SOURCE_EXCLUDE", "")
	v.SetDefault("LOGMESSAGE_MESSAGE_INCLUDE", "")
	v.SetDefault("LOGMESSAGE_MESSAGE_EXCLUDE", "")

	// Filtering capabilities for envelope types - | separated values.
	// By default, all message types are enabled.  User configurations will override this behavior.
	v.SetDefault("ENABLED_ENVELOPE_TYPES", "ContainerMetric|CounterEvent|HttpStartStop|LogMessage|ValueMetric")

	// Default account location will be US unless set to EU by cf push or tile.
	v.SetDefault("NEWRELIC_ACCOUNT_REGION", "US")

	v.SetDefault("NEWRELIC_EU_BASE_URL", "https://insights-collector.eu01.nr-data.net/v1/")

	v.SetDefault("LOGS_LOGMESSAGE", false)
	v.SetDefault("LOGS_HTTP", false)

	config := &Config{v}
	return config
}

// GetNewRelicConfig ...
func (c *Config) GetNewRelicConfig() (key string, id string, region string) {
	key = c.GetString("NEWRELIC_INSERT_KEY")
	id = c.GetString("NEWRELIC_ACCOUNT_ID")
	region = c.GetString("NEWRELIC_ACCOUNT_REGION")
	return
}

func (c *Config) GetRabbitMQConfig() (key string, id string, region string) {
	key = c.GetString("NEWRELIC_INSERT_KEY")
	id = c.GetString("NEWRELIC_ACCOUNT_ID")
	region = c.GetString("NEWRELIC_ACCOUNT_REGION")
	return
}

// AttributeName ...
func (c *Config) AttributeName(n string) string {
	return fmt.Sprintf("%s.%s", c.GetString("ATTR_PREFIX"), c.GetString(n))
}

// GetSelectors ...
func (c *Config) GetSelectors() []*loggregator_v2.Selector {
	e := c.GetString("ENABLED_ENVELOPE_TYPES")
	e = strings.ToLower(e)
	e = strings.Replace(e, " ", "", -1)
	s := make([]*loggregator_v2.Selector, 0)
	// Both ValueMetric and ContainerMetric are Gauge type v2 envelopes
	if strings.Contains(e, "valuemetric") || strings.Contains(e, "containermetric") {
		s = append(s, &loggregator_v2.Selector{Message: &loggregator_v2.Selector_Gauge{Gauge: &loggregator_v2.GaugeSelector{}}})
	}
	if strings.Contains(e, "counterevent") {
		s = append(s, &loggregator_v2.Selector{Message: &loggregator_v2.Selector_Counter{Counter: &loggregator_v2.CounterSelector{}}})
	}
	if strings.Contains(e, "httpstartstop") {
		s = append(s, &loggregator_v2.Selector{Message: &loggregator_v2.Selector_Timer{Timer: &loggregator_v2.TimerSelector{}}})
	}
	if strings.Contains(e, "logmessage") {
		s = append(s, &loggregator_v2.Selector{Message: &loggregator_v2.Selector_Log{Log: &loggregator_v2.LogSelector{}}})
	}
	return s
}

// GetNewEnvelopeTypes ...
func (c *Config) GetNewEnvelopeTypes() []string {
	e := c.GetString("ENABLED_ENVELOPE_TYPES")
	e = strings.ToLower(e)
	e = strings.Replace(e, "logmessage", "log", 1)
	e = strings.Replace(e, "counterevent", "counter", 1)
	e = strings.Replace(e, "httpstartstop", "timer", 1)
	// ENABLED_ENVELOPE_TYPES could be , or | separated.
	e = strings.Replace(e, ",", "|", -1)
	// Remove any spaces added to the configuration
	e = strings.Replace(e, " ", "", -1)
	se := strings.Split(e, "|")
	return se
}
