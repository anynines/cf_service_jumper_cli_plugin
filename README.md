# Cf Service Jumper CLI Plugin

This CF cli plugin uses the CF Service Jumper service https://github.com/anynines/cf_service_jumper
to create, delete and list forwards to services.

## Development

### Prerequisites

```shell
go get github.com/cloudfoundry/cli/plugin
go get github.com/parnurzeal/gorequest
```

### Build, Install & Uninstall

```shell
cf uninstall-plugin CfServiceJumperPlugin  
go build -o CfServiceJumperPlugin
cf install-plugin CfServiceJumperPlugin  
```


## Usage
```shell
cf create-forward SERVICE_NAME
cf delete-forward SERVICE_NAME FORWARD_ID
cf list-forwards SERVICE_NAME
```
