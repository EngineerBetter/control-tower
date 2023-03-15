#!/bin/bash

# shellcheck disable=SC1091
# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/test-setup.sh

handleVerboseMode
setDeploymentName region

# shellcheck disable=SC2034
region=eu-west-1
trapCustomCleanup

cp "$BINARY_PATH" ./cup
chmod +x ./cup

echo "DEPLOY WITH ENCRYPTED DISK"
./cup deploy --rds-disk-encryption --region "$region" "$deployment"

sleep 60

config=$(./cup info --region "$region" --json "$deployment")
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

# Download the right version of fly from Concourse UI
updateFly "${domain}"

assertPipelineIsSettableAndRunnable

echo "DEPLOY WITH ENCRYPTED DISK"
./cup deploy --rds-disk-encryption --region "$region" "$deployment"

sleep 30

aws --region "$region" s3 cp "s3://control-tower-$deployment-$region-config/terraform.tfstate" terraform.tfstate
db_identifier="$(jq -r '.resources[] | select( .type == "aws_db_instance") | .instances[0].attributes.id' < terraform.tfstate)"

storageEncypted="$(aws rds describe-db-instances --region eu-west-1 --db-instance-identifier "$db_identifier" --output json | jq -r ".DBInstances[0].StorageEncrypted")"
if [ "$storageEncypted" != "true" ]; then
  echo "RDS Disk ${db_identifier} not encrypted, StorageEncrypted is set to ${db_identifier}"
  exit 1
fi

echo "RDS Disk ${db_identifier} encrypted, StorageEncrypted is set to ${db_identifier}"

assertPipelineIsSettableAndRunnable
