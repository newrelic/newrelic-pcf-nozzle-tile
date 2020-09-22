// +build integration

package helpers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/golang/protobuf/jsonpb"
)

var marshaler = jsonpb.Marshaler{
	EmitDefaults: true,
}

type MockFirehose struct {
	lock sync.Mutex

	Server          *httptest.Server
	closeConnection chan struct{}
	stopPublishing  chan struct{}
	serveBatch      chan *loggregator_v2.EnvelopeBatch
	events          []*loggregator_v2.Envelope
	TimeTicker      *time.Ticker

	validToken string
}

func NewMockFirehose(seconds int, validToken string) *MockFirehose {
	return &MockFirehose{
		validToken:      "bearer " + validToken,
		closeConnection: make(chan struct{}),
		stopPublishing:  make(chan struct{}),
		serveBatch:      make(chan *loggregator_v2.EnvelopeBatch, 10),
		TimeTicker:      time.NewTicker(time.Second * time.Duration(seconds)),
	}
}

//The firehose receives the data from external sources and push them when the timeticker send a hearthbeat
func (mf *MockFirehose) Start() {
	mf.Server = httptest.NewUnstartedServer(mf)
	mf.Server.Start()
	go func() {
		for {
			select {
			case <-mf.TimeTicker.C:
				mf.PublishBatch()
			case <-mf.stopPublishing:
				return
			}
		}
	}()
}

func (mf *MockFirehose) Stop() {
	close(mf.stopPublishing)
	close(mf.closeConnection)
	mf.Server.Close()
}

func (mf *MockFirehose) AddEvent(event loggregator_v2.Envelope) {
	mf.lock.Lock()
	defer mf.lock.Unlock()
	mf.events = append(mf.events, &event)
}

func (mf *MockFirehose) PublishBatch() {
	mf.lock.Lock()
	defer mf.lock.Unlock()

	envelopeBatch := &loggregator_v2.EnvelopeBatch{Batch: mf.events}
	mf.serveBatch <- envelopeBatch
	mf.events = []*loggregator_v2.Envelope{}
}

func (mf *MockFirehose) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	setHeaders(rw)

	if mf.isTokenInvalid(r.Header.Get("Authorization")) {
		rw.WriteHeader(403)
		return
	}
	for {
		select {
		case b := <-mf.serveBatch:
			mf.sendData(rw, b)
		case <-mf.closeConnection:
			return
		}
	}
}

func (mf *MockFirehose) sendData(rw http.ResponseWriter, b *loggregator_v2.EnvelopeBatch) {
	mf.lock.Lock()
	defer mf.lock.Unlock()

	marshalledBatch, _ := marshaler.MarshalToString(b)
	_, _ = fmt.Fprintf(rw, "data: %s\n\n", marshalledBatch)
	flusher, _ := rw.(http.Flusher)
	flusher.Flush()
}

func (mf *MockFirehose) isTokenInvalid(token string) bool {
	mf.lock.Lock()
	defer mf.lock.Unlock()

	if token != mf.validToken {
		log.Printf("The nozzle connected to the firehose mock API making use of a wrong token: %s - %s", token, mf.validToken)
		return true
	}
	return false
}

func setHeaders(rw http.ResponseWriter) {
	// The firehoseMock keeps alive the connection streaming the data as they become available
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
}
