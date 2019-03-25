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
