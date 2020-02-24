// +build integration

package mocks

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type MockInsights struct {
	Server           *httptest.Server
	ReceivedContents chan string
	TypeCompression  string
}

func NewMockInsights(typeCompression string) *MockInsights {
	return &MockInsights{
		ReceivedContents: make(chan string, 100),
		TypeCompression:  typeCompression,
	}
}

func (mI *MockInsights) Start() {
	mI.Server = httptest.NewUnstartedServer(mI)
	mI.Server.Start()
}

func (mI *MockInsights) Stop() {
	mI.Server.Close()
	close(mI.ReceivedContents)
}

func (mI *MockInsights) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	contents, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(`{"success": true}`))
	mI.ReceivedContents <- mI.decompress(contents)
}

func (mI *MockInsights) decompress(src []byte) string {
	if mI.TypeCompression == "None" {
		return string(src)
	} else if mI.TypeCompression == "Gzip" {
		r, _ := gzip.NewReader(bytes.NewReader(src))
		defer r.Close()
		dst, _ := ioutil.ReadAll(r)
		return string(dst)
	}
	return string(src)
}
