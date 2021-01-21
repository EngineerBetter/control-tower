package concourse

import (
	"fmt"

	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/terraform"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . TFInputVarsFactory
type TFInputVarsFactory interface {
	NewInputVars(conf config.ConfigView) terraform.InputVars
}

func NewTFInputVarsFactory(provider iaas.Provider) (TFInputVarsFactory, error) {
	if provider.IAAS() == iaas.AWS {
		return &AWSInputVarsFactory{}, nil
	} else if provider.IAAS() == iaas.GCP {
		credentialsPath, err := provider.Attr("credentials_path")
		if err != nil {
			return &GCPInputVarsFactory{}, fmt.Errorf("Error finding attribute [credentials_path]: [%v]", err)
		}

		project, err := provider.Attr("project")
		if err != nil {
			return &GCPInputVarsFactory{}, fmt.Errorf("Error finding attribute [project]: [%v]", err)
		}

		return &GCPInputVarsFactory{
			credentialsPath: credentialsPath,
			project:         project,
			region:          provider.Region(),
			zone:            provider.Zone("", ""),
		}, nil
	}

	return nil, fmt.Errorf("IAAS not supported [%s]", provider.IAAS())
}

type AWSInputVarsFactory struct{}

func (f *AWSInputVarsFactory) NewInputVars(c config.ConfigView) terraform.InputVars {
	return &terraform.AWSInputVars{
		NetworkCIDR:            c.GetNetworkCIDR(),
		PublicCIDR:             c.GetPublicCIDR(),
		PrivateCIDR:            c.GetPrivateCIDR(),
		AllowIPs:               c.GetAllowIPs(),
		AvailabilityZone:       c.GetAvailabilityZone(),
		ConfigBucket:           c.GetConfigBucket(),
		Deployment:             c.GetDeployment(),
		HostedZoneID:           c.GetHostedZoneID(),
		HostedZoneRecordPrefix: c.GetHostedZoneRecordPrefix(),
		Namespace:              c.GetNamespace(),
		Project:                c.GetProject(),
		PublicKey:              c.GetPublicKey(),
		RDSDefaultDatabaseName: c.GetRDSDefaultDatabaseName(),
		RDSInstanceClass:       c.GetRDSInstanceClass(),
		RDSPassword:            c.GetRDSPassword(),
		RDSUsername:            c.GetRDSUsername(),
		RDS1CIDR:               c.GetRDS1CIDR(),
		RDS2CIDR:               c.GetRDS2CIDR(),
		Region:                 c.GetRegion(),
		SourceAccessIP:         c.GetSourceAccessIP(),
		TFStatePath:            c.GetTFStatePath(),
	}
}

type GCPInputVarsFactory struct {
	credentialsPath string
	project         string
	region          string
	zone            string
}

func (f *GCPInputVarsFactory) NewInputVars(c config.ConfigView) terraform.InputVars {
	return &terraform.GCPInputVars{
		AllowIPs:           c.GetAllowIPs(),
		ConfigBucket:       c.GetConfigBucket(),
		DBName:             c.GetRDSDefaultDatabaseName(),
		DBPassword:         c.GetRDSPassword(),
		DBTier:             c.GetRDSInstanceClass(),
		DBUsername:         c.GetRDSUsername(),
		Deployment:         c.GetDeployment(),
		DNSManagedZoneName: c.GetHostedZoneID(),
		DNSRecordSetPrefix: c.GetHostedZoneRecordPrefix(),
		ExternalIP:         c.GetSourceAccessIP(),
		GCPCredentialsJSON: f.credentialsPath,
		Namespace:          c.GetNamespace(),
		Project:            f.project,
		Region:             f.region,
		Tags:               "",
		Zone:               f.zone,
		PublicCIDR:         c.GetPublicCIDR(),
		PrivateCIDR:        c.GetPrivateCIDR(),
	}
}
