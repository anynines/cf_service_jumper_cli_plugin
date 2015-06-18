# Cf Service Jumper CLI Plugin

This CF cli plugin uses the CF Service Jumper service https://github.com/anynines/cf_service_jumper
to create, delete and list forwards to services.

## Development

```shell
cf uninstall-plugin CfServiceJumperPlugin  
go build -o CfServiceJumperPlugin
cf install-plugin CfServiceJumperPlugin  
```


## Usage
```shell
cf create-forward SERVICE_NAMe
cf delete-forward SERVICE_NAME FORWARD_ID
cf list-forwards SERVICE_NAMe
```
