#!/bin/bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/set-flags.sh

build_dir=$PWD/build-$GOOS
mkdir -p build_dir

if [ -e "version/version" ]; then
  version=$(cat version/version)
else
  version="TESTVERSION"
fi

mkdir -p "$GOPATH/src/github.com/EngineerBetter/control-tower"
mkdir -p "$GOPATH/src/github.com/EngineerBetter/control-tower-ops"
mv control-tower/* "$GOPATH/src/github.com/EngineerBetter/control-tower"
mv control-tower-ops/* "$GOPATH/src/github.com/EngineerBetter/control-tower-ops"
cd "$GOPATH/src/github.com/EngineerBetter/control-tower" || exit 1

GOOS=linux go get -u github.com/mattn/go-bindata/...

grep -lr --include=*.go --exclude-dir=vendor "go:generate go-bindata" . | xargs -I {} go generate {}
go build -ldflags "
  -X main.ControlTowerVersion=$version
  -X github.com/EngineerBetter/control-tower/fly.ControlTowerVersion=$version
" -o "$build_dir/$OUTPUT_FILE"
