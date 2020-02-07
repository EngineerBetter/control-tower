package bosh

import (
	"net"

	"github.com/EngineerBetter/control-tower/bosh/internal/boshcli"
	"github.com/apparentlymart/go-cidr/cidr"
)

// Deploy deploys a new Bosh director or converges an existing deployment
// Returns new contents of bosh state file
func (client *GCPClient) Deploy(state, creds []byte, detach bool) (newState, newBoshAndConcourseCreds []byte, err error) {
	if err != nil {
		return state, creds, err
	}

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

	newBoshAndConcourseCreds, err = client.deployConcourse(creds, detach)
	if err != nil {
		return state, creds, err
	}

	return state, newBoshAndConcourseCreds, err
}

// CreateEnv exposes bosh create-env functionality
func (client *GCPClient) CreateEnv(state, creds []byte, customOps string) (newState, newCreds []byte, err error) {
	tags, err := splitTags(client.config.GetTags())
	if err != nil {
		return state, creds, err
	}
	tags["control-tower-project"] = client.config.GetProject()
	tags["control-tower-component"] = "concourse"

	network, err1 := client.outputs.Get("Network")
	if err1 != nil {
		return state, creds, err1
	}
	publicSubnetwork, err1 := client.outputs.Get("PublicSubnetworkName")
	if err1 != nil {
		return state, creds, err1
	}
	privateSubnetwork, err1 := client.outputs.Get("PrivateSubnetworkName")
	if err1 != nil {
		return state, creds, err1
	}
	directorPublicIP, err1 := client.outputs.Get("DirectorPublicIP")
	if err1 != nil {
		return state, creds, err1
	}
	project, err1 := client.provider.Attr("project")
	if err1 != nil {
		return state, creds, err1
	}
	credentialsPath, err1 := client.provider.Attr("credentials_path")
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

	createEnvFiles, err1 := client.boshCLI.CreateEnv(&boshcli.CreateEnvFiles{StateFileContents: state, VarsFileContents: creds}, boshcli.GCPEnvironment{
		InternalCIDR:       client.config.GetPublicCIDR(),
		InternalGW:         internalGateway.String(),
		InternalIP:         directorInternalIP.String(),
		DirectorName:       "bosh",
		Zone:               client.provider.Zone("", ""),
		Network:            network,
		PublicSubnetwork:   publicSubnetwork,
		PrivateSubnetwork:  privateSubnetwork,
		Tags:               "[internal]",
		ProjectID:          project,
		GcpCredentialsJSON: credentialsPath,
		ExternalIP:         directorPublicIP,
		Spot:               client.config.IsSpot(),
		PublicKey:          client.config.GetPublicKey(),
		CustomOperations:   customOps,
		VersionFile:        client.versionFile,
	}, client.config.GetDirectorPassword(), tags)
	if err1 != nil {
		return createEnvFiles.StateFileContents, createEnvFiles.VarsFileContents, err1
	}
	return createEnvFiles.StateFileContents, createEnvFiles.VarsFileContents, err
}

// Recreate exposes BOSH recreate
func (client *GCPClient) Recreate() error {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return client.boshCLI.Recreate(boshcli.GCPEnvironment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.GetDirectorPassword(), client.config.GetDirectorCACert())
}

// Locks implements locks for GCP client
func (client *GCPClient) Locks() ([]byte, error) {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}
	return client.boshCLI.Locks(boshcli.GCPEnvironment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.GetDirectorPassword(), client.config.GetDirectorCACert())

}

func (client *GCPClient) updateCloudConfig(bosh boshcli.ICLI) error {

	privateSubnetwork, err := client.outputs.Get("PrivateSubnetworkName")
	if err != nil {
		return err
	}
	publicSubnetwork, err := client.outputs.Get("PublicSubnetworkName")
	if err != nil {
		return err
	}
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	network, err := client.outputs.Get("Network")
	if err != nil {
		return err
	}
	zone := client.provider.Zone("", "")

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

	publicCIDRStatic, err := formatIPRange(publicCIDR, ", ", []int{7})
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

	return bosh.UpdateCloudConfig(boshcli.GCPEnvironment{
		PublicCIDR:          client.config.GetPublicCIDR(),
		PublicCIDRGateway:   publicCIDRGateway,
		PublicCIDRStatic:    publicCIDRStatic,
		PublicCIDRReserved:  publicCIDRReserved,
		PrivateCIDRGateway:  privateCIDRGateway,
		PrivateCIDRReserved: privateCIDRReserved,
		PrivateCIDR:         client.config.GetPrivateCIDR(),
		Spot:                client.config.IsSpot(),
		PublicSubnetwork:    publicSubnetwork,
		PrivateSubnetwork:   privateSubnetwork,
		Zone:                zone,
		Network:             network,
	}, directorPublicIP, client.config.GetDirectorPassword(), client.config.GetDirectorCACert())
}
func (client *GCPClient) uploadConcourseStemcell(bosh boshcli.ICLI) error {
	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return err
	}
	return bosh.UploadConcourseStemcell(boshcli.GCPEnvironment{
		ExternalIP: directorPublicIP,
	}, directorPublicIP, client.config.GetDirectorPassword(), client.config.GetDirectorCACert())
}
