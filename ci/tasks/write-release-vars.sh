#!/bin/bash
# shellcheck disable=SC2091,SC2006

# Disabling SC2091 above because we want to print commands encased in $()
# Disabling SC2006 above because ``` code blocks are misinterpretted as shell execution

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/set-flags.sh
source control-tower/ci/tasks/lib/get-versions.sh

version=$(cat version/version)
pushd control-tower-ops
  getVersions
popd

pushd ops-version
  ops_version=$(cat version)
popd

name="control-tower $version"

echo "$name" > release-vars/name

cat << EOF > release-vars/body

Auto-generated release

Deploys:

**AWS**

- Concourse VM stemcell bosh-aws-xen-hvm-ubuntu-xenial-go_agent $deployment_stemcell_version
- Director stemcell     bosh-aws-xen-hvm-ubuntu-xenial-go_agent $director_stemcell_version
- Concourse [$deployment_concourse_release_version]($deployment_concourse_release_url)
- BOSH [$director_bosh_release_version]($director_bosh_release_url)
- BOSH AWS CPI [$director_bosh_cpi_release_version]($director_bosh_cpi_release_url)
- BPM [$director_bpm_release_version]($director_bpm_release_url)
- Credhub [$deployment_credhub_release_version]($deployment_credhub_release_url)
- Grafana [$deployment_grafana_release_version]($deployment_grafana_release_url)
- InfluxDB [$deployment_influxdb_release_version]($deployment_influxdb_release_url)
- UAA [$deployment_uaa_release_version]($deployment_uaa_release_url)
- BOSH CLI $bin_bosh_cli_version
- Terraform $bin_terraform_version

**GCP**

- Concourse VM stemcell bosh-google-kvm-ubuntu-xenial-go_agent $deployment_stemcell_version_gcp
- Director stemcell     bosh-google-kvm-ubuntu-xenial-go_agent $director_stemcell_version_gcp
- Concourse [$deployment_concourse_release_version_gcp]($deployment_concourse_release_url_gcp)
- BOSH [$director_bosh_release_version_gcp]($director_bosh_release_url_gcp)
- BOSH GCP CPI [$director_bosh_cpi_release_version_gcp]($director_bosh_cpi_release_url_gcp)
- BPM [$director_bpm_release_version_gcp]($director_bpm_release_url_gcp)
- Credhub [$deployment_credhub_release_version_gcp]($deployment_credhub_release_url_gcp)
- Grafana [$deployment_grafana_release_version_gcp]($deployment_grafana_release_url_gcp)
- InfluxDB [$deployment_influxdb_release_version_gcp]($deployment_influxdb_release_url_gcp)
- UAA [$deployment_uaa_release_version_gcp]($deployment_uaa_release_url_gcp)
- BOSH CLI $bin_bosh_cli_version_gcp
- Terraform $bin_terraform_version_gcp

>Note to build locally you will need to clone [control-tower-ops](https://github.com/EngineerBetter/control-tower-ops/tree/$ops_version) (version $ops_version) to the same level as control-tower to get the required manifests and ops files.
EOF

pushd control-tower
  commit=$(git rev-parse HEAD)
popd

echo "$commit" > release-vars/commit
