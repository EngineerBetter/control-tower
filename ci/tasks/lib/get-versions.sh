#!/bin/bash
# shellcheck disable=SC2034

function getVersions() {
  bin_bosh_cli_version=$(                jq -r '."bosh-cli".version' createenv-dependencies-and-cli-versions-aws.json)
  bin_terraform_version=$(               jq -r '.terraform.version' createenv-dependencies-and-cli-versions-aws.json)
  concourse_release_url=$(    jq -r '.[] | select(.value.name? == "concourse") | .value.url' ops/versions-aws.json)
  concourse_release_version=$(jq -r '.[] | select(.value.name? == "concourse") | .value.version' ops/versions-aws.json)
  credhub_release_url=$(      jq -r '.[] | select(.value.name? == "credhub") | .value.url' ops/versions-aws.json)
  credhub_release_version=$(  jq -r '.[] | select(.value.name? == "credhub") | .value.version' ops/versions-aws.json)
  grafana_release_url=$(      jq -r '.[] | select(.value.name? == "grafana") | .value.url' ops/versions-aws.json)
  grafana_release_version=$(  jq -r '.[] | select(.value.name? == "grafana") | .value.version' ops/versions-aws.json)
  influxdb_release_url=$(     jq -r '.[] | select(.value.name? == "influxdb") | .value.url' ops/versions-aws.json)
  influxdb_release_version=$( jq -r '.[] | select(.value.name? == "influxdb") | .value.version' ops/versions-aws.json)
  stemcell_version=$(         jq -r '.[] | select(.path == "/stemcells/alias=jammy/version") | .value' ops/versions-aws.json)
  uaa_release_url=$(          jq -r '.[] | select(.value.name? == "uaa") | .value.url' ops/versions-aws.json)
  uaa_release_version=$(      jq -r '.[] | select(.value.name? == "uaa") | .value.version' ops/versions-aws.json)
  director_bosh_cpi_release_url=$(       jq -r .cpi.url createenv-dependencies-and-cli-versions-aws.json)
  director_bosh_cpi_release_version=$(   jq -r .cpi.version createenv-dependencies-and-cli-versions-aws.json)
  director_bosh_release_url=$(           jq -r .bosh.url createenv-dependencies-and-cli-versions-aws.json)
  director_bosh_release_version=$(       jq -r .bosh.version createenv-dependencies-and-cli-versions-aws.json)
  director_bpm_release_url=$(            jq -r .bpm.url createenv-dependencies-and-cli-versions-aws.json)
  director_bpm_release_version=$(        jq -r .bpm.version createenv-dependencies-and-cli-versions-aws.json)
  director_stemcell_version=$(           jq -r .stemcell.version createenv-dependencies-and-cli-versions-aws.json | cut -d= -f2)

  bin_bosh_cli_version_gcp=$(                jq -r '."bosh-cli".version' createenv-dependencies-and-cli-versions-gcp.json)
  bin_terraform_version_gcp=$(               jq -r '.terraform.version' createenv-dependencies-and-cli-versions-gcp.json)
  concourse_release_url_gcp=$(    jq -r '.[] | select(.value.name? == "concourse") | .value.url' ops/versions-gcp.json)
  concourse_release_version_gcp=$(jq -r '.[] | select(.value.name? == "concourse") | .value.version' ops/versions-gcp.json)
  credhub_release_url_gcp=$(      jq -r '.[] | select(.value.name? == "credhub") | .value.url' ops/versions-gcp.json)
  credhub_release_version_gcp=$(  jq -r '.[] | select(.value.name? == "credhub") | .value.version' ops/versions-gcp.json)
  grafana_release_url_gcp=$(      jq -r '.[] | select(.value.name? == "grafana") | .value.url' ops/versions-gcp.json)
  grafana_release_version_gcp=$(  jq -r '.[] | select(.value.name? == "grafana") | .value.version' ops/versions-gcp.json)
  influxdb_release_url_gcp=$(     jq -r '.[] | select(.value.name? == "influxdb") | .value.url' ops/versions-gcp.json)
  influxdb_release_version_gcp=$( jq -r '.[] | select(.value.name? == "influxdb") | .value.version' ops/versions-gcp.json)
  stemcell_version_gcp=$(         jq -r '.[] | select(.path == "/stemcells/alias=jammy/version") | .value' ops/versions-gcp.json)
  uaa_release_url_gcp=$(          jq -r '.[] | select(.value.name? == "uaa") | .value.url' ops/versions-gcp.json)
  uaa_release_version_gcp=$(      jq -r '.[] | select(.value.name? == "uaa") | .value.version' ops/versions-gcp.json)
  director_bosh_cpi_release_url_gcp=$(       jq -r .cpi.url createenv-dependencies-and-cli-versions-gcp.json)
  director_bosh_cpi_release_version_gcp=$(   jq -r .cpi.version createenv-dependencies-and-cli-versions-gcp.json)
  director_bosh_release_url_gcp=$(           jq -r .bosh.url createenv-dependencies-and-cli-versions-gcp.json)
  director_bosh_release_version_gcp=$(       jq -r .bosh.version createenv-dependencies-and-cli-versions-gcp.json)
  director_bpm_release_url_gcp=$(            jq -r .bpm.url createenv-dependencies-and-cli-versions-gcp.json)
  director_bpm_release_version_gcp=$(        jq -r .bpm.version createenv-dependencies-and-cli-versions-gcp.json)
  director_stemcell_version_gcp=$(           jq -r .stemcell.version createenv-dependencies-and-cli-versions-gcp.json | cut -d= -f2)
}
