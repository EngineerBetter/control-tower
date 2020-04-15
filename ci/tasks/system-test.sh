#!/bin/bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/test-setup.sh

handleVerboseMode
setDeploymentName sys

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/check-config-bucket-versioned.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/check-db.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/check-cidr-ranges.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/manifest_property.sh

trapDefaultCleanup

cp "$BINARY_PATH" ./cup
chmod +x ./cup

if [ "$IAAS" = "GCP" ]; then
  custom_domain="$deployment-user.gcp.engineerbetter.com"
else
  custom_domain="$deployment-user.control-tower.engineerbetter.com"
fi

certstrap init \
  --common-name "$deployment" \
  --passphrase "" \
  --organization "" \
  --organizational-unit "" \
  --country "" \
  --province "" \
  --locality ""

certstrap request-cert \
   --passphrase "" \
   --domain "$custom_domain"

certstrap sign "$custom_domain" --CA "$deployment"

echo "DEPLOY WITH A USER PROVIDED CERT, CUSTOM DOMAIN, DEFAULT WORKERS, DEFAULT DATABASE SIZE AND DEFAULT WEB NODE SIZE"

./cup deploy "$deployment" \
  --domain "$custom_domain" \
  --spot=false \
  --tls-cert "$(cat out/"$custom_domain".crt)" \
  --tls-key "$(cat out/"$custom_domain".key)" \
  --enable-global-resources=true

sleep 60

# Check we can log into the BOSH director and SSH into a VM
# Assigning a subshell to a variable fails fast; eval "$(... doesn't
info_output="$(./cup info --env "$deployment")"
eval "$info_output"
bosh vms
bosh ssh worker true

config=$(./cup info --json "$deployment")
# shellcheck disable=SC2034
username=$(echo "$config" | jq -r '.config.concourse_username')
# shellcheck disable=SC2034
password=$(echo "$config" | jq -r '.config.concourse_password')
echo "$config" | jq -r '.config.concourse_cert' > generated-ca-cert.pem

if [ "$IAAS" = "GCP" ]
then
  gcloud auth activate-service-account --key-file="$GOOGLE_APPLICATION_CREDENTIALS"
  export CLOUDSDK_CORE_PROJECT=control-tower-233017
fi

# Check RDS instance class is db.t2.small
assertDbCorrect
assertNetworkCidrsCorrect
assertConfigBucketVersioned

# Check Concourse global resources is enabled
global_resources_path="/instance_groups/name=web/jobs/name=web/properties/enable_global_resources"
checkManifestProperty "${global_resources_path}" true

# shellcheck disable=SC2034
cert="generated-ca-cert.pem"
# shellcheck disable=SC2034
manifest="$(dirname "$0")/hello.yml"
# shellcheck disable=SC2034
job="hello"
# shellcheck disable=SC2034
domain=$custom_domain

assertPipelineIsSettableAndRunnable

echo "DEPLOY 2 LARGE WORKERS, FIREWALLED TO MY IP"


./cup deploy "$deployment" \
  --allow-ips "$(dig +short myip.opendns.com @resolver1.opendns.com)" \
  --workers 2 \
  --worker-size large \
  --enable_global_resources=false

sleep 60

# Check RDS instance class is still db.t2.small
assertDbCorrect

# Check Concourse global resources is disabled
checkManifestProperty "${global_resources_path}" false

config=$(./cup info --json "$deployment")
# shellcheck disable=SC2034
username=$(echo "$config" | jq -r '.config.concourse_username')
# shellcheck disable=SC2034
password=$(echo "$config" | jq -r '.config.concourse_password')
echo "$config" | jq -r '.config.concourse_cert' > generated-ca-cert.pem
# shellcheck disable=SC2034
cert="generated-ca-cert.pem"

assertPipelineIsRunnable
assertPipelinesCanReadFromCredhub

# Check nats certificate renewal
before="$(./cup info "$deployment" --cert-expiry)"
before_timestamp="$(date -d "$before" +"%s")"

./cup maintain --renew-nats-cert "$deployment"

after="$(./cup info "$deployment" --cert-expiry)"
after_timestamp="$(date -d "$after" +"%s")"

[[ $before_timestamp -lt $after_timestamp ]]

sleep 60

assertPipelinesCanReadFromCredhub

assertGrafanaPresent
assertConcourseDashboardPresent
