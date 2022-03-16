# Troubleshooting

If you can't see your issue on this page come ask us about it on our [Community Slack](https://join.slack.com/t/concourse-up/shared_invite/enQtNDMzNjY1MjczNDU3LWVkZDllYjE0NTI2M2NkMjM5ZWY0NGM1MzM2N2VhYzgxN2NkM2I0ZDdiOGUxMjRkZjg3ZGQwOWIwNTNjMmU3OTg) or [create an issue](https://github.com/EngineerBetter/control-tower/issues).

## General BOSH Debugging

Control Tower uses [BOSH](https://bosh.io/docs/) to deploy and manage VMs. When something isn't working right but the cause isn't obvious the best general first steps are:

1. Export bosh credentials

    ```sh
    eval "$(control-tower info --iaas [AWS|GCP] --env <deployment-name>)"
    ```

1. Check the status of the deployed VMs

    ```sh
    bosh vms
    ```

    Which gives an output like:

    ```txt
    Using environment '1.2.3.4' as client 'admin'

    Task 9312. Done

    Deployment 'concourse'

    Instance                                     Process State  AZ  IPs          VM CID               VM Type               Active  Stemcell
    web/95589e21-09af-412d-abef-a2065fa828fe     running        z1  1.2.3.4      i-00000000000000000  concourse-web-xlarge  true    bosh-aws-xen-hvm-ubuntu-bionic-go_agent/1.67
                                                                    10.0.0.8
    worker/17cedb77-a924-4e09-bb1a-952b7e8b3fc6  failing        z1  10.0.1.7     i-00000000000000000  concourse-2xlarge     true    bosh-aws-xen-hvm-ubuntu-bionic-go_agent/1.67

    2 vms

    Succeeded
    ```

    Look for vms that aren't in the `running` state.

1. If you see a VM in a non-running state you can ssh to it with

    ```sh
    bosh ssh worker/17cedb77-a924-4e09-bb1a-952b7e8b3fc6
    ```

1. Once on the VM become root (you can't do much without it) and check the state of all the processes (BOSH uses `monit` to manage processes)

    ```sh
    sudo -i
    monit summary
    ```

    ```txt
    The Monit daemon 5.2.5 uptime: 5d 12h 12m

    Process 'worker'                    running
    Process 'telegraf'                  failing
    Process 'telegraf-agent'            running
    System 'system_a85416b7-ea5c-4889-bed1-20ce16cef76e' running
    ```

1. If you see a process that is erroring you can find logs for it (and all other processes) in `/var/vcap/sys/log/<process-name>`

1. You can manually restart processes with `monit restart <process-name>`

Some other useful bosh commands are:

* `bosh tasks --all --recent` shows recent tasks including system ones - can show if a VM is flapping and BOSH keeps trying to restart it

You can read more about BOSH troubleshooting in [their own documentation](https://bosh.io/docs/tips/).

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

e.g.

```sh
control-tower deploy --iaas <AWS or GCP> --region <region> <deployment>
```

Once the certificate has been regenerated and deployed, you can check with the following command:

```sh
echo | openssl s_client -showcerts -connect <director-ip>:25555 | openssl x509 -noout -text
```
