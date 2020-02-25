// +build integration

package test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	mocks "github.com/newrelic/newrelic-pcf-nozzle-tile/tests/integration/helpers"
)

var port int
var mutex = &sync.Mutex{}

const (
	PCFContainerMetric = "PCFContainerMetric"
	PCFValueMetric     = "PCFValueMetric"
	PCFHttpStartStop   = "PCFHttpStartStop"
	PCFLogMessage      = "PCFLogMessage"
	PCFCounterEvent    = "PCFCounterEvent"
)

type apiMocks struct {
	uaa          *mocks.MockUAAC
	firehose     *mocks.MockFirehose
	cc           *mocks.MockCF
	insights     *mocks.MockInsights
	nozzle       *exec.Cmd
	nozzleStdOut io.ReadCloser
}

func runNozzleAndMocks() *apiMocks {
	m := &apiMocks{
		uaa:      mocks.NewMockUAA("bearer", "token"),
		cc:       mocks.NewMockCF("bearer", "token"),
		insights: mocks.NewMockInsights("Gzip"),
		nozzle:   exec.Command("../../dist/nr-fh-nozzle"),
	}
	m.firehose = mocks.NewMockFirehose(360, "token")
	m.uaa.Start()
	m.firehose.Start()
	m.cc.Start()
	m.insights.Start()

	os.Setenv("NRF_CF_API_URL", m.cc.Server.URL)
	os.Setenv("NRF_CF_API_UAA_URL", m.uaa.Server.URL)
	os.Setenv("NRF_CF_CLIENT_ID", "admin")
	os.Setenv("NRF_CF_CLIENT_SECRET", "token")
	os.Setenv("NRF_CF_API_USERNAME", "admin")
	os.Setenv("NRF_CF_API_PASSWORD", "token")
	os.Setenv("NRF_NEWRELIC_INSERT_KEY", "nrkey")
	os.Setenv("NRF_NEWRELIC_ACCOUNT_ID", "00000")
	os.Setenv("NRF_NEWRELIC_DRAIN_INTERVAL", "500ms")
	os.Setenv("NRF_NEWRELIC_CUSTOM_URL", m.insights.Server.URL)
	os.Setenv("NRF_CF_API_RLPG_URL", m.firehose.Server.URL)
	os.Setenv("NRF_HEALTH_PORT", "8080")

	m.nozzleStdOut, _ = m.nozzle.StdoutPipe()

	m.nozzle.Start()

	return m
}
func closeNozzleAndMocks(a *apiMocks) {
	a.nozzle.Process.Kill()
	a.uaa.Stop()
	a.firehose.Stop()
	a.cc.Stop()
	a.insights.Stop()

}

func TestValueMetric(t *testing.T) {
	m := runNozzleAndMocks()
	for i := float64(1); i < 11; i++ {
		e := loggregator_v2.Envelope{
			SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
			InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
			Message: &loggregator_v2.Envelope_Gauge{
				Gauge: &loggregator_v2.Gauge{
					Metrics: map[string]*loggregator_v2.GaugeValue{
						"name": &loggregator_v2.GaugeValue{
							Unit:  "counter",
							Value: i,
						},
					},
				},
			},
		}
		m.firehose.AddEvent(e)
	}

	m.firehose.PublishBatch()
	rc := readInsights(t, m)
	closeNozzleAndMocks(m)

	assert.EqualValues(t, 0, len(m.insights.ReceivedContents))

	r := make([]map[string]interface{}, 10)
	json.Unmarshal([]byte(rc), &r)

	assert.EqualValues(t, "name", r[0]["metric.name"])
	assert.EqualValues(t, 10, r[0]["metric.sample.last.value"])
	assert.EqualValues(t, 10, r[0]["metric.max"])
	assert.EqualValues(t, 1, r[0]["metric.min"])
	assert.EqualValues(t, 10, r[0]["metric.samples.count"])
	assert.EqualValues(t, 55, r[0]["metric.sum"])
	assert.EqualValues(t, PCFValueMetric, r[0]["eventType"])

}

func TestCapacityMetric(t *testing.T) {
	m := runNozzleAndMocks()
	m.firehose.AddEvent(loggregator_v2.Envelope{
		Tags: map[string]string{
			"job": "diego_cell",
		},
		SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: map[string]*loggregator_v2.GaugeValue{
					"CapacityRemainingContainers": &loggregator_v2.GaugeValue{
						Unit:  "bytes",
						Value: float64(25),
					},
					"CapacityTotalContainers": &loggregator_v2.GaugeValue{
						Unit:  "bytes",
						Value: float64(100),
					},
				},
			},
		},
	})
	m.firehose.PublishBatch()

	rCapacity := make(map[string]interface{})
ReadingFromInsights:
	for {
		select {
		case rc := <-m.insights.ReceivedContents:
			r := make([]map[string]interface{}, 10)
			json.Unmarshal([]byte(rc), &r)
			for i, rr := range r {
				et := rr["eventType"].(string)
				if et == "PCFCapacity" {
					rCapacity = r[i]
					break ReadingFromInsights
				}
			}
		case <-time.After(10 * time.Second):
			break ReadingFromInsights
		}
	}
	closeNozzleAndMocks(m)

	assert.EqualValues(t, "PCFCapacity", rCapacity["eventType"])
	assert.EqualValues(t, 75, rCapacity["metric.sample.last.value"])
	assert.EqualValues(t, "CapacityRemainingContainers", rCapacity["metric.source.remaining"])
	assert.EqualValues(t, 25, rCapacity["metric.source.remaining.value"])
	assert.EqualValues(t, "CapacityTotalContainers", rCapacity["metric.source.total"])
	assert.EqualValues(t, 100, rCapacity["metric.source.total.value"])
	assert.EqualValues(t, "containers.used", rCapacity["metric.name"])
	assert.EqualValues(t, 75, rCapacity["metric.sample.last.value"])

}

func TestLogMessage(t *testing.T) {
	m := runNozzleAndMocks()

	m.firehose.AddEvent(loggregator_v2.Envelope{
		SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		Message: &loggregator_v2.Envelope_Log{
			Log: &loggregator_v2.Log{
				Payload: []byte("logtest"),
				Type:    loggregator_v2.Log_OUT,
			},
		},
	})
	m.firehose.PublishBatch()
	rc := readInsights(t, m)
	closeNozzleAndMocks(m)

	assert.EqualValues(t, 0, len(m.insights.ReceivedContents))

	r := make([]map[string]interface{}, 10)
	json.Unmarshal([]byte(rc), &r)

	assert.EqualValues(t, "logtest", r[0]["log.message"])
	assert.EqualValues(t, PCFLogMessage, r[0]["eventType"])

}

func TestContainerMetric(t *testing.T) {
	m := runNozzleAndMocks()

	m.firehose.AddEvent(loggregator_v2.Envelope{
		SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: map[string]*loggregator_v2.GaugeValue{
					"cpu": &loggregator_v2.GaugeValue{
						Unit:  "percent",
						Value: float64(2),
					},
					"memory": &loggregator_v2.GaugeValue{
						Unit:  "bytes",
						Value: float64(10),
					},
					"disk": &loggregator_v2.GaugeValue{
						Unit:  "bytes",
						Value: float64(25),
					},
					"memory_quota": &loggregator_v2.GaugeValue{
						Unit:  "bytes",
						Value: float64(1000),
					},
					"disk_quota": &loggregator_v2.GaugeValue{
						Unit:  "bytes",
						Value: float64(2000),
					},
				},
			},
		},
	})
	m.firehose.PublishBatch()
	rc := readInsights(t, m)
	closeNozzleAndMocks(m)

	assert.EqualValues(t, 0, len(m.insights.ReceivedContents))

	r := make([]map[string]interface{}, 10)
	json.Unmarshal([]byte(rc), &r)

	for _, metric := range r {
		switch {
		case metric["metric.name"] == "app.cpu":
			assert.EqualValues(t, 2, metric["metric.sample.last.value"])
		case metric["metric.name"] == "app.disk":
			assert.EqualValues(t, 25, metric["metric.sample.last.value"])
			assert.EqualValues(t, 2000, metric["app.disk.quota"])
		case metric["metric.name"] == "app.memory":
			assert.EqualValues(t, 10, metric["metric.sample.last.value"])
			assert.EqualValues(t, 1000, metric["app.memory.quota"])
		default:
			assert.Fail(t, "metric.name not expected")
		}
	}

}
func TestCounterEvent(t *testing.T) {
	m := runNozzleAndMocks()

	m.firehose.AddEvent(loggregator_v2.Envelope{
		SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name:  "name",
				Delta: 10,
				Total: 100,
			},
		},
	})
	m.firehose.PublishBatch()
	rc := readInsights(t, m)
	closeNozzleAndMocks(m)

	assert.EqualValues(t, 0, len(m.insights.ReceivedContents))

	r := make([]map[string]interface{}, 10)
	json.Unmarshal([]byte(rc), &r)

	assert.EqualValues(t, "name", r[0]["metric.name"])
	assert.EqualValues(t, 10, r[0]["metric.sample.last.value"])
	assert.EqualValues(t, 100, r[0]["total.reported"])
	assert.EqualValues(t, PCFCounterEvent, r[0]["eventType"])

}
func TestHTTPStartStop(t *testing.T) {
	m := runNozzleAndMocks()

	m.firehose.AddEvent(loggregator_v2.Envelope{
		SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		Tags: map[string]string{
			"uri":        "uri",
			"method":     "GET",
			"peer_type":  "Server",
			"user_agent": "Go-http-client/1.1",
		},
		Message: &loggregator_v2.Envelope_Timer{
			Timer: &loggregator_v2.Timer{
				Name:  "test",
				Start: 10000000,
				Stop:  11000000,
			},
		},
	})
	m.firehose.PublishBatch()
	rc := readInsights(t, m)
	closeNozzleAndMocks(m)

	assert.EqualValues(t, 0, len(m.insights.ReceivedContents))

	r := make([]map[string]interface{}, 10)
	json.Unmarshal([]byte(rc), &r)

	assert.EqualValues(t, 1, r[0]["http.duration"])
	assert.EqualValues(t, 10000000, r[0]["http.start.timestamp"])
	assert.EqualValues(t, 11000000, r[0]["http.stop.timestamp"])
	assert.EqualValues(t, PCFHttpStartStop, r[0]["eventType"])
	assert.EqualValues(t, "uri", r[0]["http.uri"])
	assert.EqualValues(t, "GET", r[0]["http.method"])
	assert.EqualValues(t, "Server", r[0]["http.peer.type"])
	assert.EqualValues(t, "Go-http-client/1.1", r[0]["http.user.agent"])

}

func TestDataDump(t *testing.T) {
	m := runNozzleAndMocks()

	rf, err := ioutil.ReadFile("fhout.json")
	if err != nil {
		t.Fatal("fail open dump")
	}
	buf := bytes.NewBuffer(rf)
	jsonDec := json.NewDecoder(buf)

	eventSentCount := make(map[string]int)
	e := &loggregator_v2.Envelope{}
	go func() {
		for i := 0; ; i++ {
			if err := jsonpb.UnmarshalNext(jsonDec, e); err != nil {
				//EOF breaking condition
				m.firehose.PublishBatch()
				break
			}
			m.firehose.AddEvent(*e)
			eventSentCount[getEnvelopeType(e)]++
			// separete in chunks
			if i == 100 {
				i = 0
				m.firehose.PublishBatch()
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()
	eventReceivedCount := make(map[string]int)
ReadingFromInsights:
	for {
		select {
		case rc := <-m.insights.ReceivedContents:
			r := make([]map[string]interface{}, 10)
			json.Unmarshal([]byte(rc), &r)
			for _, rr := range r {
				et := rr["eventType"].(string)
				if et == PCFValueMetric || et == PCFContainerMetric || et == PCFCounterEvent {
					c := rr["metric.samples.count"]
					sc := int(c.(float64))
					eventReceivedCount[et] += sc
				} else {
					//rest of the metrics are not aggregated
					eventReceivedCount[et]++
				}
			}
		case <-time.After(2 * time.Second):
			break ReadingFromInsights
		}

	}
	closeNozzleAndMocks(m)

	assert.Equal(t, eventReceivedCount[PCFHttpStartStop], eventSentCount[PCFHttpStartStop])
	assert.Equal(t, eventReceivedCount[PCFCounterEvent], eventSentCount[PCFCounterEvent])

	events := []string{PCFContainerMetric, PCFValueMetric}
	for _, e := range events {
		assert.GreaterOrEqual(t, eventReceivedCount[e], eventSentCount[e])
	}
}

func TestFirehoseConnectionFail(t *testing.T) {
	m := runNozzleAndMocks()
	defer closeNozzleAndMocks(m)
	//stop firehose to force connection fail and error rising on the nozzle.
	m.firehose.Stop()

	r := bufio.NewReader(m.nozzleStdOut)
	connectionFails := false
	for start := time.Now(); time.Since(start) < time.Second; {
		line, _, _ := r.ReadLine()
		if strings.Contains(string(line), "client connection attempts exceeded max retries -- giving up") {
			connectionFails = true
			break
		}
	}
	assert.True(t, connectionFails)
}

func readInsights(t *testing.T, m *apiMocks) string {
	select {
	case rc := <-m.insights.ReceivedContents:
		return rc
	case <-time.After(10 * time.Second):
		t.Fatal("Expected data from insights.ReceivedContents")
	}
	return ""
}

func getEnvelopeType(e *loggregator_v2.Envelope) string {
	et := reflect.TypeOf(e.Message).String()
	if et == "*loggregator_v2.Envelope_Gauge" {
		if isContainerMetric(e) {
			et = PCFContainerMetric
		} else {
			et = PCFValueMetric
		}
	}
	if strings.Contains(et, "Envelope_Timer") {
		et = PCFHttpStartStop
	}
	if strings.Contains(et, "Envelope_Log") {
		et = PCFLogMessage
	}
	if strings.Contains(et, "Envelope_Counter") {
		et = PCFCounterEvent
	}
	return et
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
