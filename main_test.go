package main_test

import (
	"fmt"
	. "github.com/anynines/cf_service_jumper_cli_plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
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
			Expect(err).To(Equal(ErrCfServiceJumperEndpointUnmarshal))
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
})
