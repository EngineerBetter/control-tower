# Maintain

Maintain is a collection of operations for keeping your Concourse in working order.

## Flags

All flags are optional

### Rotating Director NATS Certificate

|**Flag**|**Description**
|:-|:-|
|`--renew-nats-cert`|Rotate the NATS certificate on the director||
|`--stage value`|Specify a specific stage at which to start the NATS certificate renewal process. If not specified, the stage will be determined automatically.||

> Note that the NATS certificate [is hardcoded to expire after 1 year](https://github.com/cloudfoundry/bosh-cli/blob/master/vendor/github.com/cloudfoundry/config-server/types/certificate_generator.go#L171). This command follows [the istructions on bosh.io](https://bosh.io/docs/nats-ca-rotation/) to rotate this certificate. **This operation _will_ cause downtime on your Concourse** as it performs multiple full recreates.

|Stage|Description|
|:-|:-|
|0|Adding new CA (create-env)|
|1|Recreating VMs for the first time (recreate)|
|2|Removing old CA (create-env)|
|3|Recreating VMs for the second time (recreate)|
|4|Cleaning up director-creds.yml|
