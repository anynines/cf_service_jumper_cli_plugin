package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/parnurzeal/gorequest"

	"github.com/a9hcp/cf_service_jumper_cli_plugin/xtunnel"
)

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, "error: ", err)
		os.Exit(1)
	}
}

var errMissingServiceInstanceArg = errors.New("[ERR] missing SERVICE_INSTANCE")
var errMissingConnectionID = errors.New("[ERR] missing CONNECTION_ID")
var errCfServiceJumperEndpointGetFailed = errors.New("[ERR] Failed to fetch information from Cloud Foundry api endpoint")
var errCfServiceJumperEndpointStatusCodeWrong = errors.New("[ERR] Failed to fetch information from Cloud Foundry api endpoint. HTTP status code != 200")
var errCfServiceJumperEndpointNotPresent = errors.New("[ERR] cf service jumper api endpoint not present/installed")

// ArgsExtractServiceInstanceName extract service instance name from args
func ArgsExtractServiceInstanceName(args []string) (string, error) {
	if len(args) < 2 {
		return "", errMissingServiceInstanceArg
	}

	return args[1], nil
}

// ArgsExtractConnectionID extracts connection ID from args
func ArgsExtractConnectionID(args []string) (string, error) {
	if len(args) < 3 {
		return "", errMissingConnectionID
	}

	return args[2], nil
}

// FetchCfServiceJumperAPIEndpoint fetches the Service Jumper API endpoint
func FetchCfServiceJumperAPIEndpoint(cfAPIEndpoint string) (string, error) {
	url := fmt.Sprintf("%s/v2/info", cfAPIEndpoint)

	request := gorequest.New()
	resp, body, errs := request.Get(url).End()

	if len(errs) > 0 {
		return "", errCfServiceJumperEndpointGetFailed
	}
	if resp.StatusCode != http.StatusOK {
		return "", errCfServiceJumperEndpointStatusCodeWrong
	}

	type CfInfo struct {
		Custom map[string]string `json:"custom"`
	}
	var cfInfo CfInfo
	err := json.Unmarshal([]byte(body), &cfInfo)
	if err != nil {
		return "", fmt.Errorf("[ERR] cf service jumper api endpoint unmarshal failed. %s", err)
	}

	serviceJumperEndpoint := cfInfo.Custom["service_jumper_endpoint"]
	if len(serviceJumperEndpoint) < 1 {
		return "", errCfServiceJumperEndpointNotPresent
	}

	return serviceJumperEndpoint, nil
}

// CfServiceJumperPlugin This is the struct implementing the interface defined by the core CLI. It can
// be found at  "https://github.com/cloudfoundry/cli/blob/master/plugin/plugin.go"
type CfServiceJumperPlugin struct {
	CfServiceJumperAccessToken string
	CfServiceJumperAPIEndpoint string
}

// FetchServiceGUID fetch service GUID by service name
func (c *CfServiceJumperPlugin) FetchServiceGUID(cliConnection plugin.CliConnection, serviceInstanceName string) (string, error) {
	cmdOutput, err := cliConnection.CliCommandWithoutTerminalOutput("service", serviceInstanceName, "--guid")
	if err != nil {
		return "", fmt.Errorf("Failed to get service guid. %s", err.Error())
	}
	serviceGUID := strings.Trim(cmdOutput[0], " \n")

	return serviceGUID, nil
}

// CreateForward create forward for service
func (c *CfServiceJumperPlugin) CreateForward(serviceGUID string) (ForwardDataSet, error) {
	var forwardDataSet ForwardDataSet

	path := fmt.Sprintf("/services/%s/forwards", serviceGUID)
	url := fmt.Sprintf("%s%s", c.CfServiceJumperAPIEndpoint, path)

	request := gorequest.New()
	resp, body, errs := request.Post(url).Set("Authorization", c.CfServiceJumperAccessToken).End()
	if errs != nil {
		return forwardDataSet, fmt.Errorf("[ERR] cf service jumper request failed. %s", errs[0])
	}
	if resp.StatusCode != http.StatusOK {
		return forwardDataSet, fmt.Errorf("[ERR] cf service jumper request failed. status != 200.\n%s", body)
	}

	err := json.Unmarshal([]byte(body), &forwardDataSet)
	if err != nil {
		return forwardDataSet, fmt.Errorf("[ERR] cf service jumper request failed. unmarshal error: %s", err)
	}
	return forwardDataSet, nil
}

// DeleteForward delete forward for service
func (c *CfServiceJumperPlugin) DeleteForward(serviceGUID string, connectionID string) error {
	path := fmt.Sprintf("/services/%s/forwards/%s", serviceGUID, connectionID)
	url := fmt.Sprintf("%s%s", c.CfServiceJumperAPIEndpoint, path)

	request := gorequest.New()
	resp, body, errs := request.Delete(url).Set("Authorization", c.CfServiceJumperAccessToken).End()
	if errs != nil {
		return fmt.Errorf("[ERR] Failed cf_service_jumper request. %s", errs[0].Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[ERR] Failed cf_service_jumper request. HTTP status code != 200.\n%s", body)
	}

	fmt.Println(body)
	return nil
}

// ListForwards list all forwards for the given service
func (c *CfServiceJumperPlugin) ListForwards(serviceGUID string) error {
	path := fmt.Sprintf("/services/%s/forwards/", serviceGUID)
	url := fmt.Sprintf("%s%s", c.CfServiceJumperAPIEndpoint, path)

	request := gorequest.New()
	resp, body, errs := request.Get(url).Set("Authorization", c.CfServiceJumperAccessToken).End()
	if errs != nil {
		return fmt.Errorf("Failed cf_service_jumper request. %s", errs[0].Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed cf_service_jumper request. HTTP status code != 200.\n%s", body)
	}

	fmt.Println(body)
	return nil
}

// Run This function must be implemented by any plugin because it is part of the
// plugin interface defined by the core CLI.
//
// Run(....) is the entry point when the core CLI is invoking a command defined
// by a plugin. The first parameter, plugin.CliConnection, is a struct that can
// be used to invoke cli commands. The second paramter, args, is a slice of
// strings. args[0] will be the name of the command, and will be followed by
// any additional arguments a cli user typed in.
//
// Any error handling should be handled with the plugin itself (this means printing
// user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
// 1 should the plugin exits nonzero.
func (c *CfServiceJumperPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		os.Exit(0)
	}

	serviceInstanceName, err := ArgsExtractServiceInstanceName(args)
	fatalIf(err)

	serviceGUID, err := c.FetchServiceGUID(cliConnection, serviceInstanceName)
	fatalIf(err)

	c.CfServiceJumperAccessToken, err = cliConnection.AccessToken()
	fatalIf(err)

	apiEndpoint, err := cliConnection.ApiEndpoint()
	fatalIf(err)

	c.CfServiceJumperAPIEndpoint, err = FetchCfServiceJumperAPIEndpoint(apiEndpoint)
	fatalIf(err)

	if args[0] == "create-forward" {
		forwardInfo, err := c.CreateForward(serviceGUID)
		fatalIf(err)
		credentials := forwardInfo.CredentialsMap()

		xt := xtunnel.NewUnencryptedXTunnel(forwardInfo.Hosts[0])
		localListenAddress, err := xt.Listen()
		fatalIf(err)

		fmt.Println(fmt.Sprintf("Listening on %s", localListenAddress))
		fmt.Println("\nCredentials:")
		for credentialKey, credentialValue := range credentials {
			fmt.Println(fmt.Sprintf("%s: %s", credentialKey, credentialValue))
		}

		go func() {
			xt.Serve()
		}()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

		_ = <-c
		err = xt.Shutdown()
		if err != nil {
			fmt.Println("[ERR] Failed to shutdown listen socket", err)
		}

		fmt.Println("\nRemember to 'cf delete-forward'!")

	} else if args[0] == "delete-forward" {
		connectionID, err := ArgsExtractConnectionID(args)
		fatalIf(err)
		err = c.DeleteForward(serviceGUID, connectionID)
	} else if args[0] == "list-forwards" {
		err = c.ListForwards(serviceGUID)
	}
	fatalIf(err)
}

// GetMetadata must be implemented as part of the plugin interface
// defined by the core CLI.
//
// GetMetadata() returns a PluginMetadata struct. The first field, Name,
// determines the name of the plugin which should generally be without spaces.
// If there are spaces in the name a user will need to properly quote the name
// during uninstall otherwise the name will be treated as seperate arguments.
// The second value is a slice of Command structs. Our slice only contains one
// Command Struct, but could contain any number of them. The first field Name
// defines the command `cf basic-plugin-command` once installed into the CLI. The
// second field, HelpText, is used by the core CLI to display help information
// to the user in the core commands `cf help`, `cf`, or `cf -h`.
func (c *CfServiceJumperPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "CfServiceJumperPlugin",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			plugin.Command{
				Name:     "create-forward",
				HelpText: "Creates forward to service instance.",
				UsageDetails: plugin.Usage{
					Usage: "\n   cf create-forward SERVICE_INSTANCE",
				},
			},
			plugin.Command{
				Name:     "delete-forward",
				HelpText: "Deletes forward to service instance.",
				UsageDetails: plugin.Usage{
					Usage: "\n   cf delete-forward SERVICE_INSTANCE CONNECTION_ID",
				},
			},
			plugin.Command{
				Name:     "list-forwards",
				HelpText: "List open forwards to service instance.",
				UsageDetails: plugin.Usage{
					Usage: "\n   cf list-forwards",
				},
			},
		},
	}
}

// https://github.com/cloudfoundry/cli/tree/master/plugin_examples
func main() {
	plugin.Start(new(CfServiceJumperPlugin))
}
