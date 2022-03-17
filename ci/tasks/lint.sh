#!/usr/bin/env bash
set -euo pipefail

cd control-tower

cp ../control-tower-ops/manifest.yml opsassets/assets/
cp -R ../control-tower-ops/ops opsassets/assets/
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-aws.json opsassets/assets/
cp ../control-tower-ops/createenv-dependencies-and-cli-versions-gcp.json opsassets/assets/

gometalinter \
--disable-all \
--enable=goconst \
--enable=ineffassign \
--enable=vetshadow \
--enable=deadcode \
--vendor \
--enable-gc \
--deadline=120s \
./...

# Globally ignoring SC2154 as it doesn't play nice with variables
# set by Concourse for use in tasks.
# https://github.com/koalaman/shellcheck/wiki/SC2154
find . -name vendor -prune ! -type d -o -name '*.sh' -exec shellcheck -e SC2154 {} +
