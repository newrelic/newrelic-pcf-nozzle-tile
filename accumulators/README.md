# **Accumulators**

Accumulators subscribe to streams of specific event types and the [router](../newrelic/router.go) forwards matching Loggregator envelopes to the accumulators for processing. 

## **Event Types**

All metrics include the PCF meta data. Only PCFContainerMetric and PCFLogMessage hold application specific data. All other metrics pertain to PCF System metrics.

| Event Type | Loggregator Envelope Type | Description | Accumulator |
| :--- | :--- | :--- | :--- |
| PCFContainerMetric | ContainerMetric | Application specific metrics | [`accumulators/container/container.go`](container/container.go)
| PCFValueMetric | ValueMetric | PCF System metrics of multiple metric types | [`accumulators/value/value.go`](value/value.go)
| PCFCounterEvent | CounterEvent | PCF System metrics as counter types only | [`accumulators/counter/counter.go`](counter/counter.go)
| PCFCapacity | ValueMetric | PCF System metric, derived from Total and Remaining samples in order to provide percent used. | [`accumulators/capacity/capacity.go`](capacity/capacity.go)
| PCFLogMessage | LogMessage | PCF Logs | [`accumulators/logmessage/logmessage.go`](logmessage/logmessage.go)
| PCFHttpStartStop | HttpStartStop | PCF HTTP request details | [`accumulators/http/http.go`](http/http.go)