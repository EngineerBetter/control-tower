# Global Control Tower Flags

These flags can be supplied to all Control Tower commands

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--region value`|AWS or GCP region (default: "eu-west-1" on AWS and "europe-west1" on GCP)|`AWS_REGION`|
|`--namespace value`|Any valid string that provides a meaningful namespace of the deployment - Used as part of the configuration bucket name|`NAMESPACE`|

> If `namespace` or `region` have been provided in the initial `deploy` they will be required for any subsequent `control-tower` calls against the same deployment.

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--iaas value`|IAAS, can be AWS or GCP|`IAAS`|

> `--iaas` is required on every command
