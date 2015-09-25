# Cf Service Jumper CLI Plugin

This CF cli plugin uses the [CF Service Jumper service](https://github.com/a9hcp/cf_service_jumper)
to create, delete and list forwards to services.

## Development

### Prerequisites

```shell
go get github.com/tools/godep
godep restore
```

### Build, Install & Uninstall

```shell
cf uninstall-plugin CfServiceJumperPlugin  
go build -o CfServiceJumperPlugin
cf install-plugin CfServiceJumperPlugin  
```

### Testing

We're using [ginko](https://github.com/onsi/ginkgo) as testing framework.
 ```shell
go test
```

## Usage
```shell
cf create-forward SERVICE_NAME
cf delete-forward SERVICE_NAME FORWARD_ID
cf list-forwards SERVICE_NAME
```

## Release

Build the release binaries and output to directory `out/`.

```sh
./bin/build.sh
```



[Create a GitHub token](https://help.github.com/articles/creating-an-access-token-for-command-line-use)
and set it as the env variable `GITHUB_TOKEN`.

```sh
export GITHUB_TOKEN=...
```
