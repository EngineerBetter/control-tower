#!/bin/bash

# shellcheck disable=SC1091
source control-tower/ci/tasks/lib/get-versions.sh
source control-tower/ci/tasks/lib/set-flags.sh

pushd control-tower-ops
  getVersions
popd

jq --null-input \
  --arg bin_bosh_cli_version "${bin_bosh_cli_version}" \
  --arg bin_terraform_version "${bin_terraform_version}" \
  --arg concourse_release_url "${concourse_release_url}" \
  --arg concourse_release_version "${concourse_release_version}" \
  --arg credhub_release_url "${credhub_release_url}" \
  --arg credhub_release_version "${credhub_release_version}" \
  --arg grafana_release_url "${grafana_release_url}" \
  --arg grafana_release_version "${grafana_release_version}" \
  --arg influxdb_release_url "${influxdb_release_url}" \
  --arg influxdb_release_version "${influxdb_release_version}" \
  --arg stemcell_version "${stemcell_version}" \
  --arg uaa_release_url "${uaa_release_url}" \
  --arg uaa_release_version "${uaa_release_version}" \
  --arg director_bosh_cpi_release_url "${director_bosh_cpi_release_url}" \
  --arg director_bosh_cpi_release_version "${director_bosh_cpi_release_version}" \
  --arg director_bosh_release_url "${director_bosh_release_url}" \
  --arg director_bosh_release_version "${director_bosh_release_version}" \
  --arg director_bpm_release_url "${director_bpm_release_url}" \
  --arg director_bpm_release_version "${director_bpm_release_version}" \
  --arg director_stemcell_version "${director_stemcell_version}" \
  --arg bin_bosh_cli_version_gcp "${bin_bosh_cli_version_gcp}" \
  --arg bin_terraform_version_gcp "${bin_terraform_version_gcp}" \
  --arg concourse_release_url_gcp "${concourse_release_url_gcp}" \
  --arg concourse_release_version_gcp "${concourse_release_version_gcp}" \
  --arg credhub_release_url_gcp "${credhub_release_url_gcp}" \
  --arg credhub_release_version_gcp "${credhub_release_version_gcp}" \
  --arg grafana_release_url_gcp "${grafana_release_url_gcp}" \
  --arg grafana_release_version_gcp "${grafana_release_version_gcp}" \
  --arg influxdb_release_url_gcp "${influxdb_release_url_gcp}" \
  --arg influxdb_release_version_gcp "${influxdb_release_version_gcp}" \
  --arg stemcell_version_gcp "${stemcell_version_gcp}" \
  --arg uaa_release_url_gcp "${uaa_release_url_gcp}" \
  --arg uaa_release_version_gcp "${uaa_release_version_gcp}" \
  --arg director_bosh_cpi_release_url_gcp "${director_bosh_cpi_release_url_gcp}" \
  --arg director_bosh_cpi_release_version_gcp "${director_bosh_cpi_release_version_gcp}" \
  --arg director_bosh_release_url_gcp "${director_bosh_release_url_gcp}" \
  --arg director_bosh_release_version_gcp "${director_bosh_release_version_gcp}" \
  --arg director_bpm_release_url_gcp "${director_bpm_release_url_gcp}" \
  --arg director_bpm_release_version_gcp "${director_bpm_release_version_gcp}" \
  --arg director_stemcell_version_gcp "${director_stemcell_version_gcp}" \
  '
    [
      {
        name: "bin_bosh_cli",
        version: $bin_bosh_cli_version
      },
      {
        name: "bin_terraform",
        version: $bin_terraform_version
      },
      {
        name: "concourse_release",
        url: $concourse_release_url,
        version: $concourse_release_version
      },
      {
        name: "credhub_release",
        url: $credhub_release_url,
        version: $credhub_release_version
      },
      {
        name: "grafana_release",
        url: $grafana_release_url,
        version: $grafana_release_version
      },
      {
        name: "influxdb_release",
        url: $influxdb_release_url,
        version: $influxdb_release_version
      },
      {
        name: "stemcell_aws",
        version: $stemcell_version
      },
      {
        name: "stemcell_gcp",
        version: $stemcell_version_gcp
      },
      {
        name: "uaa_release",
        url: $uaa_release_url,
        version: $uaa_release_version
      },
      {
        name: "director_bosh_cpi_release_aws",
        url: $director_bosh_cpi_release_url,
        version: $director_bosh_cpi_release_version
      },
      {
        name: "director_bosh_cpi_release_gcp",
        url: $director_bosh_cpi_release_url_gcp,
        version: $director_bosh_cpi_release_version_gcp
      },
      {
        name: "director_bosh_release",
        url: $director_bosh_release_url,
        version: $director_bosh_release_version
      },
      {
        name: "director_bpm_release",
        url: $director_bpm_release_url,
        version: $director_bpm_release_version
      },
      {
        name: "director_stemcell_aws",
        version: $director_stemcell_version
      },
      {
        name: "director_stemcell_gcp",
        version: $director_stemcell_version_gcp
      }
    ]
  ' > versions-file/release-versions.json

  cat versions-file/release-versions.json