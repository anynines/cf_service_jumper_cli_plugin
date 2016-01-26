package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func OutputForwardDataSets(forwardDataSetCollection []ForwardDataSet) {
	data := make([][]string, len(forwardDataSetCollection))
	for index, forwardDataSet := range forwardDataSetCollection {
		info := []string{
			strconv.Itoa(forwardDataSet.ID),
		}

		data[index] = info
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func OutputSampleCmds(sampleCmds []string) {
	if len(sampleCmds) < 1 {
		return
	}

	fmt.Printf("\nYou can connect to the service using the following command(s):\n")
	for _, sampleOutput := range sampleCmds {
		fmt.Println(sampleOutput)
	}

}

type ConnectionPrinter interface {
	SampleCallOutput(localListenAddress string) string
}

type ConnectionPrinterCredentials map[string]string

func SelectConnectionPrinter(credentials map[string]string) ConnectionPrinter {
	if strings.HasPrefix(credentials["uri"], "mongodb://") {
		return MongodbConnectionPrinter{ConnectionPrinterCredentials: credentials}
	} else if strings.HasPrefix(credentials["uri"], "postgres://") {
		return PostgresConnectionPrinter{ConnectionPrinterCredentials: credentials}
	} else if strings.HasPrefix(credentials["uri"], "amqp://") {
		return RabbitMQConnectionPrinter{ConnectionPrinterCredentials: credentials}
	}

	return DefaultConnectionPrinter{ConnectionPrinterCredentials: credentials}
}

type DefaultConnectionPrinter struct {
	ConnectionPrinterCredentials
}

func (d DefaultConnectionPrinter) SampleCallOutput(localListenAddress string) string {
	return "No sample call to connect to service on " + localListenAddress + " available."
}

type MongodbConnectionPrinter struct {
	ConnectionPrinterCredentials
}

func (d MongodbConnectionPrinter) SampleCallOutput(localListenAddress string) string {
	return "mongo " + localListenAddress + "/" + d.ConnectionPrinterCredentials["default_database"] + " -u " + d.ConnectionPrinterCredentials["username"] + " -p " + d.ConnectionPrinterCredentials["password"]
}

type PostgresConnectionPrinter struct {
	ConnectionPrinterCredentials
}

func (d PostgresConnectionPrinter) SampleCallOutput(localListenAddress string) string {
	host := strings.Split(localListenAddress, ":")[0]
	port := strings.Split(localListenAddress, ":")[1]

	return "PGPASSWORD=" + d.ConnectionPrinterCredentials["password"] + " psql -h " + host + " " + "-U " + d.ConnectionPrinterCredentials["username"] + " -p " + port + " " + d.ConnectionPrinterCredentials["name"]
}

type RabbitMQConnectionPrinter struct {
	ConnectionPrinterCredentials
}

func (d RabbitMQConnectionPrinter) SampleCallOutput(localListenAddress string) string {
	return "" // do nothing since there is no cmd line tool we might use at the moment
}
