#!/bin/bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/test-setup.sh
echo "RUNNING SYSTEM-TEST-OPTIONS-AND-DESTROY"
handleVerboseMode
setDeploymentName opt
echo "MADE IT TO: LINE 8"
# Create empty array of args that is used in sourced setup functions
args=()
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/github-auth.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/bitbucket-auth.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/microsoft-auth.sh
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
echo "MADE IT TO: LINE 27"
# shellcheck disable=SC1091
[ "$IAAS" = "AWS" ] && { source control-tower/ci/tasks/lib/destroy.sh; }

# shellcheck disable=SC1091
[ "$IAAS" = "GCP" ] && { source control-tower/ci/tasks/lib/gcp-destroy.sh; }

cp "$BINARY_PATH" ./cup
chmod +x ./cup
echo "MADE IT TO: LINE 36"
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
echo "MADE IT TO: LINE 52"
args+=(--domain "${domain}")
args+=(--public-subnet-range 192.168.0.0/27)
args+=(--private-subnet-range 192.168.0.32/27)

trapCustomCleanup
echo "MADE IT TO: LINE 58"
addBitBucketFlagsToArgs
addGitHubFlagsToArgs
addMicrosoftFlagsToArgs
addTagsFlagsToArgs
args+=(--region "$region")
./cup deploy "${args[@]}" --iaas "$IAAS" "$deployment"
echo "MADE IT TO: LINE 65"
# Download the right version of fly from Concourse UI
updateFly "${domain}"

assertTagsSet
assertBitBucketAuthConfigured
assertGitHubAuthConfigured
assertMicrosoftAuthConfigured

# Check Concourse global resources & pipeline instances are disabled (as it should be by default)
info_output="$(./cup info --region "$region" --env "$deployment")"
echo "$info_output"

eval "$info_output"
global_resources_path="/instance_groups/name=web/jobs/name=web/properties/enable_global_resources"
checkManifestProperty "${global_resources_path}" false
pipeline_instances_path="/instance_groups/name=web/jobs/name=web/properties/enable_pipeline_instances"
checkManifestProperty "${pipeline_instances_path}" false

if [ "$IAAS" = "AWS" ]
then
    assertNetworkCidrsCorrect 192.168.0.0/27 192.168.0.32/27 192.168.0.0/24 192.168.0.64/28 192.168.0.80/28
elif [ "$IAAS" = "GCP" ]
then
    assertNetworkCidrsCorrect 192.168.0.0/27 192.168.0.32/27
fi
echo "MADE IT TO: LINE 91"
assertPipelinesCanReadFromCredhub
sleep 60
recordDeployedState
echo "non-interactive destroy"
./cup --non-interactive destroy "$deployment" -iaas "$IAAS" --region "$region"
sleep 180
assertEverythingDeleted
