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

	Describe("ArgsExtractConnectionID", func() {
		It("errors if less than 2 args", func() {
			_, err := ArgsExtractConnectionID([]string{"arg0", "arg1"})
			Expect(err).To(Equal(ErrMissingConnectionID))
		})

		It("works with 2 args", func() {
			connectionId, err := ArgsExtractConnectionID([]string{"arg0", "arg1", "arg2"})
			Expect(err).To(BeNil())
			Expect(connectionId).To(Equal("arg2"))
		})
	})

	Describe("FetchCfServiceJumperAPIEndpoint", func() {
		It("returns service jumper endpoint", func() {
			fakeEndpointServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					panic(fmt.Sprintf("fake endpoint server: verb GET expected", r.Method))
				}

				jsonStr := `{ "name": "Anynines", "custom": { "service_jumper_endpoint": "https://service-jumper.de.a9sservice.eu" } }`
				fmt.Fprintln(w, jsonStr)
			}))

			sjEndpoint, err := FetchCfServiceJumperAPIEndpoint(fakeEndpointServer.URL)
			Expect(err).To(BeNil())
			Expect(sjEndpoint).To(Equal("https://service-jumper.de.a9sservice.eu"))
		})

		It("errors if json broken", func() {
			fakeEndpointServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jsonStr := `"name": "Anynines"`
				fmt.Fprintln(w, jsonStr)
			}))

			_, err := FetchCfServiceJumperAPIEndpoint(fakeEndpointServer.URL)
			Expect(err).ToNot(BeNil())
		})

		It("errors if cf service humper endpoint not present", func() {
			fakeEndpointServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jsonStr := `{ "name": "Anynines", "custom": { } }`
				fmt.Fprintln(w, jsonStr)
			}))

			_, err := FetchCfServiceJumperAPIEndpoint(fakeEndpointServer.URL)
			Expect(err).To(Equal(ErrCfServiceJumperEndpointNotPresent))
		})
	})

	Describe("create forward", func() {
		It("returns ForwardDataSet", func() {
			fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					panic(fmt.Sprintf("fake server: verb POST expected", r.Method))
				}

				jsonStr := `{ "public_uris": ["10.100.0.60:27017", "10.100.0.61:27017"], "credentials": { "credentials": { "username": "the_username", "random_entry": "a_random_entry" } }, "id": "1234" }`
				fmt.Fprintln(w, jsonStr)
			}))
			p := CfServiceJumperPlugin{
				CfServiceJumperAPIEndpoint: fakeServer.URL,
			}

			forwardDataSet, err := p.CreateForward("serviceGuid")
			Expect(err).To(BeNil())
			expectedForwardDataSet := ForwardDataSet{
				Hosts: []string{"10.100.0.60:27017", "10.100.0.61:27017"},
				Credentials: ForwardCredentials{
					ForwardSbCredentials{
						Username: "the_username",
					},
				},
			}
			Expect(forwardDataSet).To(Equal(expectedForwardDataSet))
		})
	})

	Describe("ForwardDataSet", func() {
		Describe("CredentialsMap", func() {
			It("return a human readble string", func() {
				fds := ForwardDataSet{
					Credentials: ForwardCredentials{
						Credentials: ForwardSbCredentials{
							Uri:             "theuri",
							Username:        "username",
							Password:        "password",
							DefaultDatabase: "defaultdatabase",
							Database:        "database",
						},
					},
				}
				credentials := fds.CredentialsMap()
				Expect(credentials["URI"]).To(Equal("theuri"))
				Expect(credentials["Username"]).To(Equal("username"))
				Expect(credentials["Password"]).To(Equal("password"))
				Expect(credentials["Default database"]).To(Equal("defaultdatabase"))
				Expect(credentials["Database"]).To(Equal("database"))
			})
		})
	})
})
