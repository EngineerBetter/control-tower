# Credential Management

Control Tower deploys the [credhub](https://github.com/cloudfoundry-incubator/credhub) service alongside Concourse and configures Concourse to use it. More detail on how credhub integrates with Concourse can be found [in the Concourse docs](https://concourse-ci.org/creds.html).

You can log into credhub by running:

```sh
eval "$(control-tower info --iaas [AWS|GCP] --env --region $region $deployment)"
```
