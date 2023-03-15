package config

const SPOT = "spot"
const ON_DEMAND = "on-demand"

func ConvertSpotBoolToVMProvisioningType(spot bool) string {
	if spot {
		return SPOT
	} else {
		return ON_DEMAND
	}
}

// Config represents a control-tower configuration file
type Config struct {
	AllowIPs                 string `json:"allow_ips"`
	AllowIPsUnformatted      string `json:"allow_ips_unformatted"`
	AvailabilityZone         string `json:"availability_zone"`
	BitbucketClientID        string `json:"bitbucket_client_id"`
	BitbucketClientSecret    string `json:"bitbucket_client_secret"`
	ConcourseCACert          string `json:"concourse_ca_cert"`
	ConcourseCert            string `json:"concourse_cert"`
	ConcourseKey             string `json:"concourse_key"`
	ConcoursePassword        string `json:"concourse_password"`
	ConcourseUsername        string `json:"concourse_username"`
	ConcourseWebSize         string `json:"concourse_web_size"`
	ConcourseWorkerCount     int    `json:"concourse_worker_count"`
	ConcourseWorkerSize      string `json:"concourse_worker_size"`
	ConfigBucket             string `json:"config_bucket"`
	CredhubAdminClientSecret string `json:"credhub_admin_client_secret"`
	CredhubCACert            string `json:"credhub_ca_cert"`
	CredhubPassword          string `json:"credhub_password"`
	CredhubURL               string `json:"credhub_url"`
	CredhubUsername          string `json:"credhub_username"`
	Deployment               string `json:"deployment"`
	DirectorCACert           string `json:"director_ca_cert"`
	DirectorCert             string `json:"director_cert"`
	DirectorHMUserPassword   string `json:"director_hm_user_password"`
	DirectorKey              string `json:"director_key"`
	DirectorMbusPassword     string `json:"director_mbus_password"`
	DirectorNATSPassword     string `json:"director_nats_password"`
	DirectorPassword         string `json:"director_password"`
	DirectorPublicIP         string `json:"director_public_ip"`
	DirectorRegistryPassword string `json:"director_registry_password"`
	DirectorUsername         string `json:"director_username"`
	Domain                   string `json:"domain"`
	EnableGlobalResources    bool   `json:"enable_global_resources"`
	EnablePipelineInstances  bool   `json:"enable_pipeline_instances"`
	InfluxDbRetention        string `json:"influx_db_retention_period"`
	EncryptionKey            string `json:"encryption_key"`
	GithubClientID           string `json:"github_client_id"`
	GithubClientSecret       string `json:"github_client_secret"`
	GithubHost               string `json:"github_host"`
	GithubCaCert             string `json:"github_ca_cert"`
	GrafanaPassword          string `json:"grafana_password"`
	HostedZoneID             string `json:"hosted_zone_id"`
	HostedZoneRecordPrefix   string `json:"hosted_zone_record_prefix"`
	IAAS                     string `json:"iaas"`
	MainGithubUsers          string `json:"main_github_users"`
	MainGithubTeams          string `json:"main_github_teams"`
	MainGithubOrgs           string `json:"main_github_orgs"`
	MicrosoftClientID        string `json:"microsoft_client_id"`
	MicrosoftClientSecret    string `json:"microsoft_client_secret"`
	MicrosoftTenant          string `json:"microsoft_tenant"`
	Namespace                string `json:"namespace"`
	NetworkCIDR              string `json:"network_cidr"`
	NoMetrics                bool   `json:"no_metrics"`
	PersistentDisk           string `json:"persistent_disk"`
	PrivateCIDR              string `json:"private_cidr"`
	PrivateKey               string `json:"private_key"`
	Project                  string `json:"project"`
	PublicCIDR               string `json:"public_cidr"`
	PublicKey                string `json:"public_key"`
	RDS1CIDR                 string `json:"rds1_cidr"`
	RDS2CIDR                 string `json:"rds2_cidr"`
	RDSDefaultDatabaseName   string `json:"rds_default_database_name"`
	RDSInstanceClass         string `json:"rds_instance_class"`
	RDSPassword              string `json:"rds_password"`
	RDSUsername              string `json:"rds_username"`
	RDSDiskEncryption        bool   `json:"rds_disk_encryption"`
	Region                   string `json:"region"`
	SourceAccessIP           string `json:"source_access_ip"`
	//Spot is deprecated, exists only as we need to migrate old configs to VMProvisioningType
	Spot               bool     `json:"spot"`
	Tags               []string `json:"tags"`
	TFStatePath        string   `json:"tf_state_path"`
	Version            string   `json:"version"`
	VMProvisioningType string   `json:"vm_provisioning_type"`
	WorkerType         string   `json:"worker_type"`
}

type ConfigView interface {
	GetAllowIPs() string
	GetAllowIPsUnformatted() string
	GetAvailabilityZone() string
	GetBitbucketClientID() string
	GetBitbucketClientSecret() string
	GetConcourseCACert() string
	GetConcourseCert() string
	GetConcourseKey() string
	GetConcoursePassword() string
	GetConcourseUsername() string
	GetConcourseWebSize() string
	GetConcourseWorkerCount() int
	GetConcourseWorkerSize() string
	GetConfigBucket() string
	GetCredhubAdminClientSecret() string
	GetCredhubCACert() string
	GetCredhubPassword() string
	GetCredhubURL() string
	GetCredhubUsername() string
	GetDeployment() string
	GetDirectorCACert() string
	GetDirectorCert() string
	GetDirectorHMUserPassword() string
	GetDirectorKey() string
	GetDirectorMbusPassword() string
	GetDirectorNATSPassword() string
	GetDirectorPassword() string
	GetDirectorPublicIP() string
	GetDirectorRegistryPassword() string
	GetDirectorUsername() string
	GetDomain() string
	GetEnableGlobalResources() bool
	GetEnablePipelineInstances() bool
	GetInfluxDbRetention() string
	GetEncryptionKey() string
	GetGithubClientID() string
	GetGithubClientSecret() string
	GetGithubHost() string
	GetGithubCaCert() string
	GetGrafanaPassword() string
	GetHostedZoneID() string
	GetHostedZoneRecordPrefix() string
	GetIAAS() string
	GetMainGithubUsers() string
	GetMainGithubTeams() string
	GetMainGithubOrgs() string
	GetMicrosoftClientID() string
	GetMicrosoftClientSecret() string
	GetMicrosoftTenant() string
	GetNamespace() string
	GetNetworkCIDR() string
	GetPersistentDiskSize() string
	GetPrivateCIDR() string
	GetPrivateKey() string
	GetProject() string
	GetPublicCIDR() string
	GetPublicKey() string
	GetRDS1CIDR() string
	GetRDS2CIDR() string
	GetRDSDefaultDatabaseName() string
	GetRDSInstanceClass() string
	GetRDSPassword() string
	GetRDSUsername() string
	GetRDSDiskEncryption() bool
	GetRegion() string
	GetSourceAccessIP() string
	GetTags() []string
	GetTFStatePath() string
	GetVersion() string
	GetWorkerType() string
	IsBitbucketAuthSet() bool
	IsGithubAuthSet() bool
	IsGithubEnterpriseAuthSet() bool
	IsMainGithubAuthSet() bool
	IsMicrosoftAuthSet() bool
	IsSpot() bool
	MetricsIsDisabled() bool
}

func (c Config) GetAllowIPs() string {
	return c.AllowIPs
}

func (c Config) GetAllowIPsUnformatted() string {
	return c.AllowIPsUnformatted
}

func (c Config) GetAvailabilityZone() string {
	return c.AvailabilityZone
}

func (c Config) GetBitbucketClientID() string {
	return c.BitbucketClientID
}

func (c Config) GetBitbucketClientSecret() string {
	return c.BitbucketClientSecret
}

func (c Config) GetConcourseCACert() string {
	return c.ConcourseCACert
}

func (c Config) GetConcourseCert() string {
	return c.ConcourseCert
}

func (c Config) GetConcourseKey() string {
	return c.ConcourseKey
}

func (c Config) GetConcoursePassword() string {
	return c.ConcoursePassword
}

func (c Config) GetConcourseUsername() string {
	return c.ConcourseUsername
}

func (c Config) GetConcourseWebSize() string {
	return c.ConcourseWebSize
}

func (c Config) GetPersistentDiskSize() string {
	return c.PersistentDisk
}

func (c Config) GetConcourseWorkerCount() int {
	return c.ConcourseWorkerCount
}

func (c Config) GetConcourseWorkerSize() string {
	return c.ConcourseWorkerSize
}

func (c Config) GetConfigBucket() string {
	return c.ConfigBucket
}

func (c Config) GetCredhubAdminClientSecret() string {
	return c.CredhubAdminClientSecret
}

func (c Config) GetCredhubCACert() string {
	return c.CredhubCACert
}

func (c Config) GetCredhubPassword() string {
	return c.CredhubPassword
}

func (c Config) GetCredhubURL() string {
	return c.CredhubURL
}

func (c Config) GetCredhubUsername() string {
	return c.CredhubUsername
}

func (c Config) GetDeployment() string {
	return c.Deployment
}

func (c Config) GetDirectorCACert() string {
	return c.DirectorCACert
}

func (c Config) GetDirectorCert() string {
	return c.DirectorCert
}

func (c Config) GetDirectorHMUserPassword() string {
	return c.DirectorHMUserPassword
}

func (c Config) GetDirectorKey() string {
	return c.DirectorKey
}

func (c Config) GetDirectorMbusPassword() string {
	return c.DirectorMbusPassword
}

func (c Config) GetDirectorNATSPassword() string {
	return c.DirectorNATSPassword
}

func (c Config) GetDirectorPassword() string {
	return c.DirectorPassword
}

func (c Config) GetDirectorPublicIP() string {
	return c.DirectorPublicIP
}

func (c Config) GetDirectorRegistryPassword() string {
	return c.DirectorRegistryPassword
}

func (c Config) GetDirectorUsername() string {
	return c.DirectorUsername
}

func (c Config) GetDomain() string {
	return c.Domain
}

func (c Config) GetEnableGlobalResources() bool {
	return c.EnableGlobalResources
}

func (c Config) GetInfluxDbRetention() string {
	return c.InfluxDbRetention
}

func (c Config) GetEnablePipelineInstances() bool {
	return c.EnablePipelineInstances
}

func (c Config) GetEncryptionKey() string {
	return c.EncryptionKey
}

func (c Config) GetGithubClientID() string {
	return c.GithubClientID
}

func (c Config) GetGithubClientSecret() string {
	return c.GithubClientSecret
}

func (c Config) GetGithubHost() string {
	return c.GithubHost
}

func (c Config) GetGithubCaCert() string {
	return c.GithubCaCert
}

func (c Config) GetGrafanaPassword() string {
	return c.GrafanaPassword
}

func (c Config) GetHostedZoneID() string {
	return c.HostedZoneID
}

func (c Config) GetHostedZoneRecordPrefix() string {
	return c.HostedZoneRecordPrefix
}

func (c Config) GetIAAS() string {
	return c.IAAS
}

func (c Config) GetMainGithubUsers() string {
	return c.MainGithubUsers
}

func (c Config) GetMainGithubTeams() string {
	return c.MainGithubTeams
}

func (c Config) GetMainGithubOrgs() string {
	return c.MainGithubOrgs
}

func (c Config) GetMicrosoftClientID() string {
	return c.MicrosoftClientID
}

func (c Config) GetMicrosoftClientSecret() string {
	return c.MicrosoftClientSecret
}

func (c Config) GetMicrosoftTenant() string {
	return c.MicrosoftTenant
}

func (c Config) GetNamespace() string {
	return c.Namespace
}

func (c Config) GetNetworkCIDR() string {
	return c.NetworkCIDR
}

func (c Config) GetPrivateCIDR() string {
	return c.PrivateCIDR
}

func (c Config) GetPrivateKey() string {
	return c.PrivateKey
}

func (c Config) GetProject() string {
	return c.Project
}

func (c Config) GetPublicCIDR() string {
	return c.PublicCIDR
}

func (c Config) GetPublicKey() string {
	return c.PublicKey
}

func (c Config) GetRDS1CIDR() string {
	return c.RDS1CIDR
}

func (c Config) GetRDS2CIDR() string {
	return c.RDS2CIDR
}

func (c Config) GetRDSDefaultDatabaseName() string {
	return c.RDSDefaultDatabaseName
}

func (c Config) GetRDSInstanceClass() string {
	return c.RDSInstanceClass
}

func (c Config) GetRDSPassword() string {
	return c.RDSPassword
}

func (c Config) GetRDSUsername() string {
	return c.RDSUsername
}

func (c Config) GetRDSDiskEncryption() bool {
	return c.RDSDiskEncryption
}

func (c Config) GetRegion() string {
	return c.Region
}

func (c Config) GetSourceAccessIP() string {
	return c.SourceAccessIP
}

func (c Config) GetTags() []string {
	return c.Tags
}

func (c Config) GetTFStatePath() string {
	return c.TFStatePath
}

func (c Config) GetVersion() string {
	return c.Version
}

func (c Config) GetWorkerType() string {
	return c.WorkerType
}

func (c Config) IsBitbucketAuthSet() bool {
	return c.BitbucketClientID != "" && c.BitbucketClientSecret != ""
}

func (c Config) IsGithubAuthSet() bool {
	return c.GithubClientID != "" && c.GithubClientSecret != ""
}

func (c Config) IsMainGithubAuthSet() bool {
	return c.MainGithubUsers != "" || c.MainGithubTeams != "" || c.MainGithubOrgs != ""
}

func (c Config) IsGithubEnterpriseAuthSet() bool {
	return c.GithubHost != "" && c.GithubCaCert != ""
}

func (c Config) IsMicrosoftAuthSet() bool {
	return c.MicrosoftClientID != "" && c.MicrosoftClientSecret != ""
}

func (c Config) IsSpot() bool {
	return c.VMProvisioningType == SPOT
}

func (c Config) MetricsIsDisabled() bool {
	return c.NoMetrics
}
