package main_test

import (
	. "github.com/a9hcp/cf_service_jumper_cli_plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgresConnectionPrinter", func() {
	Describe("SampleCallOutput", func() {
		It("returns cmd", func() {
			cp := PostgresConnectionPrinter{
				Credentials: map[string]string{
					"username": "the_username",
					"password": "the_password",
					"name":     "databasename",
				},
			}

			Expect(cp.SampleCallOutput("localhost:12345")).To(Equal("PGPASSWORD=the_password psql -h localhost -U the_username -p 12345 databasename"))
		})
	})
})

var _ = Describe("MongodbConnectionPrinter", func() {
	Describe("SampleCallOutput", func() {
		It("returns cmd", func() {
			cp := MongodbConnectionPrinter{
				Credentials: map[string]string{
					"username":         "the_username",
					"password":         "the_password",
					"default_database": "databasename",
				},
			}

			Expect(cp.SampleCallOutput("localhost:56789")).To(Equal("mongo localhost:56789/databasename -u the_username -p the_password"))
		})
	})
})
