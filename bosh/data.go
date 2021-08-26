package bosh

import _ "embed"

const (
	credsFilename                     = "concourse-creds.yml"
	concourseManifestFilename         = "concourse.yml"
	concourseDeploymentName           = "concourse"
	concourseVersionsFilename         = "versions.json"
	concourseSHAsFilename             = "shas.json"
	concourseGrafanaFilename          = "grafana_dashboard.yml"
	concourseBitBucketAuthFilename    = "bitbucket-auth.yml"
	concourseGitHubAuthFilename       = "github-auth.yml"
	concourseMicrosoftAuthFilename    = "microsoft-auth.yml"
	concourseEphemeralWorkersFilename = "ephemeral_workers.yml"
	extraTagsFilename                 = "extra_tags.yml"
	uaaCertFilename                   = "uaa-cert.yml"
)

//go:generate go-bindata -pkg $GOPACKAGE -ignore \.git assets/... ../../control-tower-ops/... ../resource/assets/...

var (
	//go:embed assets/grafana_dashboard.yml
	concourseGrafana []byte

	//go:embed assets/ops/bitbucket-auth.yml
	concourseBitBucketAuth []byte

	//go:embed assets/ops/github-auth.yml
	concourseGitHubAuth []byte

	//go:embed assets/ops/microsoft-auth.yml
	concourseMicrosoftAuth []byte

	//go:embed assets/ops/ephemeral_workers.yml
	concourseEphemeralWorkers []byte

	//go:embed assets/ops/extra_tags.yml
	extraTags []byte

	concourseManifestContents = MustAsset("../../control-tower-ops/manifest.yml")
	awsConcourseVersions      = MustAsset("../../control-tower-ops/ops/versions-aws.json")
	awsConcourseSHAs          = MustAsset("../../control-tower-ops/ops/shas-aws.json")
	gcpConcourseVersions      = MustAsset("../../control-tower-ops/ops/versions-gcp.json")
	gcpConcourseSHAs          = MustAsset("../../control-tower-ops/ops/shas-gcp.json")
	uaaCert                   = MustAsset("../resource/assets/gcp/uaa-cert.yml")
)
