package resource

import (
	"github.com/EngineerBetter/control-tower/resource/internal/file"
)

//go:generate go-bindata -o internal/file/file.go -ignore (\.go$)|(\.git)|(bosh/assets) -nometadata -pkg file ../../control-tower-ops/... assets/...

var (
	// DirectorManifest statically defines director-manifest.yml contents
	DirectorManifest = file.MustAssetString("assets/manifest.yml")
	// AWSDirectorCloudConfig statically defines aws cloud-config.yml
	AWSDirectorCloudConfig = file.MustAssetString("assets/aws/cloud-config.yml")
	// AWSCPIOps statically defines aws-cpi.yml contents
	AWSCPIOps = file.MustAssetString("assets/aws/cpi.yml")
	//GCPJumpboxUserOps statically defines gcp jumpbox-user.yml
	GCPJumpboxUserOps = file.MustAssetString("assets/gcp/jumpbox-user.yml")
	// GCPDirectorCloudConfig statically defines gcp cloud-config.yml
	GCPDirectorCloudConfig = file.MustAssetString("assets/gcp/cloud-config.yml")
	// GCPCPIOps statically defines gcp-cpi.yml contents
	GCPCPIOps = file.MustAssetString("assets/gcp/cpi.yml")
	// GCPExternalIPOps statically defines external-ip.yml contents
	GCPExternalIPOps = file.MustAssetString("assets/gcp/external-ip.yml")
	// GCPDirectorCustomOps statically defines custom-ops.yml contents
	GCPDirectorCustomOps = file.MustAssetString("assets/gcp/custom-ops.yml")

	// AWSTerraformConfig holds the terraform conf for AWS
	AWSTerraformConfig = file.MustAssetString("assets/aws/infrastructure.tf")

	// GCPTerraformConfig holds the terraform conf for GCP
	GCPTerraformConfig = file.MustAssetString("assets/gcp/infrastructure.tf")

	// ExternalIPOps statically defines external-ip.yml contents
	ExternalIPOps = file.MustAssetString("assets/external-ip.yml")
	// AWSDirectorCustomOps statically defines custom-ops.yml contents
	AWSDirectorCustomOps = file.MustAssetString("assets/aws/custom-ops.yml")

	// AWSReleaseVersions carries all versions of releases
	AWSReleaseVersions = file.MustAssetString("../../control-tower-ops/ops/versions-aws.json")

	// GCPReleaseVersions carries all versions of releases
	GCPReleaseVersions = file.MustAssetString("../../control-tower-ops/ops/versions-gcp.json")

	// AddNewCa carries the ops file that adds a new CA required for cert rotation
	AddNewCa = file.MustAssetString("assets/maintenance/add-new-ca.yml")

	// RemoveOldCa carries the ops file that removes the old CA required for cert rotation
	RemoveOldCa = file.MustAssetString("assets/maintenance/remove-old-ca.yml")

	// CleanupCerts moves renewed values of certs to old keys in director vars store
	CleanupCerts = file.MustAssetString("assets/maintenance/cleanup-certs.yml")

	AWSVersionFile = file.MustAsset("../../control-tower-ops/createenv-dependencies-and-cli-versions-aws.json")

	GCPVersionFile = file.MustAsset("../../control-tower-ops/createenv-dependencies-and-cli-versions-gcp.json")
)
