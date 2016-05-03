#!/bin/bash
# vim: set ft=sh

set -e

export GOPATH=$PWD/gopath
export PATH=$GOPATH/bin:$PATH

cd $GOPATH/src/github.com/anynines/cf_service_jumper_cli_plugin

export GOPATH=${PWD}/Godeps/_workspace:$GOPATH
export PATH=${PWD}/Godeps/_workspace/bin:$PATH

go get github.com/mitchellh/gox

if [[ "$(which gox)X" == "X" ]]; then
  echo "Please install gox. https://github.com/mitchellh/gox#readme"
  exit 1
fi

rm -f cf_service_jumper_cli_plugin*

gox -os linux -os windows -arch 386
gox -os darwin -os linux -os windows -arch amd64

rm -rf out
mkdir -p out
mv cf_service_jumper_cli_plugin* out/
