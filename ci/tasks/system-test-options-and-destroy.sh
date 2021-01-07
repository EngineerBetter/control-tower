#!/bin/bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/test-setup.sh

handleVerboseMode
setDeploymentName opt

# Create empty array of args that is used in sourced setup functions
args=()
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/github-auth.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/tags.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/letsencrypt.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/assert-iaas.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/check-cidr-ranges.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/manifest_property.sh

# shellcheck disable=SC1091
[ "$IAAS" = "AWS" ] && { source control-tower/ci/tasks/lib/destroy.sh; }

# shellcheck disable=SC1091
[ "$IAAS" = "GCP" ] && { source control-tower/ci/tasks/lib/gcp-destroy.sh; }

cp "$BINARY_PATH" ./cup
chmod +x ./cup

if [ "$IAAS" = "AWS" ]
then
    # shellcheck disable=SC2034
    region=us-east-1
    domain=ct.engineerbetter.com

    args+=(--vpc-network-range 192.168.0.0/24)
    args+=(--rds-subnet-range1 192.168.0.64/28)
    args+=(--rds-subnet-range2 192.168.0.80/28)
elif [ "$IAAS" = "GCP" ]
then
    # shellcheck disable=SC2034
    region=europe-west2
    domain=ct.gcp.engineerbetter.com
fi

args+=(--domain "${domain}")
args+=(--public-subnet-range 192.168.0.0/27)
args+=(--private-subnet-range 192.168.0.32/27)

trapCustomCleanup

addBitBucketFlagsToArgs
addGitHubFlagsToArgs
addTagsFlagsToArgs
args+=(--region "$region")
./cup deploy "${args[@]}" --iaas "$IAAS" "$deployment"

# Download the right version of fly from Concourse UI
updateFly "${domain}"

assertTagsSet
assertBitBucketAuthConfigured
assertGitHubAuthConfigured

# Check Concourse global resources is disabled (as it should be by default)
info_output="$(./cup info --region "$region" --env "$deployment")"
eval "$info_output"
global_resources_path="/instance_groups/name=web/jobs/name=web/properties/enable_global_resources"
checkManifestProperty "${global_resources_path}" false

if [ "$IAAS" = "AWS" ]
then
    assertNetworkCidrsCorrect 192.168.0.0/27 192.168.0.32/27 192.168.0.0/24 192.168.0.64/28 192.168.0.80/28
elif [ "$IAAS" = "GCP" ]
then
    assertNetworkCidrsCorrect 192.168.0.0/27 192.168.0.32/27
fi

assertPipelinesCanReadFromCredhub
sleep 60
recordDeployedState
echo "non-interactive destroy"
./cup --non-interactive destroy "$deployment" -iaas "$IAAS" --region "$region"
sleep 180
assertEverythingDeleted
