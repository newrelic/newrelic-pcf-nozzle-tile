// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package logmessage

import (
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/cfclient/cfapps"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/config"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/accumulators"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/attributes"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/entities"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/metrics"
	"github.com/newrelic/newrelic-pcf-nozzle-tile/newrelic/nrpcf"
)

// Nrevents extends event.Accumulator for
// Firehose LogMessage Envelope Event Types
type Nrevents struct {
	accumulators.Accumulator
	CFAppManager *cfapps.CFAppManager
}

// New satisfies event.Accumulator
func (n Nrevents) New() accumulators.Interface {
	i := Nrevents{
		Accumulator: accumulators.NewAccumulator(
			"*loggregator_v2.Envelope_Log",
		),
		CFAppManager: cfapps.GetInstance(),
	}
	return i
}

// Update satisfies event.Accumulator
// func (n Nrevents) Update(e *events.Envelope) {
func (n Nrevents) Update(e *loggregator_v2.Envelope) {

	// Check filters first.  Other work can be avoided if filters aren't matched
	// Include filters will be checked first, then exclude filters
	// Stop processing this envelope if the include filters are not matched
	if n.IsIncluded(string(e.GetLog().Payload), e.Tags["source_type"]) == false {
		return
	}

	// Stop processing this envelope if the exclude filters are matched
	if n.IsExcluded(string(e.GetLog().Payload), e.Tags["source_type"]) == true {
		return
	}

	entity := n.GetEntity(e, nrpcf.GetPCFAttributes(e))

	logEntry := attributes.NewAttributes()

	// Append application instance attributes to the log entry.
	logEntry.AppendAll(n.CFAppManager.GetAppInstanceAttributes(e.GetSourceId(), n.ConvertSourceInstance(e.GetInstanceId())))

	// msgContent := e.GetLogMessage().GetMessage()
	msgContent := e.GetLog().Payload

	// Mesages over 4K in length will be rejected by the Event API.  Trim the message before sending.
	if len(msgContent) > 4096 {
		msgContent = msgContent[0:4095]
		logEntry.SetAttribute("log.message.truncated", true)
	}

	// Add log message attributes
	logEntry.SetAttribute("log.message", string(msgContent))
	logEntry.SetAttribute("log.timestamp", time.Unix(0, e.GetTimestamp()))
	logEntry.SetAttribute("log.app.id", e.GetSourceId())
	logEntry.SetAttribute("log.source.type", n.GetTag(e, "source_type"))
	logEntry.SetAttribute("log.source.instance", e.GetInstanceId())
	logEntry.SetAttribute("log.message.type", n.getLogMessageType(e.GetLog()))
	logEntry.SetAttribute(
		"eventType",
		config.Get().GetString(config.NewRelicEventTypeLogMessage),
	)
	logEntry.SetAttribute("agent.subscription", config.Get().GetString("FIREHOSE_ID"))

	logEntry.AppendAll(entity.Attributes())
	client := nrpcf.GetInsertClientForApp(entity)
	client.EnqueueEvent(logEntry.Marshal())
}

// HarvestMetrics - stub for LogMessages, which are all events...
func (n Nrevents) HarvestMetrics(entity *entities.Entity, metric *metrics.Metric) {
}

// IsIncluded ...
func (n Nrevents) IsIncluded(logMessage string, logSource string) bool {
	matchFound := false
	srcMatchFound := n.IsIncludedLogSource(logSource)
	msgMatchFound := n.IsIncludedLogMessage(logMessage)
	if srcMatchFound && msgMatchFound {
		matchFound = true
	}
	return matchFound
}

// IsExcluded ...
func (n Nrevents) IsExcluded(logMessage string, logSource string) bool {
	// First include filter is a logSource include.
	if n.IsExcludedLogSource(logSource) {
		return true
	}
	// Second include filter is a message include.
	if n.IsExcludedLogMessage(logMessage) {
		return true
	}
	return false
}

// IsExcludedLogSource determines if envelopes with this log source should be dropped.
func (n Nrevents) IsExcludedLogSource(logSource string) bool {
	if config.Get().GetFilter("LOGMESSAGE_SOURCE_EXCLUDE") != nil {
		for _, filter := range config.Get().GetFilter("LOGMESSAGE_SOURCE_EXCLUDE") {
			if strings.Compare(strings.TrimSpace(filter), logSource) == 0 {
				return true
			}
		}
	}
	return false
}

// IsExcludedLogMessage determines if envelopes with this log message should be dropped.
func (n Nrevents) IsExcludedLogMessage(logMessage string) bool {
	if config.Get().GetFilter("LOGMESSAGE_MESSAGE_EXCLUDE") != nil {
		for _, filter := range config.Get().GetFilter("LOGMESSAGE_MESSAGE_EXCLUDE") {
			if strings.Contains(logMessage, strings.TrimSpace(filter)) {
				return true
			}
		}
	}
	return false
}

// IsIncludedLogSource determines if envelopes with this log source should be included.
func (n Nrevents) IsIncludedLogSource(logSource string) bool {
	if config.Get().GetFilter("LOGMESSAGE_SOURCE_INCLUDE") != nil {
		for _, filter := range config.Get().GetFilter("LOGMESSAGE_SOURCE_INCLUDE") {
			if strings.Compare(strings.TrimSpace(filter), logSource) == 0 {
				return true
			}
		}
	} else {
		return true
	}
	return false
}

// IsIncludedLogMessage determines if envelopes with this log message should be included.
func (n Nrevents) IsIncludedLogMessage(logMessage string) bool {
	if config.Get().GetFilter("LOGMESSAGE_MESSAGE_INCLUDE") != nil {
		for _, filter := range config.Get().GetFilter("LOGMESSAGE_MESSAGE_INCLUDE") {
			if strings.Contains(logMessage, strings.TrimSpace(filter)) {
				return true
			}
		}
	} else {
		return true
	}
	return false
}

// ConvertSourceInstance from a string to int32
func (n Nrevents) ConvertSourceInstance(i string) int32 {
	if num, err := strconv.ParseInt(i, 10, 32); err == nil {
		return int32(num)
	}
	return 0
}

// getLogMessageType returns the message type (OUT or ERR)
func (n Nrevents) getLogMessageType(log *loggregator_v2.Log) string {
	if log.Type == loggregator_v2.Log_OUT {
		return "OUT"
	}
	return "ERR"
}

// GetTag ...
func (n Nrevents) GetTag(
	e *loggregator_v2.Envelope,
	ta string,
) string {
	if tv, ok := e.Tags[ta]; ok {
		return tv
	}
	return ""
}
