#!/usr/bin/env bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/set-flags.sh

build_dir=$PWD/build-$GOOS
mkdir -p build_dir

if [ -e "version/version" ]; then
  version=$(cat version/version)
else
  version="TESTVERSION"
fi

GOOS=linux go get -u github.com/kevinburke/go-bindata/...

cd control-tower || exit 1

grep -lr --include=*.go --exclude-dir=vendor "go:generate go-bindata" . | xargs -I {} go generate {}
GO111MODULE=on go build -mod=vendor -ldflags "
  -X github.com/EngineerBetter/control-tower/fly.ControlTowerVersion=$version
  -X main.ControlTowerVersion=$version
" -o "$build_dir/$OUTPUT_FILE"
