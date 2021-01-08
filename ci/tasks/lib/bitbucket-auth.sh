#!/bin/bash

# Add flags to an array that should have been initialised previously
function addBitBucketFlagsToArgs() {
  args+=(--bitbucket-auth-client-id "$BITBUCKET_AUTH_CLIENT_ID")
  args+=(--bitbucket-auth-client-secret "$BITBUCKET_AUTH_CLIENT_SECRET")
}

function assertBitBucketAuthConfigured() {
  config=$(./cup info --region "$region" --iaas "$IAAS" --json "$deployment")
  domain=$(echo "$config" | jq -r '.config.domain')
  username=$(echo "$config" | jq -r '.config.concourse_username')
  password=$(echo "$config" | jq -r '.config.concourse_password')

  fly --target system-test login \
    --ca-cert "$cert" \
    --concourse-url "https://$domain" \
    --username "$username" \
    --password "$password"

  echo "Check that bitbucket auth is enabled"
  fly --target system-test set-team \
    --team-name=bitbucket-team \
    --bitbucket-cloud-user=EngineerBetterCI \
    --non-interactive

  ( ( fly --target system-test login --team-name=bitbucket-team 2>&1 ) >fly_out ) &

  sleep 5

  pkill -9 fly

  # Obtains url with spaces and carriage returns removed from fly_out file
  url=$(grep '/login?fly_port=' fly_out | sed 's/[ \r]//g')

  curl -sL --cacert "$cert" "$url" | grep -q '/sky/issuer/auth/bitbucket'

  echo "BitBucket Auth test passed"
}
