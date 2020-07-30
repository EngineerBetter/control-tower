#!/usr/bin/env bash
# shellcheck disable=SC2006

function configureWhitelist() {
  host_ip="$(dig +short myip.opendns.com @resolver1.opendns.com)/32"

  if [ "$IAAS" = "AWS" ]; then
    group_id=$(aws ec2 --region "${region}" describe-security-groups --filters "Name=group-name,Values=control-tower-${deployment}-director" | jq -r '.SecurityGroups[0].GroupId')

    permissions=$(ruby -e "$(cat << EOL
require 'json'
permissions = `aws ec2 --region "${region}" describe-security-groups --group-id "${group_id}" --query "SecurityGroups[0].IpPermissions"`
permissions_edited = JSON.parse(permissions.to_json).tap do |rules|
  rules.each do |rule|
    rule['IpRanges'].delete_if { |h| h['CidrIp'] != "${host_ip}" }
  end
end

puts JSON.dump(permissions_edited)
EOL
    )"
    )

    echo "Removing individual ingress rules for host's IP"
    aws ec2 --region "${region}" revoke-security-group-ingress --cli-input-json "{\"GroupId\": \"$group_id\", \"IpPermissions\": $permissions}"

    echo "Adding non-host ingress rule allowing all traffic"
    aws ec2 --region "${region}" authorize-security-group-ingress --group-id "${group_id}" --ip-permissions IpProtocol=-1,IpRanges="[{CidrIp=192.168.1.1/32}]"

    echo "Adding host ingress rule allowing all traffic"
    aws ec2 --region "${region}" authorize-security-group-ingress --group-id "${group_id}" --ip-permissions IpProtocol=-1,IpRanges="[{CidrIp=${host_ip}}]"

  elif [ "$IAAS" = "GCP" ]; then
    echo "Nothing to do on GCP"
  fi
}
