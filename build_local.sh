#!/bin/bash

set -eu

version=dev
grep -lr --include=*.go --exclude-dir=vendor "go:generate go-bindata" . | xargs -I {} go generate {}
GO111MODULE=on go build -mod=vendor -ldflags "
  -X github.com/EngineerBetter/control-tower/fly.ControlTowerVersion=$version
  -X main.ControlTowerVersion=$version
" -o control-tower

chmod +x control-tower

echo "$PWD/control-tower"
