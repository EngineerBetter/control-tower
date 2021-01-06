#!/usr/bin/env bash

function assertConfigBucketVersioned() {
  bucket_name="control-tower-$deployment-$region-config"

  if [ "$IAAS" = "AWS" ]; then
    if ! aws s3api get-bucket-versioning --bucket "$bucket_name" | grep -q Enabled; then
        echo S3 bucket "$bucket_name" did not have versioning enabled
        exit 1
    fi
  elif [ "$IAAS" = "GCP" ]; then
    if ! gsutil versioning get gs://"$bucket_name" | grep -q Enabled; then
        echo GCS bucket "$bucket_name" did not have versioning enabled
        exit 1
    fi
  else
    echo "Unknown iaas: $IAAS"
    exit 1
  fi

  echo "Config bucket $bucket_name had versioning enabled"
}

func assertBucketRegion() {
  bucket_name="control-tower-$deployment-$region-config"
  if [ "$IAAS" = "AWS" ]; then
    bucket_region=$(aws s3api get-bucket-location --bucket  "${bucket_name}" | jq  --arg region "${region}" -r '.LocationConstraint')
  fi
  elif [ "$IAAS" = "GCP" ]; then
    bucket_region=$(gsutil ls -L -b "gs://${bucket_name}" | awk '/Location constraint/ {print tolower($3)}')
  fi
  if [[ "$bucket_region" != "${region}" ]]; then
    echo "Error: bucket ${bucket_name} should be in ${region}, but was created in ${bucket_region}"
    exit 1
  fi
}
