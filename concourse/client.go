package concourse

import (
	"io"

	"github.com/EngineerBetter/control-tower/commands/maintain"

	"github.com/EngineerBetter/control-tower/bosh"
	"github.com/EngineerBetter/control-tower/commands/deploy"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/terraform"
)

// client is a concrete implementation of IClient interface
type Client struct {
	boshClientFactory  bosh.ClientFactory
	configClient       config.IClient
	deployArgs         *deploy.Args
	eightRandomLetters func() string
	flyClientFactory   func(iaas.Provider, fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error)
	ipChecker          func() (string, error)
	passwordGenerator  func(int) string
	provider           iaas.Provider
	sshGenerator       func() ([]byte, []byte, string, error)
	stderr             io.Writer
	stdout             io.Writer
	tfCLI              terraform.CLIInterface
	tfInputVarsFactory TFInputVarsFactory
	version            string
	versionFile        []byte
}

// IClient represents a control-tower client
type IClient interface {
	Deploy() error
	Destroy() error
	FetchInfo() (*Info, error)
	Maintain(maintain.Args) error
}

// New returns a new client
func NewClient(
	provider iaas.Provider,
	tfCLI terraform.CLIInterface,
	tfInputVarsFactory TFInputVarsFactory,
	boshClientFactory bosh.ClientFactory,
	flyClientFactory func(iaas.Provider, fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error),
	configClient config.IClient,
	deployArgs *deploy.Args,
	stdout, stderr io.Writer,
	ipChecker func() (string, error),
	passwordGenerator func(int) string,
	eightRandomLetters func() string,
	sshGenerator func() ([]byte, []byte, string, error),
	version string,
	versionFile []byte) *Client {
	return &Client{
		boshClientFactory:  boshClientFactory,
		configClient:       configClient,
		deployArgs:         deployArgs,
		eightRandomLetters: eightRandomLetters,
		flyClientFactory:   flyClientFactory,
		ipChecker:          ipChecker,
		passwordGenerator:  passwordGenerator,
		provider:           provider,
		sshGenerator:       sshGenerator,
		stderr:             stderr,
		stdout:             stdout,
		tfCLI:              tfCLI,
		tfInputVarsFactory: tfInputVarsFactory,
		version:            version,
		versionFile:        versionFile,
	}
}

func (client *Client) buildBoshClient(config config.ConfigView, tfOutputs terraform.Outputs) (bosh.IClient, error) {

	return client.boshClientFactory(
		config,
		tfOutputs,
		client.stdout,
		client.stderr,
		client.provider,
		client.versionFile,
	)
}
