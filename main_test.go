package main_test

import (
	. "github.com/anynines/cf_service_jumper_cli_plugin"
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

})
