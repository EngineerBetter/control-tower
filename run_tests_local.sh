#!/bin/bash
#shellcheck disable=SC2016

set -eu

docker run -it \
    -v "${PWD}:/mnt/control-tower" \
    -v "${PWD}/../control-tower-ops:/mnt/control-tower-ops" \
    engineerbetter/pcf-ops \
    bash -xc \
    'cp -r /mnt/control-tower* .; ./control-tower/ci/tasks/lint.sh && cd control-tower && go test ./...'
