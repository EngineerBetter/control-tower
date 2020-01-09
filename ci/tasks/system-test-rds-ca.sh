#!/bin/bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/test-setup.sh
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/check-cidr-ranges.sh

handleVerboseMode

setDeploymentName rdsca

trapDefaultCleanup

cp "$BINARY_PATH" ./cup
chmod +x ./cup

echo "DEPLOY"

./cup deploy "$deployment"
assertNetworkCidrsCorrect

# Assigning a subshell to a variable fails fast; eval "$(... doesn't
info_output="$(./cup info --env "$deployment")"
eval "$info_output"
config=$(./cup info --json "$deployment")
[[ -n $config ]]
domain=$(echo "$config" | jq -r '.config.domain')

echo "Waiting for bosh lock to become available"
wait_time=0
until [[ $(bosh locks --json | jq -r '.Tables[].Rows | length') -eq 0 ]]; do
  (( ++wait_time ))
  if [[ $wait_time -ge 10 ]]; then
    echo "Waited too long for lock" && exit 1
  fi
  printf '.'
  sleep 60
done
echo "Bosh lock available - Proceeding"

echo "Changing to new cert"
aws --region "$region" s3 cp "s3://control-tower-$deployment-$region-config/terraform.tfstate" terraform.tfstate
db_identifier="$(jq -r '.modules[0].resources."aws_db_instance.default".primary.id' < terraform.tfstate)"
aws rds modify-db-instance --db-instance-identifier "$db_identifier" --ca-certificate-identifier rds-ca-2019 --apply-immediately

echo "Waiting for CA cert to change"
wait_time=0
until [[ $(aws --region "$region" rds describe-db-instances --db-instance-identifier="$db_identifier" | jq -r '.DBInstances[0].PendingModifiedValues') == '{}' ]]; do
  (( ++wait_time ))
  if [[ $wait_time -ge 24 ]]; then
    echo "Waited too long for AWS to effect CA cert change" && exit 1
  fi
  printf '.'
  sleep 5
done
echo "AWS have changed the CA cert - proceeding"

config=$(./cup info --json "$deployment")
# shellcheck disable=SC2034
domain=$(echo "$config" | jq -r '.config.domain')
# shellcheck disable=SC2034
username=$(echo "$config" | jq -r '.config.concourse_username')
# shellcheck disable=SC2034
password=$(echo "$config" | jq -r '.config.concourse_password')
echo "$config" | jq -r '.config.concourse_ca_cert' > generated-ca-cert.pem

# shellcheck disable=SC2034
cert="generated-ca-cert.pem"
# shellcheck disable=SC2034
manifest="$(dirname "$0")/hello.yml"
# shellcheck disable=SC2034
job="hello"

assertPipelineIsSettableAndRunnable
assertNetworkCidrsCorrect
