# Troubleshooting

If you can't see your issue on this page come ask us about it on our [Community Slack](https://join.slack.com/t/concourse-up/shared_invite/enQtNDMzNjY1MjczNDU3LWVkZDllYjE0NTI2M2NkMjM5ZWY0NGM1MzM2N2VhYzgxN2NkM2I0ZDdiOGUxMjRkZjg3ZGQwOWIwNTNjMmU3OTg) or [create an issue](https://github.com/EngineerBetter/control-tower/issues).

## NATS Certificate is Expired

[NATS](https://bosh.io/docs/bosh-components/#nats) handles communication between the director VM and the bosh-agent processes that run on each VM that it manages (web and worker(s)). When it expires this communication is no longer possible and any running VMs will appear as `unresponsive agent` in `bosh vms`.

You can check the expiry of the NATS certs on your Control Tower deployment with:

```sh
control-tower info --iaas <AWS|GCP> --region <region> --cert-expiry <deployment-name>
```

and if it is getting close to expiry you can rotate it with [the maintain command](maintain.md#rotating-director-nats-certificate).

If the certificate has already expired you will see an error when deploying which resembles:

```sh
Deploying:
  Creating instance 'bosh/0':
    Waiting until instance is ready:
      Post https://mbus:<redacted>@<IP>:6868/agent: x509: certificate has expired or is not yet valid
Exit code 1
```

Solution:

1. Download `director-creds.yml` from the config bucket of your deployment (in S3 or GCS depending on your IAAS)
1. Delete all the certs in that file ([more info](https://github.com/cloudfoundry/bosh-deployment/issues/396#issuecomment-668962407))

    > Note that each certificate will contain keys for `ca`, `private_key`, and `certificate`. You need to delete all three keys for each certificate

1. Overwrite the `director-creds.yml` in your bucket with your newly modified one
1. Run `control-tower deploy` to force BOSH to generate new certs

    > Note that the Concourse deploy will fail and all the VMs will appear in BOSH as `unresponsive agent`

1. Export bosh credentials with `eval "$(control-tower info --iaas [AWS|GCP] --env <deployment-name>)"`
1. Run `bosh deploy --recreate --fix <(bosh manifest)` to push the new NATs cert to each vm
1. Run `control-tower deploy` which should now run all the way through
1. Optionally run the `renew-https-cert` job in the `control-tower-self-update` pipeline in your main team to renew the outward facing SSL cert

Further information can be found in [the BOSH docs](https://bosh.io/docs/nats-ca-rotation/#expired).


## BOSH Director certificate has expired

If the certificate (the Director API endpoint) has expired then you'll see the following error when interacting with `control-tower` which remsembles:

```sh
Succeeded
Fetching info:
  Performing request GET 'https://<redacted>:25555/info':
    Performing GET request:
      Retry: Get https://<redacted>:25555/info: x509: certificate has expired or is not yet valid

Exit code 1
exit status 1
```

You can check the certificate expiry dates using the following command:

```sh
echo | openssl s_client -showcerts -connect <director-ip>:25555 | openssl x509 -noout -text
```

Solution:


1. Download `config.json` from the config bucket of your deployment (in S3 or GCS depending on your IAAS), whose name *should* resemble `control-tower-<deployment>-<region>-config`
1. Delete the `director_ca_cert`, `director_cert` and `director_key` from the `config.json` file.
1. Overwrite the `config.json` in your bucket with your newly modified one
1. Run `control-tower deploy` to force BOSH to generate new certs:

*e.g*
```sh
$ control-tower deploy --iaas <AWS or GCP> --region <region> <deployment>
```

Once the certificate has been regenerated and deployed, you can check with the following command:

```sh
echo | openssl s_client -showcerts -connect <director-ip>:25555 | openssl x509 -noout -text
```
