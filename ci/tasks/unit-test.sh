#!/bin/bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/set-flags.sh

mkdir -p "$GOPATH/src/github.com/EngineerBetter/control-tower"
mkdir -p "$GOPATH/src/github.com/EngineerBetter/control-tower-ops"
mv control-tower/* "$GOPATH/src/github.com/EngineerBetter/control-tower"
mv control-tower-ops/* "$GOPATH/src/github.com/EngineerBetter/control-tower-ops"
cd "$GOPATH/src/github.com/EngineerBetter/control-tower" || exit 1

GO111MODULE=off go get -u github.com/maxbrunsfeld/counterfeiter

cp "$GOPATH/src/github.com/EngineerBetter/control-tower-ops/manifest.yml" opsassets/assets/
cp -R "$GOPATH/src/github.com/EngineerBetter/control-tower-ops/ops" opsassets/assets/ 
cp "$GOPATH/src/github.com/EngineerBetter/control-tower-ops/createenv-dependencies-and-cli-versions-aws.json" opsassets/assets/
cp "$GOPATH/src/github.com/EngineerBetter/control-tower-ops/createenv-dependencies-and-cli-versions-gcp.json" opsassets/assets/

go generate github.com/EngineerBetter/control-tower/...
go test ./...
