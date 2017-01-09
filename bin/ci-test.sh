#!/bin/bash
# vim: set ft=sh

set -eu

export GOPATH=$PWD/gopath
export PATH=$GOPATH/bin:$PATH

cd $GOPATH/src/github.com/anynines/cf_service_jumper_cli_plugin

go test
