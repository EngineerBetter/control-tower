#!/bin/bash

set -eu

version=dev
cp ../control-tower-ops/manifest.yml opsassets/assets/
cp -R ../control-tower-ops/ops opsassets/assets/ 
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-aws.json opsassets/assets/
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-gcp.json opsassets/assets/
GO111MODULE=on go build -mod=vendor -ldflags "
  -X github.com/EngineerBetter/control-tower/fly.ControlTowerVersion=$version
  -X main.ControlTowerVersion=$version
" -o control-tower

chmod +x control-tower

echo "$PWD/control-tower"
