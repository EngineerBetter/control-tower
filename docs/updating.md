# Updating your Concourse

## Self-update

When Control Tower deploys Concourse, it now adds a pipeline to the new Concourse called `control-tower-self-update`. This pipeline continuously monitors our Github repo for new releases and updates Concourse in place whenever a new version of Control Tower comes out.

This pipeline is paused by default, so just unpause it in the UI to enable the feature.

## Upgrading manually

Patch releases of `control-tower` are compiled, tested and released automatically whenever a new stemcell or component release appears on [bosh.io](https://bosh.io).

To upgrade your Concourse, grab the [latest release](https://github.com/EngineerBetter/control-tower/releases/latest) and run `control-tower deploy --iaas [AWS|GCP] <your-project-name>` again.
