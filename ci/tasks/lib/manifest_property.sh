#!/bin/bash

# Takes two arguments- path and value
# Checks if path=value in the Bosh manifest for 'concourse' deployment
function checkManifestProperty() {
  # Checks if exactly two arguments were passed to the function
  # shellcheck disable=SC2086
  : ${2?"Usage: manifestProperty PATH VALUE"}
  path="$1"
  expected_value="$2"
  actual_value=$(bosh int --path "${path}" <(bosh manifest -d concourse))
  if [ "$actual_value" != "$expected_value" ]; then
    echo "Error: Wants '${path}: ${expected_value}', got '${path}: ${actual_value}'"
    exit 1
  fi
}