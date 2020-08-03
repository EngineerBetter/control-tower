package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/EngineerBetter/control-tower/bosh"
	"github.com/EngineerBetter/control-tower/certs"
	"github.com/EngineerBetter/control-tower/commands/destroy"
	"github.com/EngineerBetter/control-tower/concourse"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/resource"
	"github.com/EngineerBetter/control-tower/terraform"
	"github.com/EngineerBetter/control-tower/util"

	"gopkg.in/urfave/cli.v1"
)

//var destroyArgs config.DestroyArgs
var initialDestroyArgs destroy.Args

var destroyFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &initialDestroyArgs.Region,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(required) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Destination: &initialDestroyArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialDestroyArgs.Namespace,
	},
}

func destroyAction(c *cli.Context, destroyArgs destroy.Args, provider iaas.Provider) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `control-tower destroy <name>`")
	}

	if !NonInteractiveModeEnabled() {
		confirm, err := util.CheckConfirmation(os.Stdin, os.Stdout, name)
		if err != nil {
			return err
		}

		if !confirm {
			fmt.Println("Bailing out...")
			return nil
		}
	}

	version := c.App.Version

	client, err := buildDestroyClient(name, version, destroyArgs, provider)
	if err != nil {
		return err
	}
	return client.Destroy()
}

func validateDestroyArgs(c *cli.Context, destroyArgs destroy.Args) (destroy.Args, error) {
	err := destroyArgs.MarkSetFlags(c)
	if err != nil {
		return destroyArgs, fmt.Errorf("failed to mark set Destroy  flags: [%v]", err)
	}

	if err = destroyArgs.Validate(); err != nil {
		return destroyArgs, fmt.Errorf("failed to validate Destroy flags: [%v]", err)
	}

	return destroyArgs, nil
}

func buildDestroyClient(name, version string, destroyArgs destroy.Args, provider iaas.Provider) (*concourse.Client, error) {
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
		config.New(provider, name, destroyArgs.Namespace),
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

var destroyCmd = cli.Command{
	Name:      "destroy",
	Aliases:   []string{"x"},
	Usage:     "Destroys a Concourse",
	ArgsUsage: "<name>",
	Flags:     destroyFlags,
	Action: func(c *cli.Context) error {
		destroyArgs, err := validateDestroyArgs(c, initialDestroyArgs)
		if err != nil {
			return fmt.Errorf("Error validating args on destroy: [%v]", err)
		}
		iaasName, err := iaas.Validate(destroyArgs.IAAS)
		if err != nil {
			return fmt.Errorf("Error mapping to supported IAASes on destroy: [%v]", err)
		}
		provider, err := iaas.New(iaasName, destroyArgs.Region)
		if err != nil {
			return fmt.Errorf("Error creating IAAS provider on destroy: [%v]", err)
		}
		return destroyAction(c, destroyArgs, provider)
	},
}
