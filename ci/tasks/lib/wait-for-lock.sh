#!/bin/bash

function waitForBoshLock() {
  for i in {1..60}; do
    bosh_locks=$(bosh locks --json | jq '.Tables[0].Rows[] | select(.resource=="concourse")') 
    if [[ -n "${bosh_locks}" ]]; then
        echo "Attempt ${i}: Concourse still has a BOSH lock, waiting another 60 seconds..."
        sleep 60
    else
        break
    fi
  done  
}
