package main

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

func OutputForwardDataSets(forwardDataSetCollection []ForwardDataSet) {
	data := make([][]string, len(forwardDataSetCollection))
	for index, forwardDataSet := range forwardDataSetCollection {
		info := []string{
			forwardDataSet.ID,
			forwardDataSet.Credentials.Credentials.Uri,
			forwardDataSet.Credentials.Credentials.Username,
			forwardDataSet.Credentials.Credentials.Password,
			forwardDataSet.Credentials.Credentials.DefaultDatabase,
			forwardDataSet.Credentials.Credentials.Database,
			forwardDataSet.SharedSecret,
		}

		data[index] = info
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Uri", "Username", "Password", "Database", "DefaultDatabase", "SharedSecret"})
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}
