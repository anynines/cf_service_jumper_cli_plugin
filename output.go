package main

import (
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

type ConnectionPrinter interface {
	SampleCallOutput(localListenAddress string) string
}

func SelectConnectionPrinter(credentials map[string]string) ConnectionPrinter {
	if strings.HasPrefix(credentials["uri"], "mongodb://") {
		return MongodbConnectionPrinter{Credentials: credentials}
	} else if strings.HasPrefix(credentials["uri"], "postgres://") {
		return PostgresConnectionPrinter{Credentials: credentials}
	}

	return DefaultConnectionPrinter{Credentials: credentials}
}

type DefaultConnectionPrinter struct {
	Credentials map[string]string
}

func (d DefaultConnectionPrinter) SampleCallOutput(localListenAddress string) string {
	return "No sample call to connect to service on " + localListenAddress + " available."
}

type MongodbConnectionPrinter struct {
	Credentials map[string]string
}

func (d MongodbConnectionPrinter) SampleCallOutput(localListenAddress string) string {
	return "mongo " + localListenAddress + "/" + d.Credentials["default_database"] + " -u " + d.Credentials["username"] + " -p " + d.Credentials["password"]
}

type PostgresConnectionPrinter struct {
	Credentials map[string]string
}

func (d PostgresConnectionPrinter) SampleCallOutput(localListenAddress string) string {
	host := strings.Split(localListenAddress, ":")[0]
	port := strings.Split(localListenAddress, ":")[1]

	return "PGPASSWORD=" + d.Credentials["password"] + " psql -h " + host + " " + "-U " + d.Credentials["username"] + " -p " + port + " " + d.Credentials["name"]
}
