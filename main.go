// New Relic Firehse Nozzle
package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	//"net"
	"net/url"
	"net/http"
	"strings"
	"strconv"
	"time"
	"log"
	"os"
    "bytes"
	"errors"

	"runtime"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/kelseyhightower/envconfig"

	// "github.com/cf-platform-eng/api"
	"github.com/cf-platform-eng/config"
	"github.com/cf-platform-eng/uaa"
)

const (
	nrLogDataset     = "Log Messages"
	nrErrorDataset   = "Error Messages"
	nrMetricsDataset = "Metric Messages"

	insightsMaxEvents = 500 // size of insights json insert packet
)

type NewRelicConfig struct {
	INSIGHTS_BASE_URL				string
	INSIGHTS_RPM_ID					string
	INSIGHTS_INSERT_KEY				string

	// APIURL							string
	// UAAURL							string
	// SkipSSL							bool
	// Username						string
	// Password						string
	// TrafficControllerURL			string
	// FirehoseSubscriptionID			string
	// SelectedEvents []events.Envelope_EventType
}

type NREventType map[string]interface{}

type PcfCounters struct {
	valueMetricEvents 		uint64
	counterEvents 			uint64
	containerEvents 		uint64
	httpStartStopEvents 	uint64
	logMessageEvents 		uint64
	errors 					uint64
}

var NREventsMap = make([]NREventType, 0)
var ee uint64
var pcfCounters PcfCounters
var mem runtime.MemStats
var pcfInstanceIp string
var pcfDomain string
func main() {
	fmt.Println("hello world!")

	logger := log.New(os.Stdout, ">>> ", 0)

	pcfConfig, err := config.Parse()
	if err != nil {
		panic(err)
	}
	logger.Printf("pcfConfig: %v\n", pcfConfig)


	nrConfig := NewRelicConfig{}
	if err := envconfig.Process("newrelic", &nrConfig); err != nil {
		panic(err)
	}
	logger.Printf("nrConfig: %v\n", nrConfig)

	// ###########################################################################

	url := fmt.Sprintf("%s/accounts/%s/events", nrConfig.INSIGHTS_BASE_URL, nrConfig.INSIGHTS_RPM_ID)
	insertKey := nrConfig.INSIGHTS_INSERT_KEY
	logger.Printf("insights url: %v\n", url)
	logger.Printf("insertkey: %v\n", insertKey)
	logger.Printf("pcfConfig.SkipSSL: %v\n", pcfConfig.SkipSSL)
	//logger.Printf("pcfConfig.APIURL: %v\n", pcfConfig.APIURL)
	logger.Printf("pcfConfig.UAAURL: %v\n", pcfConfig.UAAURL)
	logger.Printf("pcfConfig.Username: %v\n", pcfConfig.Username)
	logger.Printf("pcfConfig.Password: %v\n", pcfConfig.Password)
	pcfInstanceIp = os.Getenv("CF_INSTANCE_IP")
	logger.Printf("CF_INSTANCE_IP: %v\n", pcfInstanceIp)
	pcfDomain = strings.SplitN(parseUrl(pcfConfig.UAAURL), ".", 2)[1]
	logger.Printf("PCF Domain: %v\n", pcfDomain)

	// authenticate client
	var token, trafficControllerURL string

	if pcfConfig.UAAURL != "" {
		logger.Printf("Fetching auth token via UAA: %v\n", pcfConfig.UAAURL)

		trafficControllerURL = pcfConfig.TrafficControllerURL
		if trafficControllerURL == "" {
			logger.Fatal(errors.New("NOZZLE_TRAFFIC_CONTROLLER_URL is required when authenticating via UAA"))
		}

		fetcher := uaa.NewUAATokenFetcher(pcfConfig.UAAURL, pcfConfig.Username, pcfConfig.Password, pcfConfig.SkipSSL)
		token, err = fetcher.FetchAuthToken()
		if err != nil {
			logger.Fatal("Unable to fetch token via UAA", err)
		}
	} else {
		logger.Fatal(errors.New("Either of NOZZLE_API_URL or NOZZLE_UAA_URL are required"))
	}

	logger.Printf("token: %v\n", token)

	// consume PCF logs
	consumer := consumer.New(pcfConfig.TrafficControllerURL, &tls.Config{
		InsecureSkipVerify: pcfConfig.SkipSSL,
	}, nil)

	evs, errors := consumer.Firehose(pcfConfig.FirehoseSubscriptionID, token)

	fmt.Printf("events and errors are %+v and %+v\n", evs, errors)

	i := 0
	fmt.Printf("about to print events\n")
	for {
		i++
		//nrEvent := make(map[string]interface{})
		nrEvent := make(NREventType)

		select {
		case ev := <-evs:
			// fmt.Printf("event %d: %v\n", i, ev)
			if err := transformEvent(ev, nrEvent); err != nil {
				panic(err)
			}

			pushToInsights(nrEvent, url, insertKey)

		case ev := <-errors:
			fmt.Printf("%d: ev is %+s\n", i, ev.Error())
			nrEvent["error"] = ev.Error()
		}
	}
}

func pushToInsights(nrEvent map[string]interface{}, url string, insertKey string) {

//ee = ee + 1
//checkMem(ee)
	NREventsMap = append(NREventsMap, nrEvent)
//checkMem(ee)
	// fmt.Println(nrEvent)

	if(len(NREventsMap) >= insightsMaxEvents) {
		jsonStr, err := json.Marshal(NREventsMap)
		if err != nil {
			fmt.Println("error:", err)
		}
		// fmt.Println("jsonstr:", string(jsonStr)) // TEMP
		fmt.Printf("Value Metrics: %d, Counter Events: %d, Container Events: %d, Http StartStop Events: %d, Log Messages: %d, Errors: %d\n",
			pcfCounters.valueMetricEvents, pcfCounters.counterEvents, pcfCounters.containerEvents, 
			pcfCounters.httpStartStopEvents, pcfCounters.logMessageEvents, pcfCounters.errors)


	    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	    req.Header.Set("X-Insert-Key", insertKey)
	    req.Header.Set("Content-Type", "application/json")

	    client := &http.Client{}
	    resp, err := client.Do(req)
	    if err != nil {
	        panic(err)
	    }
	    defer resp.Body.Close()

//checkMem(ee)
		NREventsMap = nil
	    NREventsMap = make([]NREventType, 0)
//checkMem(ee)
	}

}

func transformEvent(cfEvent *events.Envelope, nrEvent map[string]interface{}) error {
	// add generic fields
	nrEvent["origin"] = cfEvent.GetOrigin()
	nrEvent["eventType"] = "PcfFirehoseEvent"
	nrEvent["FirehoseEventType"] = cfEvent.GetEventType().String()
	if pcfInstanceIp > "" {
		nrEvent["pcfInstanceIp"] = pcfInstanceIp
	}
	if pcfDomain > "" {
		nrEvent["pcfDomain"] = pcfDomain
	}
	nrEvent["deployment"] = cfEvent.GetDeployment()
	nrEvent["job"] = cfEvent.GetJob()
	nrEvent["index"] = cfEvent.GetIndex()
	nrEvent["ip"] = cfEvent.GetIp()
	nrEvent["timestamp"] = cfEvent.GetTimestamp() / 1000000 // nanoseconds -> milliseconds
	for name, val := range cfEvent.Tags {
		nrEvent["tag_"+name] = val
	}

	// get in to event type specific stuff
	switch *cfEvent.EventType {
		case events.Envelope_HttpStartStop:
			pcfCounters.httpStartStopEvents++
			nrEvent["DatasetName"] = nrLogDataset
			transformHttpStartStopEvent(cfEvent, nrEvent)

		case events.Envelope_LogMessage:
			pcfCounters.logMessageEvents++
			nrEvent["DatasetName"] = nrLogDataset
			transformLogMessage(cfEvent, nrEvent)

		case events.Envelope_ContainerMetric:
			pcfCounters.containerEvents++
			nrEvent["DatasetName"] = nrMetricsDataset
			transformContainerMetric(cfEvent, nrEvent)
			//nrEvent["containerMetric"] = cfEvent.GetContainerMetric().String()

		case events.Envelope_CounterEvent:
			pcfCounters.counterEvents++
			nrEvent["DatasetName"] = nrMetricsDataset
			transformCounterEvent(cfEvent, nrEvent)
			// nrEvent["counterEvent"] = cfEvent.GetCounterEvent().String()

		case events.Envelope_ValueMetric:
			pcfCounters.valueMetricEvents++
			nrEvent["DatasetName"] = nrMetricsDataset
			transformValueMetric(cfEvent, nrEvent)
			//nrEvent["valueMetric"] = cfEvent.GetValueMetric().String()

		case events.Envelope_Error:
			pcfCounters.errors++
			nrEvent["DatasetName"] = nrErrorDataset
			//nrEvent["errorField"] = cfEvent.GetError().String()
	}
	return nil
}

// process ValueMetric events to new relic event format
func transformValueMetric(cfEvent *events.Envelope, nrEvent map[string]interface{}) {
	// event: origin:"DopplerServer" eventType:ValueMetric timestamp:1497038365914920486 deployment:"cf" job:"doppler" index:"ca858dc5-2a09-465a-831d-c31fa5fb8802" ip:"192.168.16.26" valueMetric:<name:"messageRouter.numberOfFirehoseSinks" value:1 unit:"sinks" > 
	vm := cfEvent.ValueMetric
	prefix := "valueMetric"
	if vm.Name != nil {
		nrEvent[prefix+"Name"] = vm.GetName()
	}
	if vm.Value != nil {
		nrEvent[prefix+"Value"] = vm.GetValue()
	}
	if vm.Unit != nil {
		nrEvent[prefix+"Unit"] = vm.GetUnit()
	}
}

// process CounterEvent events to new relic event format
func transformCounterEvent(cfEvent *events.Envelope, nrEvent map[string]interface{}) {
	// event: origin:"DopplerServer" eventType:CounterEvent timestamp:1497038366107650076 deployment:"cf" job:"doppler" index:"ca858dc5-2a09-465a-831d-c31fa5fb8802" ip:"192.168.16.26" counterEvent:<name:"udpListener.receivedByteCount" delta:152887 total:25671098577 > 
	ce := cfEvent.CounterEvent
	prefix := "counterEvent"
	if ce.Name != nil {
		nrEvent[prefix+"Name"] = ce.GetName()
	}
	if ce.Delta != nil {
		nrEvent[prefix+"Delta"] = ce.GetDelta()
	}
	if ce.Total != nil {
		nrEvent[prefix+"Total"] = ce.GetTotal()
	}
}

// process ContainerMetric events to new relic event format
func transformContainerMetric(cfEvent *events.Envelope, nrEvent map[string]interface{}) {
	// event: origin:"rep" eventType:ContainerMetric timestamp:1497038370673051301 deployment:"cf" job:"diego_cell" index:"302e37ef-f847-4b96-bdff-5c6e4f0d1259" ip:"192.168.16.23" containerMetric:<applicationId:"a0bc8fd4-8980-4e0e-81b3-7f9709ff407e" instanceIndex:0 cpuPercentage:0.07382914424191898 memoryBytes:359899136 diskBytes:142286848 memoryBytesQuota:536870912 diskBytesQuota:1073741824 > 
	cm := cfEvent.ContainerMetric
	prefix := "containerMetric"
	if cm.ApplicationId != nil {
		nrEvent[prefix+"ApplicationId"] = cm.GetApplicationId()
	}
	if cm.InstanceIndex != nil {
		nrEvent[prefix+"InstanceIndex"] = cm.GetInstanceIndex()
	}
	if cm.CpuPercentage != nil {
		nrEvent[prefix+"CpuPercentage"] = cm.GetCpuPercentage()
	}
	if cm.MemoryBytes != nil {
		nrEvent[prefix+"MemoryBytes"] = cm.GetMemoryBytes()
	}
	if cm.DiskBytes != nil {
		nrEvent[prefix+"DiskBytes"] = cm.GetDiskBytes()
	}
	if cm.MemoryBytesQuota != nil {
		nrEvent[prefix+"MemoryBytesQuota"] = cm.GetMemoryBytesQuota()
	}
	if cm.DiskBytesQuota != nil {
		nrEvent[prefix+"DiskBytesQuota"] = cm.GetDiskBytesQuota()
	}
}

// process application log events to new relic event format
func transformLogMessage(cfEvent *events.Envelope, nrEvent map[string]interface{}) {
	// event: origin:"rep" eventType:LogMessage timestamp:1497038366041617814 deployment:"cf" job:"diego_cell" index:"0f4dc7bd-c941-42bf-a835-7c29445ddf8b" ip:"192.168.16.24" logMessage:<message:"[{\"DatasetName\":\"Metric Messages\",\"FirehoseEventType\":\"CounterEvent\",\"ceDelta\":166908,\"ceName\":\"dropsondeListener.receivedByteCount\",\"ceTotal\":25664179951,\"deployment\":\"cf\",\"eventType\":\"FirehoseEventTest\",\"index\":\"ca858dc5-2a09-465a-831d-c31fa5fb8802\",\"ip\":\"192.168.16.26\",\"job\":\"doppler\",\"origin\":\"DopplerServer\",\"timestamp\":1497038161107}]" message_type:OUT timestamp:1497038366041615818 app_id:"f22aac70-c5a9-47a9-b74c-355dd99abbe2" source_type:"APP/PROC/WEB" source_instance:"0" > 
	message := cfEvent.LogMessage
	prefix := "log"
	if message.Message != nil {
		msgContent := message.GetMessage()
		nrEvent[prefix+"Message"] = string(msgContent)
		parsedContent := make(map[string]interface{})
		if err := json.Unmarshal(msgContent, &parsedContent); err == nil {
			for k, v := range parsedContent {
				nrEvent[prefix+"Message"+k] = v
			}
		}
	}
	if message.MessageType != nil {
		nrEvent[prefix+"MessageType"] = message.GetMessageType().String()
	}
	if message.Timestamp != nil {
		nrEvent[prefix+"Timestamp"] = time.Unix(0, message.GetTimestamp())
	}
	if message.AppId != nil {
		nrEvent[prefix+"AppId"] = message.GetAppId()
	}
	if message.SourceType != nil {
		nrEvent[prefix+"SourceType"] = message.GetSourceType()
	}
	if message.SourceInstance != nil {
		nrEvent[prefix+"SourceInstance"] = message.GetSourceInstance()
	}
	// nrEvent.Add(message)
}

// process http start/stop events to new relic event format
func transformHttpStartStopEvent(cfEvent *events.Envelope, nrEvent map[string]interface{}) {
	// event: origin:"gorouter" eventType:HttpStartStop timestamp:1497038373295178447 deployment:"cf" job:"router" index:"1276dbaa-f5a4-4c48-bcbe-d06ff0dba58d" ip:"192.168.16.16" httpStartStop:<startTimestamp:1497038373206213992 stopTimestamp:1497038373295152451 requestId:<low:7513566559519661218 high:8828490834936076361 > peerType:Client method:GET uri:"http://api.sys.pie-22.cfplatformeng.com/v2/syslog_drain_urls" remoteAddress:"130.211.2.63:61939" userAgent:"Go-http-client/1.1" statusCode:200 contentLength:42 instanceId:"89a53ed9-cf20-404b-5728-33a19c1e13ef" forwarded:"104.197.98.14" forwarded:"35.186.215.103" forwarded:"130.211.2.63" > 
	httpEvent := cfEvent.HttpStartStop
	prefix := "http"
	start := time.Unix(0, httpEvent.GetStartTimestamp())
	end := time.Unix(0, httpEvent.GetStopTimestamp())
	duration := float64(end.Sub(start)) / float64(time.Millisecond)
	nrEvent[prefix+"StartTimestamp"] = start
	nrEvent[prefix+"StopTimestamp"] = end
	nrEvent[prefix+"DurationMs"] = duration
	if httpEvent.RequestId != nil {
		nrEvent[prefix+"RequestId"] = httpEvent.GetRequestId().String()
	}
	if httpEvent.PeerType != nil {
		nrEvent[prefix+"PeerType"] = httpEvent.GetPeerType().String()
	}
	if httpEvent.Method != nil {
		nrEvent[prefix+"Method"] = httpEvent.GetMethod().String()
	}
	if httpEvent.Uri != nil {
		nrEvent[prefix+"Uri"] = httpEvent.GetUri()
	}
	if httpEvent.RemoteAddress != nil {
		nrEvent[prefix+"RemoteAddress"] = httpEvent.GetRemoteAddress()
	}
	if httpEvent.UserAgent != nil {
		nrEvent[prefix+"UserAgent"] = httpEvent.GetUserAgent()
	}
	if httpEvent.StatusCode != nil {
		nrEvent[prefix+"StatusCode"] = httpEvent.GetStatusCode()
	}
	if httpEvent.ContentLength != nil {
		nrEvent[prefix+"ContentLength"] = httpEvent.GetContentLength()
	}
	if httpEvent.ApplicationId != nil {
		nrEvent[prefix+"ApplicationId"] = httpEvent.GetApplicationId()
	}
	if httpEvent.InstanceIndex != nil {
		nrEvent[prefix+"InstanceIndex"] = httpEvent.GetInstanceIndex()
	}
	if httpEvent.InstanceId != nil {
		nrEvent[prefix+"InstanceId"] = httpEvent.GetInstanceId()
	}
	for i, forwardedIp := range httpEvent.Forwarded {
		index := strconv.Itoa(i)
		nrEvent[prefix+"Forwarded"+index] = forwardedIp
	}
}

func checkMem(seq int) {
	runtime.ReadMemStats(&mem)
	log.Println(seq, ": allocated: ",mem.Alloc, " - total allocated: ", mem.TotalAlloc, " - heap allocated: ", mem.HeapAlloc, " - heap sys: ", mem.HeapSys)
}

func parseUrl(uaaUrl string) string {

  u, err := url.Parse(uaaUrl)
  if err != nil {
      panic(err)
  }

  return u.Host
}

