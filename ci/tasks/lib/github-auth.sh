#!/bin/bash

# Add flags to an array that should have been initialised previously
function addGitHubFlagsToArgs() {
  args+=(--github-auth-client-id "$GITHUB_AUTH_CLIENT_ID")
  args+=(--github-auth-client-secret "$GITHUB_AUTH_CLIENT_SECRET")
}

function addMainTeamGitHubAuthFlags() {
  args+=(--main-team-github-users "EngineerBetterCI")
  args+=(--main-team-github-teams "EngineerBetter:fake-team")
  args+=(--main-team-github-orgs "EngineerBetter")
}

function assertGitHubAuthConfigured() {
  config=$(./cup info --region "$region" --iaas "$IAAS" --json "$deployment")
  domain=$(echo "$config" | jq -r '.config.domain')
  username=$(echo "$config" | jq -r '.config.concourse_username')
  password=$(echo "$config" | jq -r '.config.concourse_password')

  fly --target system-test login \
    --ca-cert "$cert" \
    --concourse-url "https://$domain" \
    --username "$username" \
    --password "$password" \
    --team-name main

  echo "Check that github auth is enabled"
  fly --target system-test set-team \
    --team-name=git-team \
    --github-user=EngineerBetterCI \
    --non-interactive

  ( ( fly --target system-test login --team-name=git-team 2>&1 ) >fly_out ) &

  sleep 5

  pkill -9 fly

  # Obtains url with spaces and carriage returns removed from fly_out file
  url=$(grep '/login?fly_port=' fly_out | sed 's/[ \r]//g')

  curl -sL --cacert "$cert" "$url" | grep -q '/sky/issuer/auth/github'

  echo "GitHub Auth test passed"
}

function assertMainTeamGitHubConfigured() {
  config=$(./cup info --region "$region" --iaas "$IAAS" --json "$deployment")
  domain=$(echo "$config" | jq -r '.config.domain')
  username=$(echo "$config" | jq -r '.config.concourse_username')
  password=$(echo "$config" | jq -r '.config.concourse_password')

  fly --target system-test login \
    --ca-cert "$cert" \
    --concourse-url "https://$domain" \
    --username "$username" \
    --password "$password" \
    --team-name main

  main_team_auth=$(fly --target system-test teams --details --json | jq '.[] | select(.name == "main") | .auth.owner')

  echo "Checking Main Team GitHub user auth"
  echo "$main_team_auth" | jq -r '.users' | grep -q '"github:engineerbetterci"'
  echo "Checking Main Team GitHub team auth"
  echo "$main_team_auth" | jq -r '.groups' | grep -q '"github:engineerbetter:fake-team"'
  echo "Checking Main Team GitHub org auth"
  echo "$main_team_auth" | jq -r '.groups' | grep -q '"github:engineerbetter"'

  echo "Main Team GitHub Auth test passed"
}
