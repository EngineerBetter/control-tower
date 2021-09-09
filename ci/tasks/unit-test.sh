#!/usr/bin/env bash
set -euo pipefail

cd control-tower

go install github.com/maxbrunsfeld/counterfeiter/v6

cp ../control-tower-ops/manifest.yml opsassets/assets/
cp -R ../control-tower-ops/ops opsassets/assets/ 
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-aws.json opsassets/assets/
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-gcp.json opsassets/assets/

go generate ./...
go test ./...
