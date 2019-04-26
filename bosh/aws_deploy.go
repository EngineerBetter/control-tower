package bosh

import (
	"net"

	"github.com/EngineerBetter/control-tower/bosh/internal/boshcli"
	"github.com/EngineerBetter/control-tower/db"
	"github.com/apparentlymart/go-cidr/cidr"
)

// Deploy implements deploy for AWS client
func (client *AWSClient) Deploy(state, creds []byte, detach bool) (newState, newCreds []byte, err error) {
	state, creds, err = client.CreateEnv(state, creds, "")
	if err != nil {
		return state, creds, err
	}

	if err = client.updateCloudConfig(client.boshCLI); err != nil {
		return state, creds, err
	}
	if err = client.uploadConcourseStemcell(client.boshCLI); err != nil {
		return state, creds, err
	}
	if err = client.createDefaultDatabases(); err != nil {
		return state, creds, err
	}

	creds, err = client.deployConcourse(creds, detach)
	if err != nil {
		return state, creds, err
	}

	return state, creds, err
}

// Locks implements locks for AWS client
func (client *AWSClient) Locks() ([]byte, error) {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}
	return client.boshCLI.Locks(boshcli.AWSEnvironment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.GetDirectorPassword(), client.config.GetDirectorCACert())

}

// CreateEnv exposes bosh create-env functionality
func (client *AWSClient) CreateEnv(state, creds []byte, customOps string) (newState, newCreds []byte, err error) {
	tags, err := splitTags(client.config.GetTags())
	if err != nil {
		return state, creds, err
	}
	tags["control-tower-project"] = client.config.GetProject()
	tags["control-tower-component"] = "concourse"
	//TODO(px): pull up this so that we use aws.Store
	store := temporaryStore{
		"vars.yaml":  creds,
		"state.json": state,
	}

	boshUserAccessKeyID, err1 := client.outputs.Get("BoshUserAccessKeyID")
	if err1 != nil {
		return state, creds, err1
	}
	boshSecretAccessKey, err1 := client.outputs.Get("BoshSecretAccessKey")
	if err1 != nil {
		return state, creds, err1
	}
	publicSubnetID, err1 := client.outputs.Get("PublicSubnetID")
	if err1 != nil {
		return state, creds, err1
	}
	privateSubnetID, err1 := client.outputs.Get("PrivateSubnetID")
	if err1 != nil {
		return state, creds, err1
	}
	directorPublicIP, err1 := client.outputs.Get("DirectorPublicIP")
	if err1 != nil {
		return state, creds, err1
	}
	atcSecurityGroupID, err1 := client.outputs.Get("ATCSecurityGroupID")
	if err1 != nil {
		return state, creds, err1
	}
	vmSecurityGroupID, err1 := client.outputs.Get("VMsSecurityGroupID")
	if err1 != nil {
		return state, creds, err1
	}
	blobstoreBucket, err1 := client.outputs.Get("BlobstoreBucket")
	if err1 != nil {
		return state, creds, err1
	}
	boshDBAddress, err1 := client.outputs.Get("BoshDBAddress")
	if err1 != nil {
		return state, creds, err1
	}
	boshDbPort, err1 := client.outputs.Get("BoshDBPort")
	if err1 != nil {
		return state, creds, err1
	}
	blobstoreUserAccessKeyID, err1 := client.outputs.Get("BlobstoreUserAccessKeyID")
	if err1 != nil {
		return state, creds, err1
	}
	blobstoreSecretAccessKey, err1 := client.outputs.Get("BlobstoreSecretAccessKey")
	if err1 != nil {
		return state, creds, err1
	}
	directorKeyPair, err1 := client.outputs.Get("DirectorKeyPair")
	if err1 != nil {
		return state, creds, err1
	}
	directorSecurityGroup, err1 := client.outputs.Get("DirectorSecurityGroupID")
	if err1 != nil {
		return state, creds, err1
	}

	publicCIDR := client.config.GetPublicCIDR()
	_, pubCIDR, err1 := net.ParseCIDR(publicCIDR)
	if err1 != nil {
		return state, creds, err1
	}
	internalGateway, err1 := cidr.Host(pubCIDR, 1)
	if err1 != nil {
		return state, creds, err1
	}
	directorInternalIP, err1 := cidr.Host(pubCIDR, 6)
	if err1 != nil {
		return state, creds, err1
	}

	err1 = client.boshCLI.CreateEnv(store, boshcli.AWSEnvironment{
		InternalCIDR:    client.config.GetPublicCIDR(),
		InternalGateway: internalGateway.String(),
		InternalIP:      directorInternalIP.String(),
		AccessKeyID:     boshUserAccessKeyID,
		SecretAccessKey: boshSecretAccessKey,
		Region:          client.config.GetRegion(),
		AZ:              client.config.GetAvailabilityZone(),
		DefaultKeyName:  directorKeyPair,
		DefaultSecurityGroups: []string{
			directorSecurityGroup,
			vmSecurityGroupID,
		},
		PrivateKey:           client.config.GetPrivateKey(),
		PublicSubnetID:       publicSubnetID,
		PrivateSubnetID:      privateSubnetID,
		ExternalIP:           directorPublicIP,
		ATCSecurityGroup:     atcSecurityGroupID,
		VMSecurityGroup:      vmSecurityGroupID,
		BlobstoreBucket:      blobstoreBucket,
		DBCACert:             db.RDSRootCert,
		DBHost:               boshDBAddress,
		DBName:               client.config.GetRDSDefaultDatabaseName(),
		DBPassword:           client.config.GetRDSPassword(),
		DBPort:               boshDbPort,
		DBUsername:           client.config.GetRDSUsername(),
		S3AWSAccessKeyID:     blobstoreUserAccessKeyID,
		S3AWSSecretAccessKey: blobstoreSecretAccessKey,
		Spot:                 client.config.IsSpot(),
		WorkerType:           client.config.GetWorkerType(),
		CustomOperations:     customOps,
	}, client.config.GetDirectorPassword(), client.config.GetDirectorCert(), client.config.GetDirectorKey(), client.config.GetDirectorCACert(), tags)
	if err1 != nil {
		return store["state.json"], store["vars.yaml"], err1
	}
	return store["state.json"], store["vars.yaml"], err
}

// Recreate exposes BOSH recreate
func (client *AWSClient) Recreate() error {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return client.boshCLI.Recreate(boshcli.AWSEnvironment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.GetDirectorPassword(), client.config.GetDirectorCACert())
}

func (client *AWSClient) updateCloudConfig(bosh boshcli.ICLI) error {
	publicSubnetID, err := client.outputs.Get("PublicSubnetID")
	if err != nil {
		return err
	}
	privateSubnetID, err := client.outputs.Get("PrivateSubnetID")
	if err != nil {
		return err
	}
	aTCSecurityGroupID, err := client.outputs.Get("ATCSecurityGroupID")
	if err != nil {
		return err
	}
	vMsSecurityGroupID, err := client.outputs.Get("VMsSecurityGroupID")
	if err != nil {
		return err
	}
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}

	publicCIDR := client.config.GetPublicCIDR()
	_, pubCIDR, err := net.ParseCIDR(publicCIDR)
	if err != nil {
		return err
	}
	pubGateway, err := cidr.Host(pubCIDR, 1)
	if err != nil {
		return err
	}
	publicCIDRGateway := pubGateway.String()
	publicCIDRStatic, err := formatIPRange(publicCIDR, ", ", []int{6, 7})
	if err != nil {
		return err
	}
	publicCIDRReserved, err := formatIPRange(publicCIDR, "-", []int{1, 5})
	if err != nil {
		return err
	}

	privateCIDR := client.config.GetPrivateCIDR()
	_, privCIDR, err := net.ParseCIDR(privateCIDR)
	if err != nil {
		return err
	}
	privGateway, err := cidr.Host(privCIDR, 1)
	if err != nil {
		return err
	}
	privateCIDRGateway := privGateway.String()
	privateCIDRReserved, err := formatIPRange(privateCIDR, "-", []int{1, 5})
	if err != nil {
		return err
	}

	return bosh.UpdateCloudConfig(boshcli.AWSEnvironment{
		AZ:                  client.config.GetAvailabilityZone(),
		PublicSubnetID:      publicSubnetID,
		PrivateSubnetID:     privateSubnetID,
		ATCSecurityGroup:    aTCSecurityGroupID,
		VMSecurityGroup:     vMsSecurityGroupID,
		Spot:                client.config.IsSpot(),
		ExternalIP:          directorPublicIP,
		WorkerType:          client.config.GetWorkerType(),
		PublicCIDR:          publicCIDR,
		PublicCIDRGateway:   publicCIDRGateway,
		PublicCIDRStatic:    publicCIDRStatic,
		PublicCIDRReserved:  publicCIDRReserved,
		PrivateCIDR:         privateCIDR,
		PrivateCIDRGateway:  privateCIDRGateway,
		PrivateCIDRReserved: privateCIDRReserved,
	}, directorPublicIP, client.config.GetDirectorPassword(), client.config.GetDirectorCACert())
}
func (client *AWSClient) uploadConcourseStemcell(bosh boshcli.ICLI) error {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return bosh.UploadConcourseStemcell(boshcli.AWSEnvironment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.GetDirectorPassword(), client.config.GetDirectorCACert())
}
