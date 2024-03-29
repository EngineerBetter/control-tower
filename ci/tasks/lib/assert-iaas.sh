#!/bin/bash

if [ "$IAAS" = "AWS" ]; then
    [[ -n "$AWS_ACCESS_KEY_ID" ]]
    [[ -n "$AWS_SECRET_ACCESS_KEY" ]]
    # shellcheck disable=SC2034
    region=eu-west-1
    # https://docs.aws.amazon.com/cli/latest/userguide/cli-usage-pagination.html#cli-usage-pagination-clientside
    export AWS_PAGER=""
elif [ "$IAAS" = "GCP" ]; then
    [[ -n "$GOOGLE_APPLICATION_CREDENTIALS_CONTENTS" ]]
    # shellcheck disable=SC1091
    source control-tower/ci/tasks/lib/gcreds.sh
    setGoogleCreds
    # shellcheck disable=SC2034
    region=europe-west1

    gcloud auth activate-service-account --key-file="$GOOGLE_APPLICATION_CREDENTIALS"
    export CLOUDSDK_CORE_PROJECT=control-tower-233017
fi
