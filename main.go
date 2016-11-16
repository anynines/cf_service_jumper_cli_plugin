package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/anynines/cf_service_jumper_cli_plugin/plugin/config"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/parnurzeal/gorequest"
)

func fatalIf(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, "error: ", err)
		os.Exit(1)
	}
}

var (
	ErrMissingServiceInstanceArg              = errors.New("[ERR] missing SERVICE_INSTANCE")
	ErrMissingConnectionID                    = errors.New("[ERR] missing CONNECTION_ID")
	ErrCfServiceJumperEndpointGetFailed       = errors.New("[ERR] Failed to fetch information from Cloud Foundry api endpoint")
	ErrCfServiceJumperEndpointStatusCodeWrong = errors.New("[ERR] Failed to fetch information from Cloud Foundry api endpoint. HTTP status code != 200")
	ErrCfServiceJumperEndpointNotPresent      = errors.New("[ERR] cf service jumper api endpoint not present/installed")
	PcfServiceJumperHostname                  = "a9s-service-jumper"
)

// ArgsExtractServiceInstanceName extract service instance name from args
func ArgsExtractServiceInstanceName(args []string) (string, error) {
	if len(args) < 2 {
		return "", ErrMissingServiceInstanceArg
	}

	return args[1], nil
}

// ArgsExtractConnectionID extracts connection ID from args
func ArgsExtractConnectionID(args []string) (string, error) {
	if len(args) < 3 {
		return "", ErrMissingConnectionID
	}

	return args[2], nil
}

// FetchCfServiceJumperAPIEndpoint fetches the Service Jumper API endpoint
func FetchCfServiceJumperAPIEndpoint(cfAPIEndpoint string, isSSLDisabled bool) (string, error) {
	endpoint, err := FetchCfServiceJumperAPIEndpointFromConfig()
	if err == nil && endpoint != "" {
		return endpoint, nil
	}

	endpoint, err = FetchCfServiceJumperAPIEndpointFromInfo(cfAPIEndpoint, isSSLDisabled)
	if err == nil {
		return endpoint, nil
	}

	return FetchCfServiceJumperAPIEndpointFromSharedDomain(cfAPIEndpoint)
}

func FetchCfServiceJumperAPIEndpointFromConfig() (string, error) {
	forwardConfig, err := config.GetConfig()
	if err == config.ErrForwardConfigMissing {
		return "", config.ErrTargetBlank
	}
	if forwardConfig.Target == "" {
		return "", config.ErrTargetBlank
	}
	return forwardConfig.Target, nil
}

func FetchCfServiceJumperAPIEndpointFromInfo(cfAPIEndpoint string, isSSLDisabled bool) (string, error) {
	url := fmt.Sprintf("%s/v2/info", cfAPIEndpoint)

	httpClient := NewHttpClient(isSSLDisabled)
	resp, body, errs := httpClient.Get(url).End()

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

func FetchCfServiceJumperAPIEndpointFromSharedDomain(cfAPIEndpoint string) (string, error) {
	u, err := url.Parse(cfAPIEndpoint)
	if err != nil {
		return "", err
	}

	u.Host = strings.TrimPrefix(u.Host, "api.")
	u.Host = PcfServiceJumperHostname + "." + u.Host
	return u.String(), nil
}

// CfServiceJumperPlugin This is the struct implementing the interface defined by the core CLI. It can
// be found at  "https://github.com/cloudfoundry/cli/blob/master/plugin/plugin.go"
type CfServiceJumperPlugin struct {
	CfServiceJumperAccessToken string
	CfServiceJumperAPIEndpoint string

	isSSLDisabled bool
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

func (c *CfServiceJumperPlugin) NewHttpClient() *gorequest.SuperAgent {
	return NewHttpClient(c.isSSLDisabled)
}

func (c *CfServiceJumperPlugin) NewUrl(path string) string {
	u := fmt.Sprintf("%s%s", c.CfServiceJumperAPIEndpoint, path)
	if c.isSSLDisabled {
		u = u + "?skip-ssl-validation=true"
	}
	return u
}

// CreateForward create forward for service
func (c *CfServiceJumperPlugin) CreateForward(serviceGUID string) (ForwardDataSet, error) {
	var forwardDataSet ForwardDataSet

	path := fmt.Sprintf("/services/%s/forwards", serviceGUID)
	url := c.NewUrl(path)

	httpClient := c.NewHttpClient()
	resp, body, errs := httpClient.Post(url).Set("Authorization", c.CfServiceJumperAccessToken).End()
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
	url := c.NewUrl(path)

	httpClient := c.NewHttpClient()
	resp, body, errs := httpClient.Delete(url).Set("Authorization", c.CfServiceJumperAccessToken).End()
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
	url := c.NewUrl(path)

	httpClient := c.NewHttpClient()
	resp, body, errs := httpClient.Get(url).Set("Authorization", c.CfServiceJumperAccessToken).End()
	if errs != nil {
		return fmt.Errorf("Failed cf_service_jumper request. %s", errs[0].Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed cf_service_jumper request. HTTP status code != 200.\n%s", body)
	}

	var forwardDataSetCollection []ForwardDataSet
	err := json.Unmarshal([]byte(body), &forwardDataSetCollection)
	if err != nil {
		return fmt.Errorf("[ERR] cf service jumper request failed. unmarshal error: %s", err)
	}

	OutputForwardDataSets(forwardDataSetCollection)

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
	var err error

	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		os.Exit(0)
	}

	if args[0] == "forward-api" {
		if len(args) > 1 {
			if args[1] == "-d" {
				// delete forward api endpoint
				err = config.SetTarget("")
				fatalIf(err)
				return
			}

			// set forward endpoint
			_, err := url.ParseRequestURI(args[1])
			fatalIf(err)

			err = config.SetTarget(args[1])
			fatalIf(err)
			fmt.Printf("forwarf-api set to %s\n", args[1])
			return
		}

		// show forward endpoint
		endpoint, err := FetchCfServiceJumperAPIEndpointFromConfig()
		fatalIf(err)
		fmt.Printf("forward-api %s\n", endpoint)
		return
	}

	c.isSSLDisabled, err = cliConnection.IsSSLDisabled()
	fatalIf(err)

	serviceInstanceName, err := ArgsExtractServiceInstanceName(args)
	fatalIf(err)

	serviceGUID, err := c.FetchServiceGUID(cliConnection, serviceInstanceName)
	fatalIf(err)

	c.CfServiceJumperAccessToken, err = cliConnection.AccessToken()
	fatalIf(err)

	apiEndpoint, err := cliConnection.ApiEndpoint()
	fatalIf(err)

	c.CfServiceJumperAPIEndpoint, err = FetchCfServiceJumperAPIEndpoint(apiEndpoint, c.isSSLDisabled)
	fatalIf(err)

	if args[0] == "create-forward" {
		forwardInfo, err := c.CreateForward(serviceGUID)
		fatalIf(err)
		credentials := forwardInfo.CredentialsMap()

		fmt.Println("\nCredentials:")
		for credentialKey, credentialValue := range credentials {
			if stringInStrSlice(credentialKey, []string{"uri", "host", "port"}) {
				continue
			}

			fmt.Println(fmt.Sprintf("%s: %s", credentialKey, credentialValue))
		}
		fmt.Printf("\n")

		connectionPrinter := SelectConnectionPrinter(credentials)
		ListenAndOutputInfo(forwardInfo.Hosts, forwardInfo.SharedSecret, connectionPrinter)

		fmt.Println("\nRemember to 'cf delete-forward'!")

	} else if args[0] == "delete-forward" {
		connectionID, err := ArgsExtractConnectionID(args)
		fatalIf(err)
		err = c.DeleteForward(serviceGUID, connectionID)
		fatalIf(err)
	} else if args[0] == "list-forwards" {
		err = c.ListForwards(serviceGUID)
		fatalIf(err)
	}
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
			Major: 2,
			Minor: 0,
			Build: 0,
		},
		Commands: []plugin.Command{
			plugin.Command{
				Name:     "create-forward",
				HelpText: "Creates/Recycles forward to service instance.",
				UsageDetails: plugin.Usage{
					Usage: "cf create-forward SERVICE_INSTANCE",
				},
			},
			plugin.Command{
				Name:     "delete-forward",
				HelpText: "Deletes forward to service instance.",
				UsageDetails: plugin.Usage{
					Usage: "cf delete-forward SERVICE_INSTANCE CONNECTION_ID",
				},
			},
			plugin.Command{
				Name:     "list-forwards",
				HelpText: "List open forwards to service instance.",
				UsageDetails: plugin.Usage{
					Usage: "cf list-forwards",
				},
			},
			plugin.Command{
				Name:     "forward-api",
				HelpText: "Show/Set/Delete service jumper api url.",
				UsageDetails: plugin.Usage{
					Usage: "cf forward-api SERVICE_JUMPER_API_URL\ncf forward-api [-d]",
				},
			},
		},
	}
}

// https://github.com/cloudfoundry/cli/tree/master/plugin_examples
func main() {
	plugin.Start(new(CfServiceJumperPlugin))
}
