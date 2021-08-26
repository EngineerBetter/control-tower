package resource

import (
	_ "embed"

	"github.com/EngineerBetter/control-tower/resource/internal/file"
)

//go:generate go-bindata -o internal/file/file.go -ignore (\.go$)|(\.git)|(bosh/assets) -nometadata -pkg file ../../control-tower-ops/... assets/...

var (
	// DirectorManifest statically defines director-manifest.yml contents
	//go:embed assets/manifest.yml
	DirectorManifest string

	// AWSDirectorCloudConfig statically defines aws cloud-config.yml
	//go:embed assets/aws/cloud-config.yml
	AWSDirectorCloudConfig string

	// AWSCPIOps statically defines aws-cpi.yml contents
	//go:embed assets/aws/cpi.yml
	AWSCPIOps string

	// ExternalIPOps statically defines external-ip.yml contents
	//go:embed assets/aws/external-ip.yml
	AWSExternalIPOps string

	// AWSDirectorCustomOps statically defines custom-ops.yml contents
	//go:embed assets/aws/custom-ops.yml
	AWSDirectorCustomOps string

	// AWSBlobstoreOps defines s3-blobstore-ops.yml contents
	//go:embed assets/aws/s3-blobstore-ops.yml
	AWSBlobstoreOps string

	// GCPDirectorCloudConfig statically defines gcp cloud-config.yml
	//go:embed assets/gcp/cloud-config.yml
	GCPDirectorCloudConfig string

	// GCPCPIOps statically defines gcp-cpi.yml contents
	//go:embed assets/gcp/cpi.yml
	GCPCPIOps string

	// GCPExternalIPOps statically defines external-ip.yml contents
	//go:embed assets/gcp/external-ip.yml
	GCPExternalIPOps string

	// GCPDirectorCustomOps statically defines custom-ops.yml contents
	//go:embed assets/gcp/custom-ops.yml
	GCPDirectorCustomOps string

	//GCPJumpboxUserOps statically defines gcp jumpbox-user.yml
	//go:embed assets/gcp/jumpbox-user.yml
	GCPJumpboxUserOps string

	// AWSTerraformConfig holds the terraform conf for AWS
	//go:embed assets/aws/infrastructure.tf
	AWSTerraformConfig string

	// GCPTerraformConfig holds the terraform conf for GCP
	//go:embed assets/gcp/infrastructure.tf
	GCPTerraformConfig string

	// AWSReleaseVersions carries all versions of releases
	AWSReleaseVersions = file.MustAssetString("../../control-tower-ops/ops/versions-aws.json")

	// GCPReleaseVersions carries all versions of releases
	GCPReleaseVersions = file.MustAssetString("../../control-tower-ops/ops/versions-gcp.json")

	// AddNewCa carries the ops file that adds a new CA required for cert rotation
	//go:embed assets/maintenance/add-new-ca.yml
	AddNewCa string

	// RemoveOldCa carries the ops file that removes the old CA required for cert rotation
	//go:embed assets/maintenance/remove-old-ca.yml
	RemoveOldCa string

	// CleanupCerts moves renewed values of certs to old keys in director vars store
	//go:embed assets/maintenance/cleanup-certs.yml
	CleanupCerts string

	AWSVersionFile = file.MustAsset("../../control-tower-ops/createenv-dependencies-and-cli-versions-aws.json")

	GCPVersionFile = file.MustAsset("../../control-tower-ops/createenv-dependencies-and-cli-versions-gcp.json")
)
