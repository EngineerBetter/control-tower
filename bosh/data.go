package bosh

import (
	_ "embed"

	"github.com/EngineerBetter/control-tower/opsassets"
	"github.com/EngineerBetter/control-tower/resource"
)

const (
	credsFilename                     = "concourse-creds.yml"
	concourseManifestFilename         = "concourse.yml"
	concourseDeploymentName           = "concourse"
	concourseVersionsFilename         = "versions.json"
	concourseSHAsFilename             = "shas.json"
	concourseGrafanaFilename          = "grafana_dashboard.yml"
	concourseBitBucketAuthFilename    = "bitbucket-auth.yml"
	concourseGitHubAuthFilename       = "github-auth.yml"
	concourseMainGitHubAuthFilename   = "main-github-auth.yml"
	concourseMicrosoftAuthFilename    = "microsoft-auth.yml"
	concourseEphemeralWorkersFilename = "ephemeral_workers.yml"
	concourseNoMetricsFilename        = "no_metrics.yml"
	extraTagsFilename                 = "extra_tags.yml"
	uaaCertFilename                   = "uaa-cert.yml"
)

var (
	//go:embed assets/grafana_dashboard.yml
	concourseGrafana []byte

	//go:embed assets/ops/bitbucket-auth.yml
	concourseBitBucketAuth []byte

	//go:embed assets/ops/github-auth.yml
	concourseGitHubAuth []byte

	//go:embed assets/ops/main-github-auth.yml
	concourseMainGitHubAuth []byte

	//go:embed assets/ops/microsoft-auth.yml
	concourseMicrosoftAuth []byte

	//go:embed assets/ops/ephemeral_workers.yml
	concourseEphemeralWorkers []byte

	//go:embed assets/ops/no_metrics.yml
	concourseNoMetrics []byte

	//go:embed assets/ops/extra_tags.yml
	extraTags []byte

	concourseManifestContents = opsassets.ConcourseManifestContents
	awsConcourseVersions      = opsassets.AwsConcourseVersions
	awsConcourseSHAs          = opsassets.AwsConcourseSHAs
	gcpConcourseVersions      = opsassets.GcpConcourseVersions
	gcpConcourseSHAs          = opsassets.GcpConcourseSHAs
	uaaCert                   = resource.UAACert
)
