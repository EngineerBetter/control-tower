#!/usr/bin/env bash
set -euo pipefail

build_dir=$PWD/build-$GOOS
mkdir -p build_dir

if [ -e "version/version" ]; then
  version=$(cat version/version)
else
  version="TESTVERSION"
fi
cd control-tower || exit 1

cp ../control-tower-ops/manifest.yml opsassets/assets/
cp -R ../control-tower-ops/ops opsassets/assets/ 
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-aws.json opsassets/assets/
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-gcp.json opsassets/assets/
GO111MODULE=on go build -mod=vendor -ldflags "
  -X github.com/EngineerBetter/control-tower/fly.ControlTowerVersion=$version
  -X main.ControlTowerVersion=$version
" -o "$build_dir/$OUTPUT_FILE"
