// +build integration

package helpers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

type MockUAAC struct {
	Server      *httptest.Server
	tokenString string
}

func NewMockUAA(tokenType string, accessToken string) *MockUAAC {
	return &MockUAAC{
		tokenString: fmt.Sprintf(`{"token_type": "%s","access_token": "%s"}`, tokenType, accessToken),
	}
}

func (mU *MockUAAC) Start() {
	mU.Server = httptest.NewUnstartedServer(mU)
	mU.Server.Start()
}

func (mU *MockUAAC) Stop() {
	mU.Server.Close()
}

func (mU *MockUAAC) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	rw.Write([]byte(mU.tokenString))
}
