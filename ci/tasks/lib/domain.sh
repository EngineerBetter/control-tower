#!/usr/bin/env bash

# Check that concourse is present on given domain
function assertConcoursePresent() {
    curl -ksLo/dev/null --fail https://"$domain"
    printf "Concourse is running at %s\\n" "$domain"
}
