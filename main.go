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
)

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, "error: ", err)
		os.Exit(1)
	}
}

/*
*	This is the struct implementing the interface defined by the core CLI. It can
*	be found at  "github.com/cloudfoundry/cli/plugin/plugin.go"
*
 */
type CfServiceJumperPlugin struct{}

func (c *CfServiceJumperPlugin) ExtractServiceInstanceName(args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("missing SERVICE_INSTANCE")
	}

	return args[1], nil
}

func (c *CfServiceJumperPlugin) FetchServiceGuid(cliConnection plugin.CliConnection, serviceInstanceName string) (string, error) {
	cmdOutput, err := cliConnection.CliCommandWithoutTerminalOutput("service", serviceInstanceName, "--guid")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to get service guid. %s", err.Error()))
	}
	service_guid := strings.Trim(cmdOutput[0], " \n")

	return service_guid, nil
}

func (c *CfServiceJumperPlugin) FetchCfServiceJumperApiEndpoint(cliConnection plugin.CliConnection, serviceGuid string) (string, error) {
	apiEndpoint, err := cliConnection.ApiEndpoint()
	if err != nil {
		return "", err

	}

	url := fmt.Sprintf("%s/v2/info", apiEndpoint)

	request := gorequest.New()
	resp, body, errs := request.Get(url).End()
	if errs != nil {
		return "", errors.New(fmt.Sprintf("Failed to fetch cf service jumper api endpoint. %s", errs[0].Error()))
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Failed to fetch cf api endpoint (http status != 200)")
	}

	type CfInfo struct {
		Custom map[string]string `json:"custom"`
	}
	var cfInfo CfInfo
	err = json.Unmarshal([]byte(body), &cfInfo)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to fetch cf service jumper api endpoint. %s", err.Error()))
	}
	serviceJumperEndpoint := cfInfo.Custom["service_jumper_endpoint"]
	if len(serviceJumperEndpoint) < 1 {
		return "", errors.New("Failed to fetch cf service jumper api endpoint")
	}
	return serviceJumperEndpoint, nil
}

func (c *CfServiceJumperPlugin) CreateForward(cliConnection plugin.CliConnection, serviceGuid string, cfServiceJumperApiEndpoint string) error {
	path := fmt.Sprintf("/services/%s/forwards", serviceGuid)
	url := fmt.Sprintf("%s%s", cfServiceJumperApiEndpoint, path)

	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return err
	}

	request := gorequest.New()
	resp, body, errs := request.Post(url).Set("Authorization", accessToken).End()
	if errs != nil {
		return errors.New(fmt.Sprintf("Failed cf_service_jumper request. %s", errs[0].Error()))
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Failed cf_service_jumper request. status != 200 ; body = %s", body))
	}

	fmt.Println(body)
	return nil
}

func (c *CfServiceJumperPlugin) DeleteForward(cliConnection plugin.CliConnection, serviceGuid string, connectionId string, cfServiceJumperApiEndpoint string) error {
	path := fmt.Sprintf("/services/%s/forwards/%s", serviceGuid, connectionId)
	url := fmt.Sprintf("%s%s", cfServiceJumperApiEndpoint, path)

	accessToken, err := cliConnection.AccessToken()
	if err != nil {
		return err
	}

	request := gorequest.New()
	resp, body, errs := request.Post(url).Set("Authorization", accessToken).End()
	if errs != nil {
		return errors.New(fmt.Sprintf("Failed cf_service_jumper request. %s", errs[0].Error()))
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("Failed cf_service_jumper request. status != 200 ; body = %s", body))
	}

	fmt.Println(body)
	return nil
}

/*
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
	serviceInstanceName, err := c.ExtractServiceInstanceName(args)
	fatalIf(err)

	serviceGuid, err := c.FetchServiceGuid(cliConnection, serviceInstanceName)
	fatalIf(err)

	cfServiceJumperApiEndpoint, err := c.FetchCfServiceJumperApiEndpoint(cliConnection, serviceGuid)
	fatalIf(err)

	err = nil
	if args[0] == "create-forward" {
		err = c.CreateForward(cliConnection, serviceGuid, cfServiceJumperApiEndpoint)
	} else if args[0] == "delete-forward" {
		connectionId := args[2]
		err = c.DeleteForward(cliConnection, serviceGuid, connectionId, cfServiceJumperApiEndpoint)
	}
	fatalIf(err)
}

/*
*	This function must be implemented as part of the	plugin interface
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
					Usage: "\n   cf delete-forward SERVICE_INSTANCE",
				},
			},
		},
	}
}

/*
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
