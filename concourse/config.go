package concourse

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/EngineerBetter/control-tower/commands/deploy"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/asaskevich/govalidator"
	"github.com/imdario/mergo"
)

func (client *Client) getInitialConfig() (config.Config, bool, error) {
	priorConfigExists, err := client.configClient.ConfigExists()
	if err != nil {
		return config.Config{}, false, fmt.Errorf("error determining if config already exists [%v]", err)
	}

	var isDomainUpdated bool
	var conf config.Config

	defaultConf := client.configClient.NewConfig()
	defaultConf, err = populateConfigWithDefaults(defaultConf, client.provider, client.passwordGenerator, client.sshGenerator, client.eightRandomLetters)
	if err != nil {
		return config.Config{}, false, fmt.Errorf("error generating default config: [%v]", err)
	}

	if priorConfigExists {
		conf, err = client.configClient.Load()
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error loading existing config [%v]", err)
		}
		writeConfigLoadedSuccessMessage(client.stdout)

		err = mergo.Merge(&conf, defaultConf)
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error layering stored config on top default config [%v]", err)
		}

		err = assertImmutableFieldsNotChanging(client.deployArgs, conf)
		if err != nil {
			return config.Config{}, false, err
		}

		conf, isDomainUpdated, err = applyArgumentsToConfig(conf, client.deployArgs, client.provider)
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error merging new options with existing config: [%v]", err)
		}
	} else {
		conf, _, err = applyArgumentsToConfig(defaultConf, client.deployArgs, client.provider)
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error applying arguments to default config: [%v]", err)
		}

		conf = applyImmutableArgumentsToConfig(conf, client.deployArgs, client.provider)

		err = client.configClient.Update(conf)
		if err != nil {
			return config.Config{}, false, fmt.Errorf("error persisting new config after setting values [%v]", err)
		}

		isDomainUpdated = true
	}

	return conf, isDomainUpdated, nil
}

func assertImmutableFieldsNotChanging(deployArgs *deploy.Args, conf config.ConfigView) error {
	if deployArgs.NetworkCIDRIsSet || deployArgs.PrivateCIDRIsSet || deployArgs.PublicCIDRIsSet {
		return fmt.Errorf("custom CIDRs cannot be applied after intial deploy")
	}

	// This is a safeguard for a redeployment where zone does not belong to the region where the original deployment has happened
	if deployArgs.ZoneIsSet && deployArgs.Zone != conf.GetAvailabilityZone() {
		return fmt.Errorf("Existing deployment uses zone %s and cannot change to zone %s", conf.GetAvailabilityZone(), deployArgs.Zone)
	}

	return nil
}

func populateConfigWithDefaults(conf config.Config, provider iaas.Provider, passwordGenerator func(int) string, sshGenerator func() ([]byte, []byte, string, error), eightRandomLetters func() string) (config.Config, error) {
	const defaultPasswordLength = 20

	privateKey, publicKey, _, err := sshGenerator()
	if err != nil {
		return config.Config{}, fmt.Errorf("error generating SSH keypair for new config: [%v]", err)
	}

	conf.AvailabilityZone = ""
	conf.ConcourseWebSize = "small"
	conf.ConcourseWorkerCount = 1
	conf.ConcourseWorkerSize = "xlarge"
	conf.DirectorHMUserPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorMbusPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorNATSPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorRegistryPassword = passwordGenerator(defaultPasswordLength)
	conf.DirectorUsername = "admin"
	conf.EncryptionKey = passwordGenerator(32)
	conf.IAAS = provider.IAAS().String()
	conf.PrivateKey = strings.TrimSpace(string(privateKey))
	conf.PublicKey = strings.TrimSpace(string(publicKey))
	conf.NoMetrics = false
	conf.RDSInstanceClass = provider.DBType("small")
	conf.RDSPassword = passwordGenerator(defaultPasswordLength)
	conf.RDSUsername = "admin" + passwordGenerator(7)
	conf.VMProvisioningType = config.SPOT
	conf.WorkerType = "m4"
	conf = populateConfigWithDefaultCIDRs(conf, provider)

	switch provider.IAAS() {
	case iaas.AWS:
		conf.RDSDefaultDatabaseName = fmt.Sprintf("bosh_%s", eightRandomLetters())
	case iaas.GCP:
		conf.RDSDefaultDatabaseName = fmt.Sprintf("bosh-%s", eightRandomLetters())
	}

	return conf, nil
}

func applyArgumentsToConfig(conf config.Config, deployArgs *deploy.Args, provider iaas.Provider) (config.Config, bool, error) {
	allow, err := parseAllowedIPsCIDRs(deployArgs.AllowIPs)
	if err != nil {
		return config.Config{}, false, fmt.Errorf("error determining IP addresses to allow access from: [%v]", err)
	}

	allowedIPs, err := getUpdatedAllowedIPs(allow)
	if err != nil {
		return config.Config{}, false, fmt.Errorf("error updating IP addresses to allow access from: [%v]", err)
	}

	// Moved validation here from deploy_args to support checking for github auth in config as well as in deployargs
	if deployArgs.MainGithubAuthIsSet {
		if !deployArgs.GithubAuthIsSet && (conf.GithubClientID == "" || conf.GithubClientSecret == "") {
			return config.Config{}, false, errors.New("Main team github auth flags can only be used when github auth is also configured")
		}
	}

	conf.AllowIPs = allowedIPs
	conf.AllowIPsUnformatted = deployArgs.AllowIPs

	if deployArgs.ZoneIsSet {
		conf.AvailabilityZone = deployArgs.Zone
	}
	if deployArgs.WorkerCountIsSet {
		conf.ConcourseWorkerCount = deployArgs.WorkerCount
	}
	if deployArgs.WorkerSizeIsSet {
		conf.ConcourseWorkerSize = deployArgs.WorkerSize
	}
	if deployArgs.WebSizeIsSet {
		conf.ConcourseWebSize = deployArgs.WebSize
	}
	if deployArgs.DBSizeIsSet {
		conf.RDSInstanceClass = provider.DBType(deployArgs.DBSize)
	}
	if deployArgs.BitbucketAuthIsSet {
		conf.BitbucketClientID = deployArgs.BitbucketAuthClientID
		conf.BitbucketClientSecret = deployArgs.BitbucketAuthClientSecret
	}
	if deployArgs.GithubAuthIsSet {
		conf.GithubClientID = deployArgs.GithubAuthClientID
		conf.GithubClientSecret = deployArgs.GithubAuthClientSecret
	}
	if deployArgs.MainGithubAuthIsSet {
		conf.MainGithubUsers = deployArgs.MainGithubUsers
		conf.MainGithubTeams = deployArgs.MainGithubTeams
		conf.MainGithubOrgs = deployArgs.MainGithubOrgs
	}
	if deployArgs.MicrosoftAuthIsSet {
		conf.MicrosoftClientID = deployArgs.MicrosoftAuthClientID
		conf.MicrosoftClientSecret = deployArgs.MicrosoftAuthClientSecret
		conf.MicrosoftTenant = deployArgs.MicrosoftAuthTenant
	}
	if deployArgs.NoMetricsIsSet {
		conf.NoMetrics = deployArgs.NoMetrics
	}
	if deployArgs.TagsIsSet {
		conf.Tags = deployArgs.Tags
	}
	if deployArgs.SpotIsSet {
		conf.VMProvisioningType = config.ConvertSpotBoolToVMProvisioningType(deployArgs.Spot)
	}
	if deployArgs.WorkerTypeIsSet {
		conf.WorkerType = deployArgs.WorkerType
	}

	if deployArgs.EnableGlobalResourcesIsSet {
		conf.EnableGlobalResources = deployArgs.EnableGlobalResources
	}
	if deployArgs.EnablePipelineInstancesIsSet {
		conf.EnablePipelineInstances = deployArgs.EnablePipelineInstances
	}

	// Flag has default value, hence it's always set.
	conf.InfluxDbRetention = deployArgs.InfluxDbRetention

	var isDomainUpdated bool
	if deployArgs.DomainIsSet {
		if conf.Domain != deployArgs.Domain {
			isDomainUpdated = true
		}
		conf.Domain = deployArgs.Domain
	} else {
		if govalidator.IsIPv4(conf.Domain) {
			conf.Domain = ""
		}
	}

	return conf, isDomainUpdated, nil
}

// Set config fields that are only valid on first deployment
func applyImmutableArgumentsToConfig(conf config.Config, deployArgs *deploy.Args, provider iaas.Provider) config.Config {
	if hasCIDRFlagsSet(deployArgs, provider) {
		conf = populateConfigWithDeployArgsCIDRs(conf, deployArgs, provider)
	}

	conf.AvailabilityZone = provider.Zone(deployArgs.Zone, conf.ConcourseWorkerSize)
	return conf
}

func hasCIDRFlagsSet(deployArgs *deploy.Args, provider iaas.Provider) bool {
	switch provider.IAAS() {
	case iaas.AWS:
		return deployArgs.NetworkCIDRIsSet && deployArgs.PublicCIDRIsSet && deployArgs.PrivateCIDRIsSet
	case iaas.GCP:
		return deployArgs.PublicCIDRIsSet && deployArgs.PrivateCIDRIsSet
	default:
		return false
	}
}

func populateConfigWithDeployArgsCIDRs(conf config.Config, deployArgs *deploy.Args, provider iaas.Provider) config.Config {
	switch provider.IAAS() {
	case iaas.AWS:
		conf.NetworkCIDR = deployArgs.NetworkCIDR
		conf.PublicCIDR = deployArgs.PublicCIDR
		conf.PrivateCIDR = deployArgs.PrivateCIDR
		conf.RDS1CIDR = deployArgs.RDS1CIDR
		conf.RDS2CIDR = deployArgs.RDS2CIDR
	case iaas.GCP:
		conf.PublicCIDR = deployArgs.PublicCIDR
		conf.PrivateCIDR = deployArgs.PrivateCIDR
	}
	return conf
}

func populateConfigWithDefaultCIDRs(conf config.Config, provider iaas.Provider) config.Config {
	switch provider.IAAS() {
	case iaas.AWS:
		conf.NetworkCIDR = "10.0.0.0/16"
		conf.PrivateCIDR = "10.0.1.0/24"
		conf.PublicCIDR = "10.0.0.0/24"
		conf.RDS1CIDR = "10.0.4.0/24"
		conf.RDS2CIDR = "10.0.5.0/24"
	case iaas.GCP:
		conf.PrivateCIDR = "10.0.1.0/24"
		conf.PublicCIDR = "10.0.0.0/24"
	}
	return conf
}

func getUpdatedAllowedIPs(ingressAddresses cidrBlocks) (string, error) {
	addr, err := ingressAddresses.String()
	if err != nil {
		return "", err
	}
	return addr, nil
}

type cidrBlocks []*net.IPNet

func parseAllowedIPsCIDRs(s string) (cidrBlocks, error) {
	var x cidrBlocks
	for _, ip := range strings.Split(s, ",") {
		ip = strings.TrimSpace(ip)
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			ipNet = &net.IPNet{
				IP:   net.ParseIP(ip),
				Mask: net.CIDRMask(32, 32),
			}
		}
		if ipNet.IP == nil {
			return nil, fmt.Errorf("could not parse %q as an IP address or CIDR range", ip)
		}
		x = append(x, ipNet)
	}
	return x, nil
}

func (b cidrBlocks) String() (string, error) {
	var buf bytes.Buffer
	for i, ipNet := range b {
		if i > 0 {
			_, err := fmt.Fprintf(&buf, ", %q", ipNet)
			if err != nil {
				return "", err
			}
		} else {
			_, err := fmt.Fprintf(&buf, "%q", ipNet)
			if err != nil {
				return "", err
			}
		}
	}
	return buf.String(), nil
}
