package terraform

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/resource"
	"github.com/EngineerBetter/control-tower/util"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// InputVars exposes ConfigureDirectorManifestCPI
type InputVars interface {
	ConfigureTerraform(string) (string, error)
}

//counterfeiter:generate . Outputs
// Outputs holds IAAS specific terraform outputs
type Outputs interface {
	AssertValid() error
	Init(*bytes.Buffer) error
	Get(string) (string, error)
}

//counterfeiter:generate . CLIInterface
//CLIInterface is the abstraction of execCmd
type CLIInterface interface {
	Apply(InputVars) error
	Destroy(InputVars) error
	BuildOutput(InputVars) (Outputs, error)
}

// CLI struct holds the abstraction of execCmd
type CLI struct {
	execCmd func(string, ...string) *exec.Cmd
	Path    string
	iaas    iaas.Name
}

//Factory function to return iaas-specific outputs
func outputsFor(name iaas.Name) (Outputs, error) {
	switch name {
	case iaas.AWS:
		return &AWSOutputs{}, nil
	case iaas.GCP:
		return &GCPOutputs{}, nil
	}
	return &NullOutputs{}, errors.New("terraform: " + name.String() + " not a valid iaas provider")
}

// Option defines the arbitrary element of Options for New
type Option func(*CLI) error

// Path returns the path of the terraform-cli as an Option
func Path(path string) Option {
	return func(c *CLI) error {
		c.Path = path
		return nil
	}
}

// DownloadTerraform returns the downloaded CLI path Option
func DownloadTerraform(versionFile []byte) Option {
	return func(c *CLI) error {
		var binaries map[string]util.BinaryPaths

		err := json.Unmarshal(versionFile, &binaries)
		if err != nil {
			return err
		}

		path, err := util.DownloadTerraformCLI(binaries)
		c.Path = path
		return err
	}
}

// New provides a new CLI
func New(iaas iaas.Name, ops ...Option) (*CLI, error) {
	cli := &CLI{
		execCmd: exec.Command,
		Path:    "terraform",
		iaas:    iaas,
	}
	for _, op := range ops {
		if err := op(cli); err != nil {
			return nil, err
		}
	}
	return cli, nil
}

type NullInputVars struct{}

func (n *NullInputVars) ConfigureTerraform(string) (string, error) { return "", nil }

func (n *NullInputVars) Build(map[string]interface{}) error { return nil }

type NullOutputs struct{}

func (n *NullOutputs) AssertValid() error { return nil }

func (n *NullOutputs) Init(*bytes.Buffer) error { return nil }

func (n *NullOutputs) Get(string) (string, error) { return "", nil }

func (c *CLI) init(config InputVars) (string, error) {
	var (
		tfConfig string
		err      error
	)
	switch c.iaas {
	case iaas.AWS:
		tfConfig, err = config.ConfigureTerraform(resource.AWSTerraformConfig)
		if err != nil {
			return "", err
		}
	case iaas.GCP:
		tfConfig, err = config.ConfigureTerraform(resource.GCPTerraformConfig)
		if err != nil {
			return "", err
		}
	}

	terraformConfigPath, err := writeTempFile([]byte(tfConfig))
	if err != nil {
		return "", err
	}
	cmd := c.execCmd(c.Path, "init")
	cmd.Dir = terraformConfigPath
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		os.RemoveAll(terraformConfigPath)
		return "", err
	}
	return terraformConfigPath, nil
}

// Apply runs terraform apply for a given config
func (c *CLI) Apply(config InputVars) error {
	terraformConfigPath, err := c.init(config)
	if err != nil {
		return err
	}

	defer os.RemoveAll(terraformConfigPath)

	cmd := c.execCmd(c.Path, "apply", "-input=false", "-auto-approve")
	cmd.Dir = terraformConfigPath

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

// Destroy destroys terraform resources specified in a config file
func (c *CLI) Destroy(config InputVars) error {
	terraformConfigPath, err := c.init(config)
	if err != nil {
		return err
	}

	defer os.RemoveAll(terraformConfigPath)

	cmd := c.execCmd(c.Path, "destroy", "-auto-approve")
	cmd.Dir = terraformConfigPath
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// BuildOutput builds the terraform output
func (c *CLI) BuildOutput(config InputVars) (Outputs, error) {
	terraformConfigPath, err := c.init(config)
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(terraformConfigPath)

	stdoutBuffer := bytes.NewBuffer(nil)
	cmd := c.execCmd(c.Path, "output", "-json")
	cmd.Dir = terraformConfigPath
	cmd.Stderr = os.Stderr
	cmd.Stdout = stdoutBuffer
	if err = cmd.Run(); err != nil {
		return nil, err
	}

	outputs, err := outputsFor(c.iaas)
	if err != nil {
		return nil, fmt.Errorf("Error creating blank TF Outputs before population: [%v]", err)
	}
	err = outputs.Init(stdoutBuffer)
	if err != nil {
		return nil, fmt.Errorf("Error populating blank TF Outputs: [%v]", err)
	}
	return outputs, nil
}

func writeTempFile(data []byte) (string, error) {
	mode := int(0740)
	perm := os.FileMode(mode)
	dirName := randomString()
	filePath := path.Join(os.TempDir(), dirName)
	err := os.MkdirAll(filePath, perm)
	if err != nil {
		return "", err
	}
	f, err := ioutil.TempFile(filePath, "*.tf")
	if err != nil {
		return "", err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.RemoveAll(filePath)
		return "", err
	}
	return filePath, err
}

func randomString() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x%x%x%x", b[0:2], b[2:4], b[4:6], b[6:8])
}
