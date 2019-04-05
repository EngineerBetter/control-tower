# Control Tower Prerequisites

## Preparing your local environment

Ensure you have the [dependencies required by `bosh-cli`](https://bosh.io/docs/cli-v2-install/#additional-dependencies) installed in your local environment.

> Under the hood Control Tower uses BOSH for creating and managing VMs. Most compilation will occur on VMs but the dependencies for the BOSH director itself must be compiled in your local environment.

## Setting up credentials

Control Tower requires credentials to your IaaS in order to deploy Concourse.

**Ensure your credentials are *long lived credentials* and not *temporary security credentials***

### AWS

One of

- The environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` are set.
- Credentials for the default profile in `~/.aws/credentials` are present.
- Credentials for a profile in `~/.aws/credentials` are present and `AWS_DEFAULT_PROFILE` is set.

#### Using a dedicated AWS IAM account

If you'd like to run control-tower with it's own IAM account, create a user with the following permissions:

![Required IAM policies](http://i.imgur.com/Q0mOUjv.png)

### GCP

- The environment variable `GOOGLE_APPLICATION_CREDENTIALS_CONTENTS` set to the path to a GCP credentials json file

On GCP you must also ensure the following APIs are activated in your project:

- Compute Engine API (`gcloud services enable compute.googleapis.com`)
- Identity and Access Management (IAM) API (`gcloud services enable iam.googleapis.com`)
- Cloud Resource Manager API (`gcloud services enable cloudresourcemanager.googleapis.com`)
- SQL Admin API (`gcloud services enable sqladmin.googleapis.com`)

#### Using a dedicated GCP IAM member

A IAM Primitive role of `roles/owner` for the target GCP Project is required
