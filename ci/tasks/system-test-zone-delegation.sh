#!/bin/bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/test-setup.sh

handleVerboseMode
setDeploymentName del

# Create empty array of args that is used in sourced setup functions
args=()
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/assert-iaas.sh

# shellcheck disable=SC1091
[ "$IAAS" = "AWS" ] && { source control-tower/ci/tasks/lib/destroy.sh; }

# shellcheck disable=SC1091
[ "$IAAS" = "GCP" ] && { source control-tower/ci/tasks/lib/gcp-destroy.sh; }

cp "$BINARY_PATH" ./cup
chmod +x ./cup

if [ "$IAAS" = "AWS" ]
then
    # shellcheck disable=SC2034
    domain="ct-delegation.engineerbetter.com"
elif [ "$IAAS" = "GCP" ]
then
    # shellcheck disable=SC2034
    domain="ct-delegation.gcp.engineerbetter.com"
fi

args+=(--domain "$domain")

trapDefaultCleanup

./cup deploy "${args[@]}" --iaas "$IAAS" "$deployment"

sleep 60

assertConcoursePresent
assertGrafanaPresent

recordDeployedState
echo "non-interactive destroy"
./cup --non-interactive destroy "$deployment" -iaas "$IAAS"
sleep 180
assertEverythingDeleted
