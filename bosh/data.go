package bosh

const concourseManifestFilename = "concourse.yml"
const credsFilename = "concourse-creds.yml"
const concourseDeploymentName = "concourse"
const concourseVersionsFilename = "versions.json"
const concourseSHAsFilename = "shas.json"
const concourseGrafanaFilename = "grafana_dashboard.yml"
const concourseGitHubAuthFilename = "github-auth.yml"
const extraTagsFilename = "extra_tags.yml"
const uaaCertFilename = "uaa-cert.yml"

//go:generate go-bindata -pkg $GOPACKAGE -ignore \.git assets/... ../../control-tower-ops/... ../resource/assets/...
var concourseGrafana = MustAsset("assets/grafana_dashboard.yml")
var concourseGitHubAuth = MustAsset("assets/ops/github-auth.yml")
var extraTags = MustAsset("assets/ops/extra_tags.yml")
var concourseManifestContents = MustAsset("../../control-tower-ops/manifest.yml")
var awsConcourseVersions = MustAsset("../../control-tower-ops/ops/versions-aws.json")
var awsConcourseSHAs = MustAsset("../../control-tower-ops/ops/shas-aws.json")
var gcpConcourseVersions = MustAsset("../../control-tower-ops/ops/versions-gcp.json")
var gcpConcourseSHAs = MustAsset("../../control-tower-ops/ops/shas-gcp.json")
var uaaCert = MustAsset("../resource/assets/gcp/uaa-cert.yml")
