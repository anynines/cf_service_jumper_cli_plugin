package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

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

var ErrMissingServiceInstanceArg = errors.New("[ERR] missing SERVICE_INSTANCE")
var ErrMissingConnectionId = errors.New("[ERR] missing CONNECTION_ID")
var ErrCfServiceJumperEndpointGetFailed = errors.New("[ERR] Failed to fetch information from Cloud Foundry api endpoint.")
var ErrCfServiceJumperEndpointStatusCodeWrong = errors.New("[ERR] Failed to fetch information from Cloud Foundry api endpoint. HTTP status code != 200.")
var ErrCfServiceJumperEndpointNotPresent = errors.New("[ERR] cf service jumper api endpoint not present/installed.")

func ArgsExtractServiceInstanceName(args []string) (string, error) {
	if len(args) < 2 {
		return "", ErrMissingServiceInstanceArg
	}

	return args[1], nil
}

func ArgsExtractConnectionId(args []string) (string, error) {
	if len(args) < 3 {
		return "", ErrMissingConnectionId
	}

	return args[2], nil
}

func FetchCfServiceJumperApiEndpoint(apiEndpoint string) (string, error) {
	url := fmt.Sprintf("%s/v2/info", apiEndpoint)

	request := gorequest.New()
	resp, body, errs := request.Get(url).End()

	if len(errs) > 0 {
		return "", ErrCfServiceJumperEndpointGetFailed
	}
	if resp.StatusCode != http.StatusOK {
		return "", ErrCfServiceJumperEndpointStatusCodeWrong
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
		return "", ErrCfServiceJumperEndpointNotPresent
	}

	return serviceJumperEndpoint, nil
}

/**
 *	This is the struct implementing the interface defined by the core CLI. It can
 *	be found at  "https://github.com/cloudfoundry/cli/blob/master/plugin/plugin.go"
 *
 */
type CfServiceJumperPlugin struct {
	CfServiceJumperAccessToken string
	CfServiceJumperApiEndpoint string
}

func (c *CfServiceJumperPlugin) FetchServiceGuid(cliConnection plugin.CliConnection, serviceInstanceName string) (string, error) {
	cmdOutput, err := cliConnection.CliCommandWithoutTerminalOutput("service", serviceInstanceName, "--guid")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to get service guid. %s", err.Error()))
	}
	service_guid := strings.Trim(cmdOutput[0], " \n")

	return service_guid, nil
}

func (c *CfServiceJumperPlugin) CreateForward(serviceGuid string) (ForwardDataSet, error) {
	var forwardDataSet ForwardDataSet

	path := fmt.Sprintf("/services/%s/forwards", serviceGuid)
	url := fmt.Sprintf("%s%s", c.CfServiceJumperApiEndpoint, path)

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

func (c *CfServiceJumperPlugin) DeleteForward(serviceGuid string, connectionId string) error {
	path := fmt.Sprintf("/services/%s/forwards/%s", serviceGuid, connectionId)
	url := fmt.Sprintf("%s%s", c.CfServiceJumperApiEndpoint, path)

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

func (c *CfServiceJumperPlugin) ListForwards(serviceGuid string) error {
	path := fmt.Sprintf("/services/%s/forwards/", serviceGuid)
	url := fmt.Sprintf("%s%s", c.CfServiceJumperApiEndpoint, path)

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

/**
 *	This function must be implemented by any plugin because it is part of the
 *	plugin interface defined by the core CLI.
 *
 *	Run(....) is the entry point when the core CLI is invoking a command defined
 *	by a plugin. The first parameter, plugin.CliConnection, is a struct that can
 *	be used to invoke cli commands. The second paramter, args, is a slice of
 *	strings. args[0] will be the name of the command, and will be followed by
 *	any additional arguments a cli user typed in.
 *
 *	Any error handling should be handled with the plugin itself (this means printing
 *	user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
 *	1 should the plugin exits nonzero.
 */
func (c *CfServiceJumperPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	serviceInstanceName, err := ArgsExtractServiceInstanceName(args)
	fatalIf(err)

	serviceGuid, err := c.FetchServiceGuid(cliConnection, serviceInstanceName)
	fatalIf(err)

	c.CfServiceJumperAccessToken, err = cliConnection.AccessToken()
	fatalIf(err)

	apiEndpoint, err := cliConnection.ApiEndpoint()
	fatalIf(err)

	c.CfServiceJumperApiEndpoint, err = FetchCfServiceJumperApiEndpoint(apiEndpoint)
	fatalIf(err)

	if args[0] == "create-forward" {
		forwardInfo, err := c.CreateForward(serviceGuid)
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

		err = xt.Serve() // forever
		fatalIf(err)
	} else if args[0] == "delete-forward" {
		connectionId, err := ArgsExtractConnectionId(args)
		fatalIf(err)
		err = c.DeleteForward(serviceGuid, connectionId)
	} else if args[0] == "list-forwards" {
		err = c.ListForwards(serviceGuid)
	}
	fatalIf(err)
}

/**
 *	This function must be implemented as part of the plugin interface
 *	defined by the core CLI.
 *
 *	GetMetadata() returns a PluginMetadata struct. The first field, Name,
 *	determines the name of the plugin which should generally be without spaces.
 *	If there are spaces in the name a user will need to properly quote the name
 *	during uninstall otherwise the name will be treated as seperate arguments.
 *	The second value is a slice of Command structs. Our slice only contains one
 *	Command Struct, but could contain any number of them. The first field Name
 *	defines the command `cf basic-plugin-command` once installed into the CLI. The
 *	second field, HelpText, is used by the core CLI to display help information
 *	to the user in the core commands `cf help`, `cf`, or `cf -h`.
 */
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

/**
 * Unlike most Go programs, the `Main()` function will not be used to run all of the
 * commands provided in your plugin. Main will be used to initialize the plugin
 * process, as well as any dependencies you might require for your
 * plugin.
 */
func main() {
	// Any initialization for your plugin can be handled here
	//
	// Note: to run the plugin.Start method, we pass in a pointer to the struct
	// implementing the interface defined at "github.com/cloudfoundry/cli/plugin/plugin.go"
	//
	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(CfServiceJumperPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}
