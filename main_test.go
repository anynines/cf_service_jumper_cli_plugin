package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/a9hcp/cf_service_jumper_cli_plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("main", func() {
	Describe("ArgsExtractServiceInstanceName", func() {
		It("errors if less than 2 args", func() {
			_, err := ArgsExtractServiceInstanceName([]string{"arg0"})
			Expect(err).To(Equal(ErrMissingServiceInstanceArg))
		})

		It("works with 2 args", func() {
			instanceName, err := ArgsExtractServiceInstanceName([]string{"arg0", "arg1"})
			Expect(err).To(BeNil())
			Expect(instanceName).To(Equal("arg1"))
		})
	})

	Describe("ArgsExtractConnectionId", func() {
		It("errors if less than 2 args", func() {
			_, err := ArgsExtractConnectionId([]string{"arg0", "arg1"})
			Expect(err).To(Equal(ErrMissingConnectionId))
		})

		It("works with 2 args", func() {
			connectionId, err := ArgsExtractConnectionId([]string{"arg0", "arg1", "arg2"})
			Expect(err).To(BeNil())
			Expect(connectionId).To(Equal("arg2"))
		})
	})

	Describe("FetchCfServiceJumperApiEndpoint", func() {
		It("returns service jumper endpoint", func() {
			fakeEndpointServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					panic(fmt.Sprintf("fake endpoint server: verb GET expected", r.Method))
				}

				jsonStr := `{ "name": "Anynines", "custom": { "service_jumper_endpoint": "https://service-jumper.de.a9sservice.eu" } }`
				fmt.Fprintln(w, jsonStr)
			}))

			sjEndpoint, err := FetchCfServiceJumperApiEndpoint(fakeEndpointServer.URL)
			Expect(err).To(BeNil())
			Expect(sjEndpoint).To(Equal("https://service-jumper.de.a9sservice.eu"))
		})

		It("errors if json broken", func() {
			fakeEndpointServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jsonStr := `"name": "Anynines"`
				fmt.Fprintln(w, jsonStr)
			}))

			_, err := FetchCfServiceJumperApiEndpoint(fakeEndpointServer.URL)
			Expect(err).ToNot(BeNil())
		})

		It("errors if cf service humper endpoint not present", func() {
			fakeEndpointServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jsonStr := `{ "name": "Anynines", "custom": { } }`
				fmt.Fprintln(w, jsonStr)
			}))

			_, err := FetchCfServiceJumperApiEndpoint(fakeEndpointServer.URL)
			Expect(err).To(Equal(ErrCfServiceJumperEndpointNotPresent))
		})
	})

	Describe("create forward", func() {
		It("returns ForwardDataSet", func() {
			fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					panic(fmt.Sprintf("fake server: verb POST expected", r.Method))
				}

				jsonStr := `{ "hosts": ["10.100.0.60:27017", "10.100.0.61:27017"], "sb_credentials": "{}", "binding_id": "1234" }`
				fmt.Fprintln(w, jsonStr)
			}))
			p := CfServiceJumperPlugin{
				CfServiceJumperApiEndpoint: fakeServer.URL,
			}

			forwardDataSet, err := p.CreateForward("serviceGuid")
			Expect(err).To(BeNil())
			expectedForwardDataSet := ForwardDataSet{
				Hosts:         []string{"10.100.0.60:27017", "10.100.0.61:27017"},
				SbCredentials: "{}",
				BindingId:     "1234",
			}
			Expect(forwardDataSet).To(Equal(expectedForwardDataSet))
		})

		It("errors if json broken", func() {
			fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jsonStr := `{ "hosts": {} }`
				fmt.Fprintln(w, jsonStr)
			}))
			p := CfServiceJumperPlugin{
				CfServiceJumperApiEndpoint: fakeServer.URL,
			}

			_, err := p.CreateForward("serviceGuid")
			Expect(err).ToNot(BeNil())
		})
	})
})
