# Cf Service Jumper CLI Plugin

This CF cli plugin uses the [CF Service Jumper service](https://github.com/anynines/cf_service_jumper)
to create, delete and list forwards to services.

## Installation

Download the latest release for your platform from the [release page](https://github.com/anynines/cf_service_jumper_cli_plugin/releases).
You must have `read` and `execute` permissions to the plugin binary file.
Install the plugin using the following command:
```
cf install-plugin cf_service_jumper_cli_plugin_YOUR_OS_AND_ARCH
```

If you want to remove the plugin, execute:
```
cf uninstall-plugin CfServiceJumperPlugin
```

## Development

### Prerequisites

```shell
go get -u github.com/kardianos/govendor
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

Increase version in file `VERSION` if required.

Create a new release with version `VERSION` on the [release page](https://github.com/anynines/cf_service_jumper_cli_plugin/releases).

```sh
./bin/release.sh
```
