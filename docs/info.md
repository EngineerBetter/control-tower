# Info

To fetch information about your Control Tower deployment in a human readable format:

```sh
control-tower info --iaas [AWS|GCP] <your-project-name>
```

To fetch Information about your Control Tower deployment in a machine parseable format:

```sh
control-tower info --iaas [AWS|GCP] --json <your-project-name>
```

To load credentials into your environment from your Control Tower deployment:

```sh
eval "$(control-tower info --iaas [AWS|GCP] --env <your-project-name>)"
```

To check the expiry of the BOSH Director's NATS CA certificate:

```sh
control-tower info --iaas [AWS|GCP] --cert-expiry <your-project-name>
```

**Warning: if your deployment is approaching a year old, it may stop working due to expired certificates. For information please see this issue https://github.com/EngineerBetter/control-tower/issues/81.**

## Flags

All flags are optional

|**Flag**|**Description**|**Environment Variable**|
|:-|:-|:-|
|`--json`|Output as json|`JSON`
|`--env`|Output environment variables||
|`--cert-expiry`|Output the expiry of the BOSH director's NATS certificate||
