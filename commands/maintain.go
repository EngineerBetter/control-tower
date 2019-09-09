package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/EngineerBetter/control-tower/commands/maintain"
	"github.com/EngineerBetter/control-tower/resource"

	"github.com/EngineerBetter/control-tower/bosh"
	"github.com/EngineerBetter/control-tower/certs"
	"github.com/EngineerBetter/control-tower/concourse"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/terraform"
	"github.com/EngineerBetter/control-tower/util"
	"gopkg.in/urfave/cli.v1"
)

var initialMaintainArgs maintain.Args
var provider iaas.Provider

var maintainFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &initialMaintainArgs.Region,
	},
	cli.BoolFlag{
		Name:        "renew-nats-cert",
		Usage:       "(optional) Rotate nats certificate",
		Destination: &initialMaintainArgs.RenewNatsCert,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(required) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Destination: &initialMaintainArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialMaintainArgs.Namespace,
	},
	cli.IntFlag{
		Name:        "stage",
		Usage:       "(optional) Set the desired stage for nats rotation tasks",
		EnvVar:      "STAGE",
		Destination: &initialMaintainArgs.Stage,
	},
}

func maintainAction(c *cli.Context, maintainArgs maintain.Args, provider iaas.Provider) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `control-tower maintain <name>`")
	}

	version := c.App.Version

	client, err := buildMaintainClient(name, version, maintainArgs, provider)
	if err != nil {
		return err
	}
	err = client.Maintain(maintainArgs)
	if err != nil {
		return err
	}
	//this will never run
	return nil
}

func validateMaintainArgs(c *cli.Context, maintainArgs maintain.Args) (maintain.Args, error) {
	err := maintainArgs.MarkSetFlags(c)
	if err != nil {
		return maintainArgs, fmt.Errorf("failed to mark set Maintain flags: [%v]", err)
	}

	if err = maintainArgs.Validate(); err != nil {
		return maintainArgs, fmt.Errorf("failed to validate Maintain flags: [%v]", err)
	}

	return maintainArgs, nil
}

func buildMaintainClient(name, version string, maintainArgs maintain.Args, provider iaas.Provider) (*concourse.Client, error) {
	versionFile, _ := provider.Choose(iaas.Choice{
		AWS: resource.AWSVersionFile,
		GCP: resource.GCPVersionFile,
	}).([]byte)

	terraformClient, err := terraform.New(provider.IAAS(), terraform.DownloadTerraform(versionFile))
	if err != nil {
		return nil, err
	}

	tfInputVarsFactory, err := concourse.NewTFInputVarsFactory(provider)
	if err != nil {
		return nil, fmt.Errorf("Error creating TFInputVarsFactory [%v]", err)
	}

	client := concourse.NewClient(
		provider,
		terraformClient,
		tfInputVarsFactory,
		bosh.New,
		fly.New,
		certs.Generate,
		config.New(provider, name, maintainArgs.Namespace),
		nil,
		os.Stdout,
		os.Stderr,
		util.FindUserIP,
		certs.NewAcmeClient,
		util.GeneratePasswordWithLength,
		util.EightRandomLetters,
		util.GenerateSSHKeyPair,
		version,
		versionFile,
	)

	return client, nil
}

var maintainCmd = cli.Command{
	Name:      "maintain",
	Aliases:   []string{"m"},
	Usage:     "Handles maintenance operations in control-tower",
	ArgsUsage: "<name>",
	Flags:     maintainFlags,
	Action: func(c *cli.Context) error {
		maintainArgs, err := validateMaintainArgs(c, initialMaintainArgs)
		if err != nil {
			return fmt.Errorf("Error validating args on maintain: [%v]", err)
		}
		iaasName, err := iaas.Validate(maintainArgs.IAAS)
		if err != nil {
			return fmt.Errorf("Error mapping to supported IAASes on maintain: [%v]", err)
		}
		provider, err := iaas.New(iaasName, maintainArgs.Region)
		if err != nil {
			return fmt.Errorf("Error creating IAAS provider on maintain: [%v]", err)
		}
		return maintainAction(c, maintainArgs, provider)
	},
}
