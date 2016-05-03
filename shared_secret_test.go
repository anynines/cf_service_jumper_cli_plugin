package main_test

import (
	. "github.com/anynines/cf_service_jumper_cli_plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetIdentityAndKey", func() {
	Describe("incorrect shared secret", func() {
		It("errors if not valid", func() {
			var err error

			_, _, err = GetIdentityAndKey("")
			Expect(err).ToNot(BeNil())

			_, _, err = GetIdentityAndKey("identity:")
			Expect(err).ToNot(BeNil())

			_, _, err = GetIdentityAndKey(":key")
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("correct shared secret", func() {
		It("returns identity and key", func() {
			identity, key, err := GetIdentityAndKey("identity:key")
			Expect(err).To(BeNil())
			Expect(identity).To(Equal("identity"))
			Expect(key).To(Equal("key"))
		})
	})
})
