package boshcli

import (
	"fmt"
	"io/ioutil"

	"github.com/EngineerBetter/control-tower/resource"
	"github.com/EngineerBetter/control-tower/util"
	"github.com/EngineerBetter/control-tower/util/yaml"
)

// Environment holds all the parameters GCP IAAS needs
type GCPEnvironment struct {
	CustomOperations    string
	DirectorName        string
	ExternalIP          string
	GcpCredentialsJSON  string
	InternalCIDR        string
	InternalGW          string
	InternalIP          string
	Network             string
	PrivateCIDR         string
	PrivateCIDRGateway  string
	PrivateCIDRReserved string
	PrivateSubnetwork   string
	ProjectID           string
	PublicCIDR          string
	PublicCIDRGateway   string
	PublicCIDRReserved  string
	PublicCIDRStatic    string
	PublicKey           string
	PublicSubnetwork    string
	Spot                bool
	Tags                string
	Zone                string
}

// ConfigureDirectorManifestCPI interpolates all the Environment parameters and
// required release versions into ready to use Director manifest
func (e GCPEnvironment) ConfigureDirectorManifestCPI() (string, error) {
	gcpCreds, err := ioutil.ReadFile(e.GcpCredentialsJSON)
	if err != nil {
		return "", err
	}

	var allOperations = resource.GCPCPIOps + resource.GCPExternalIPOps + resource.GCPDirectorCustomOps + resource.GCPJumpboxUserOps

	return yaml.Interpolate(resource.DirectorManifest, allOperations+e.CustomOperations, map[string]interface{}{
		"internal_cidr":        e.InternalCIDR,
		"internal_gw":          e.InternalGW,
		"internal_ip":          e.InternalIP,
		"director_name":        e.DirectorName,
		"zone":                 e.Zone,
		"network":              e.Network,
		"subnetwork":           e.PublicSubnetwork,
		"private_subnetwork":   e.PrivateSubnetwork,
		"project_id":           e.ProjectID,
		"gcp_credentials_json": string(gcpCreds),
		"external_ip":          e.ExternalIP,
		"public_key":           e.PublicKey,
	})
}

type gcpCloudConfigParams struct {
	Zone                string
	Spot                bool
	PublicSubnetwork    string
	PrivateSubnetwork   string
	Network             string
	PublicCIDR          string
	PublicCIDRGateway   string
	PublicCIDRStatic    string
	PublicCIDRReserved  string
	PrivateCIDR         string
	PrivateCIDRGateway  string
	PrivateCIDRReserved string
}

// ConfigureDirectorCloudConfig inserts values from the environment into the config template passed as argument
func (e GCPEnvironment) ConfigureDirectorCloudConfig() (string, error) {
	templateParams := gcpCloudConfigParams{
		Zone:                e.Zone,
		PublicSubnetwork:    e.PublicSubnetwork,
		PrivateSubnetwork:   e.PrivateSubnetwork,
		Spot:                e.Spot,
		Network:             e.Network,
		PublicCIDR:          e.PublicCIDR,
		PublicCIDRGateway:   e.PublicCIDRGateway,
		PublicCIDRStatic:    e.PublicCIDRStatic,
		PublicCIDRReserved:  e.PublicCIDRReserved,
		PrivateCIDR:         e.PrivateCIDR,
		PrivateCIDRGateway:  e.PrivateCIDRGateway,
		PrivateCIDRReserved: e.PrivateCIDRReserved,
	}

	cc, err := util.RenderTemplate("cloud-config", resource.GCPDirectorCloudConfig, templateParams)
	if cc == nil {
		return "", err
	}
	return string(cc), err
}

// ConcourseStemcellURL returns the stemcell location string for an AWS specific stemcell for the required concourse version
func (e GCPEnvironment) ConcourseStemcellURL() (string, error) {
	version, err := getStemcellVersionFromOpsFile(resource.GCPReleaseVersions)
	if err != nil {
		return "", fmt.Errorf("Error getting GCP stemcell version for Concourse [%v]", err)
	}
	return fmt.Sprintf("https://s3.amazonaws.com/bosh-gce-light-stemcells/%s/light-bosh-stemcell-%s-google-kvm-ubuntu-xenial-go_agent.tgz", version, version), nil
}
