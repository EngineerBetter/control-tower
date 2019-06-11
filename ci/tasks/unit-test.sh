#!/bin/bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/set-flags.sh

mkdir -p "$GOPATH/src/github.com/EngineerBetter/control-tower"
mkdir -p "$GOPATH/src/github.com/EngineerBetter/control-tower-ops"
mv control-tower/* "$GOPATH/src/github.com/EngineerBetter/control-tower"
mv control-tower-ops/* "$GOPATH/src/github.com/EngineerBetter/control-tower-ops"
cd "$GOPATH/src/github.com/EngineerBetter/control-tower" || exit 1

GO111MODULE=off go get -u github.com/mattn/go-bindata/...
GO111MODULE=off go get -u github.com/maxbrunsfeld/counterfeiter
go generate bosh/data.go
go generate resource/package.go
go generate github.com/EngineerBetter/control-tower/...
go test ./...
