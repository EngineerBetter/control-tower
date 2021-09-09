#!/bin/bash
set -euo pipefail

# shellcheck disable=SC2091,SC2006

# Disabling SC2091 above because we want to print commands encased in $()
# Disabling SC2006 above because ``` code blocks are misinterpretted as shell execution

# shellcheck disable=SC1091

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

new_release=$(jq --compact-output '.[]' new-versions/release-versions.json)
old_release=$(cat old-versions/release-versions.json)

cat << EOF > release-vars/body

Auto-generated release

EOF

for component in $new_release; do
  name=$(echo "$component" | jq --raw-output '.name')
  new_version=$(echo "$component" | jq --raw-output '.version')
  old_version=$(echo "$old_release" | jq --raw-output --arg name "$name" '.[] | select(.name==$name).version')

  if [ "$(printf '%s\n' "$new_version" "$old_version" | sort -V | head -n1)" != "$new_version" ]; then
    echo "$name: $old_version > $new_version" >> release-vars/body
  fi
done

cat << EOF >> release-vars/body

Deploys:

**AWS**

- Concourse VM stemcell bosh-aws-xen-hvm-ubuntu-bionic-go_agent $stemcell_version
- Director stemcell bosh-aws-xen-hvm-ubuntu-bionic-go_agent $director_stemcell_version
- Concourse [$concourse_release_version]($concourse_release_url)
- BOSH [$director_bosh_release_version]($director_bosh_release_url)
- BOSH AWS CPI [$director_bosh_cpi_release_version]($director_bosh_cpi_release_url)
- BPM [$director_bpm_release_version]($director_bpm_release_url)
- Credhub [$credhub_release_version]($credhub_release_url)
- Grafana [$grafana_release_version]($grafana_release_url)
- InfluxDB [$influxdb_release_version]($influxdb_release_url)
- UAA [$uaa_release_version]($uaa_release_url)
- BOSH CLI $bin_bosh_cli_version
- Terraform $bin_terraform_version

**GCP**

- Concourse VM stemcell bosh-google-kvm-ubuntu-bionic-go_agent $stemcell_version_gcp
- Director stemcell bosh-google-kvm-ubuntu-bionic-go_agent $director_stemcell_version_gcp
- Concourse [$concourse_release_version_gcp]($concourse_release_url_gcp)
- BOSH [$director_bosh_release_version_gcp]($director_bosh_release_url_gcp)
- BOSH GCP CPI [$director_bosh_cpi_release_version_gcp]($director_bosh_cpi_release_url_gcp)
- BPM [$director_bpm_release_version_gcp]($director_bpm_release_url_gcp)
- Credhub [$credhub_release_version_gcp]($credhub_release_url_gcp)
- Grafana [$grafana_release_version_gcp]($grafana_release_url_gcp)
- InfluxDB [$influxdb_release_version_gcp]($influxdb_release_url_gcp)
- UAA [$uaa_release_version_gcp]($uaa_release_url_gcp)
- BOSH CLI $bin_bosh_cli_version_gcp
- Terraform $bin_terraform_version_gcp

>Note to build locally you will need to clone [control-tower-ops](https://github.com/EngineerBetter/control-tower-ops/tree/$ops_version) (version $ops_version) to the same level as control-tower to get the required manifests and ops files.
EOF

pushd control-tower
  commit=$(git rev-parse HEAD)
popd

echo "$commit" > release-vars/commit
