package opsassets

import _ "embed"

var (
	//go:embed assets/manifest.yml
	ConcourseManifestContents []byte

	//go:embed assets/ops/versions-aws.json
	AwsConcourseVersions []byte

	//go:embed assets/ops/shas-aws.json
	AwsConcourseSHAs []byte

	//go:embed assets/ops/versions-gcp.json
	GcpConcourseVersions []byte

	//go:embed assets/ops/shas-gcp.json
	GcpConcourseSHAs []byte

	//go:embed assets/createenv-dependencies-and-cli-versions-aws.json
	AWSVersionFile []byte

	//go:embed assets/createenv-dependencies-and-cli-versions-gcp.json
	GCPVersionFile []byte
)
