#!/bin/bash

function updateFly() {
  domain=$1
  if [ -z "$domain" ]; then
    echo "Error: cannot download fly. No Concourse IP or domain name provided"
    exit 1
  fi
flyPath=$(which fly)
rm -rf "${flyPath}"
curl -o fly -L "https://${domain}/api/v1/cli?arch=amd64&platform=linux"
mv fly "${flyPath}"
chmod +x "${flyPath}"
}