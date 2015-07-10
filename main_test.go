package main_test

import (
	. "github.com/anynines/cf_service_jumper_cli_plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ArgsExtractServiceInstanceName", func() {
	It("errors if less than 2 args", func() {
		_, err := ArgsExtractServiceInstanceName([]string{"arg0"})
		Expect(err).To(Equal(ErrMissingServiceInstanceArg))
	})

	It("works with 2 args", func() {
		_, err := ArgsExtractServiceInstanceName([]string{"arg0", "arg1"})
		Expect(err).To(BeNil())
	})
})
