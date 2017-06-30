package api_test

import (
	. "github.com/cf-platform-eng/firehose-nozzle/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"net/http"
	"fmt"
)

var _ = Describe("Api", func() {
	var testServer *httptest.Server
	var capturedRequests []*http.Request
	var responses []string
	var v2InfoResponse string

	BeforeEach(func() {
		capturedRequests = []*http.Request{}
		responses = []string{}

		testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			capturedRequests = append(capturedRequests, request)
			response := responses[0]
			responses = responses[1:]
			jsonData := []byte(response)
			writer.Write(jsonData)
		}))
		v2InfoResponse = fmt.Sprintf(`{
			    "authorization_endpoint": "%v",
			    "token_endpoint": "%v",
			    "min_cli_version": "6.7.0",
			    "min_recommended_cli_version": "6.11.2",
			    "api_version": "2.65.0",
			    "routing_endpoint": "https://routing.example.com",
			    "logging_endpoint": "wss://loggregator.example.com",
			    "doppler_logging_endpoint": "wss://doppler.example.com"
			}`, testServer.URL, testServer.URL)
	})

	AfterEach(func() {
		testServer.Close()
	})

	It("constructs client", func() {
		responses = []string{v2InfoResponse, `{}`}

		apiClient, err := NewAPIClient(testServer.URL, "admin", "password", true)
		Expect(err).To(BeNil())
		Expect(apiClient).NotTo(BeNil())

		Expect(capturedRequests).To(HaveLen(2))
		Expect(capturedRequests[0].URL.Path).To(Equal("/v2/info"))
		Expect(capturedRequests[1].URL.Path).To(Equal("/oauth/token"))
	})

	It("returns doppler endpoint", func() {
		responses = []string{
			v2InfoResponse,
			`{}`,
		}

		apiClient, _ := NewAPIClient(testServer.URL, "admin", "password", true)
		Expect(apiClient.FetchTrafficControllerURL()).To(Equal("wss://doppler.example.com"))
	})

})
