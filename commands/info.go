package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/EngineerBetter/control-tower/bosh"
	"github.com/EngineerBetter/control-tower/certs"
	"github.com/EngineerBetter/control-tower/commands/info"
	"github.com/EngineerBetter/control-tower/concourse"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/resource"
	"github.com/EngineerBetter/control-tower/terraform"
	"github.com/EngineerBetter/control-tower/util"
	"gopkg.in/urfave/cli.v1"
)

var initialInfoArgs info.Args

var infoFlags = []cli.Flag{
	cli.StringFlag{
		Name:        "region",
		Usage:       "(optional) AWS region",
		EnvVar:      "AWS_REGION",
		Destination: &initialInfoArgs.Region,
	},
	cli.BoolFlag{
		Name:        "json",
		Usage:       "(optional) Output as json",
		EnvVar:      "JSON",
		Destination: &initialInfoArgs.JSON,
	},
	cli.BoolFlag{
		Name:        "env",
		Usage:       "(optional) Output environment variables",
		Destination: &initialInfoArgs.Env,
	},
	cli.BoolFlag{
		Name:        "cert-expiry",
		Usage:       "(optional) Output only the expiration date of the director nats certificate",
		Destination: &initialInfoArgs.CertExpiry,
	},
	cli.StringFlag{
		Name:        "iaas",
		Usage:       "(required) IAAS, can be AWS or GCP",
		EnvVar:      "IAAS",
		Destination: &initialInfoArgs.IAAS,
	},
	cli.StringFlag{
		Name:        "namespace",
		Usage:       "(optional) Specify a namespace for deployments in order to group them in a meaningful way",
		EnvVar:      "NAMESPACE",
		Destination: &initialInfoArgs.Namespace,
	},
}

func infoAction(c *cli.Context, infoArgs info.Args, provider iaas.Provider) error {
	name := c.Args().Get(0)
	if name == "" {
		return errors.New("Usage is `control-tower info <name>`")
	}

	version := c.App.Version

	client, err := buildInfoClient(name, version, infoArgs, provider)
	if err != nil {
		return err
	}
	i, err := client.FetchInfo()
	if err != nil {
		return err
	}
	switch {
	case infoArgs.JSON:
		return json.NewEncoder(os.Stdout).Encode(i)
	case infoArgs.Env:
		env, err := i.Env()
		if err != nil {
			return err
		}
		_, err = os.Stdout.WriteString(env)
		return err
	case infoArgs.CertExpiry:
		os.Stdout.WriteString(i.CertExpiry)
		return nil
	default:
		_, err := fmt.Fprint(os.Stdout, i)
		return err
	}
}

func validateInfoArgs(c *cli.Context, infoArgs info.Args) (info.Args, error) {
	err := infoArgs.MarkSetFlags(c)
	if err != nil {
		return infoArgs, fmt.Errorf("failed to mark set Info flags: [%v]", err)
	}

	if err = infoArgs.Validate(); err != nil {
		return infoArgs, fmt.Errorf("failed to validate Info flags: [%v]", err)
	}

	return infoArgs, nil
}

func buildInfoClient(name, version string, infoArgs info.Args, provider iaas.Provider) (*concourse.Client, error) {
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
		config.New(provider, name, infoArgs.Namespace),
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

var infoCmd = cli.Command{
	Name:      "info",
	Aliases:   []string{"i"},
	Usage:     "Fetches information on a deployed environment",
	ArgsUsage: "<name>",
	Flags:     infoFlags,
	Action: func(c *cli.Context) error {
		infoArgs, err := validateInfoArgs(c, initialInfoArgs)
		if err != nil {
			return fmt.Errorf("Error validating args on info: [%v]", err)
		}
		iaasName, err := iaas.Validate(infoArgs.IAAS)
		if err != nil {
			return fmt.Errorf("Error mapping to supported IAASes on info: [%v]", err)
		}
		provider, err := iaas.New(iaasName, infoArgs.Region)
		if err != nil {
			return fmt.Errorf("Error creating IAAS provider on info: [%v]", err)
		}
		return infoAction(c, infoArgs, provider)
	},
}
