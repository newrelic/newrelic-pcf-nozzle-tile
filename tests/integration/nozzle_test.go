// +build integration

package test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	loggregator "code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	mocks "github.com/newrelic/newrelic-pcf-nozzle-tile/tests/integration/helpers"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
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

func TestMain(m *testing.M) {
	port = 8080
	os.Exit(m.Run())
}

type apiMocks struct {
	uaa      *mocks.MockUAAC
	firehose *mocks.MockFirehose
	cc       *mocks.MockCF
	insights *mocks.MockInsights
	nozzle   *exec.Cmd
}

func runNozzleAndMocks() *apiMocks {
	m := &apiMocks{
		uaa:      mocks.NewMockUAA("bearer", "token"),
		cc:       mocks.NewMockCF("bearer", "token"),
		insights: mocks.NewMockInsights("Gzip"),
		nozzle:   exec.Command("../../dist/nr-fh-nozzle"), //TODO add int test to make file to
	}
	m.firehose = mocks.NewMockFirehose(360, "token")
	m.uaa.Start()
	m.firehose.Start()
	m.cc.Start()
	m.insights.Start()

	mutex.Lock()
	defer mutex.Unlock()
	os.Setenv("NRF_CF_API_URL", m.cc.Server.URL)
	os.Setenv("NRF_CF_API_UAA_URL", m.uaa.Server.URL)
	os.Setenv("NRF_CF_CLIENT_ID", "admin")
	os.Setenv("NRF_CF_CLIENT_SECRET", "token")
	os.Setenv("NRF_CF_API_USERNAME", "admin")
	os.Setenv("NRF_CF_API_PASSWORD", "token")
	os.Setenv("NRF_NEWRELIC_INSERT_KEY", "nrkey")
	os.Setenv("NRF_NEWRELIC_ACCOUNT_ID", "00000")
	os.Setenv("NRF_NEWRELIC_DRAIN_INTERVAL", "500ms")
	os.Setenv("NRF_NEWRELIC_ACCOUNT_REGION", "EU")
	os.Setenv("NRF_NEWRELIC_EU_BASE_URL", m.insights.Server.URL)
	os.Setenv("NRF_CF_API_RLPG_URL", m.firehose.Server.URL)
	os.Setenv("NRF_HEALTH_PORT", strconv.Itoa(port))
	port++

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
	t.Parallel()
	m := runNozzleAndMocks()
	for i := float64(1); i < 11; i++ {
		e := loggregator.Envelope{
			SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
			InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
			Message: &loggregator.Envelope_Gauge{
				Gauge: &loggregator.Gauge{
					Metrics: map[string]*loggregator.GaugeValue{
						"name": &loggregator.GaugeValue{
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

func TestLogMessage(t *testing.T) {
	t.Parallel()
	m := runNozzleAndMocks()

	m.firehose.AddEvent(loggregator.Envelope{
		SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		Message: &loggregator.Envelope_Log{
			Log: &loggregator.Log{
				Payload: []byte("logtest"),
				Type:    loggregator.Log_OUT,
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
	t.Parallel()
	m := runNozzleAndMocks()

	m.firehose.AddEvent(loggregator.Envelope{
		SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		Message: &loggregator.Envelope_Gauge{
			Gauge: &loggregator.Gauge{
				Metrics: map[string]*loggregator.GaugeValue{
					"cpu": &loggregator.GaugeValue{
						Unit:  "percent",
						Value: float64(2),
					},
					"memory": &loggregator.GaugeValue{
						Unit:  "bytes",
						Value: float64(10),
					},
					"disk": &loggregator.GaugeValue{
						Unit:  "bytes",
						Value: float64(25),
					},
					"memory_quota": &loggregator.GaugeValue{
						Unit:  "bytes",
						Value: float64(1000),
					},
					"disk_quota": &loggregator.GaugeValue{
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
	t.Parallel()
	m := runNozzleAndMocks()

	m.firehose.AddEvent(loggregator.Envelope{
		SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		Message: &loggregator.Envelope_Counter{
			Counter: &loggregator.Counter{
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
	t.Parallel()
	m := runNozzleAndMocks()

	m.firehose.AddEvent(loggregator.Envelope{
		SourceId:   "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		InstanceId: "c70684e2-4443-4ed5-8dc8-28b7cf7d97ed",
		Tags: map[string]string{
			"uri":        "uri",
			"method":     "GET",
			"peer_type":  "Server",
			"user_agent": "Go-http-client/1.1",
		},
		Message: &loggregator.Envelope_Timer{
			Timer: &loggregator.Timer{
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
	e := &loggregator.Envelope{}
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
			if err := json.Unmarshal([]byte(rc), &r); err != nil {
				break
			}
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

func readInsights(t *testing.T, m *apiMocks) string {
	select {
	case rc := <-m.insights.ReceivedContents:
		return rc
	case <-time.After(10 * time.Second):
		t.Fatal("Expected data from insights.ReceivedContents")
	}
	return ""
}

func getEnvelopeType(e *loggregator.Envelope) string {
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
func isContainerMetric(e *loggregator.Envelope) bool {
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
