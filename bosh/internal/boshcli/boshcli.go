package boshcli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/EngineerBetter/control-tower/util"
	"github.com/EngineerBetter/control-tower/util/yaml"
)

//counterfeiter:generate . ICLI
type ICLI interface {
	CreateEnv(createEnvFiles *CreateEnvFiles, config IAASEnvironment, password, cert, key, ca string, tags map[string]string) (*CreateEnvFiles, error)
	RunAuthenticatedCommand(action, ip, password, ca string, detach bool, stdout io.Writer, flags ...string) error
	Locks(config IAASEnvironment, ip, password, ca string) ([]byte, error)
	Recreate(config IAASEnvironment, ip, password, ca string) error
	UpdateCloudConfig(config IAASEnvironment, ip, password, ca string) error
	UploadConcourseStemcell(config IAASEnvironment, ip, password, ca string) error
}

type CreateEnvFiles struct {
	StateFileContents []byte
	VarsFileContents  []byte
}

// CLI struct holds the abstraction of execCmd
type CLI struct {
	execCmd  func(string, ...string) *exec.Cmd
	boshPath string
}

// New provides a new CLI
func New(boshPath string, execCmdFunc func(string, ...string) *exec.Cmd) ICLI {
	return &CLI{
		execCmd:  execCmdFunc,
		boshPath: boshPath,
	}
}

type IAASEnvironment interface {
	ConfigureDirectorManifestCPI() (string, error)
	ConfigureDirectorCloudConfig() (string, error)
	ConcourseStemcellURL() (string, error)
	ExtractBOSHandBPM() (util.Resource, util.Resource, error)
}

func concourseStemcellURL(releaseVersionsFile, urlFormat string) (string, error) {
	var ops []struct {
		Path  string
		Value json.RawMessage
	}
	err := json.Unmarshal([]byte(releaseVersionsFile), &ops)
	if err != nil {
		return "", err
	}
	var version string
	for _, op := range ops {
		if op.Path != "/stemcells/alias=xenial/version" {
			continue
		}
		err := json.Unmarshal(op.Value, &version)
		if err != nil {
			return "", err
		}
	}
	if version == "" {
		return "", errors.New("did not find stemcell version in versions.json")
	}

	return fmt.Sprintf(urlFormat, version, version), nil
}

// UpdateCloudConfig generates cloud config from template and use it to update bosh cloud config
func (c *CLI) UpdateCloudConfig(config IAASEnvironment, ip, password, ca string) error {
	var cloudConfig string
	var err error

	cloudConfig, err = config.ConfigureDirectorCloudConfig()
	if err != nil {
		return err
	}
	cloudConfigPath, err := writeTempFile([]byte(cloudConfig))
	if err != nil {
		return err
	}
	defer os.Remove(cloudConfigPath)
	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return err
	}
	defer os.Remove(caPath)
	ip = fmt.Sprintf("https://%s", ip)
	cmd := c.execCmd(c.boshPath, "--non-interactive", "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "update-cloud-config", cloudConfigPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// Locks runs bosh locks
func (c *CLI) Locks(config IAASEnvironment, ip, password, ca string) ([]byte, error) {
	var out bytes.Buffer
	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return nil, err
	}
	defer os.Remove(caPath)
	cmd := c.execCmd(c.boshPath, "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "locks", "--json")
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// UploadConcourseStemcell uploads a stemcell for the chosen IAAS
func (c *CLI) UploadConcourseStemcell(config IAASEnvironment, ip, password, ca string) error {
	var (
		stemcell string
		err      error
	)

	stemcell, err = config.ConcourseStemcellURL()
	if err != nil {
		return err
	}

	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return err
	}
	defer os.Remove(caPath)
	ip = fmt.Sprintf("https://%s", ip)
	cmd := c.execCmd(c.boshPath, "--non-interactive", "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "upload-stemcell", stemcell)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// Recreate runs BOSH recreate
func (c *CLI) Recreate(config IAASEnvironment, ip, password, ca string) error {
	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return err
	}
	defer os.Remove(caPath)
	ip = fmt.Sprintf("https://%s", ip)
	cmd := c.execCmd(c.boshPath, "--non-interactive", "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "--deployment", "concourse", "recreate")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func (c *CLI) CreateEnv(createEnvFiles *CreateEnvFiles, config IAASEnvironment, password, cert, key, ca string, tags map[string]string) (*CreateEnvFiles, error) {
	manifest, err := config.ConfigureDirectorManifestCPI()
	if err != nil {
		return &CreateEnvFiles{}, err
	}

	boshResource, bpmResource, err := config.ExtractBOSHandBPM()
	if err != nil {
		return &CreateEnvFiles{}, err
	}

	vars := map[string]interface{}{
		"director_name":            "bosh",
		"admin_password":           password,
		"director_ssl.certificate": cert,
		"director_ssl.private_key": key,
		"director_ssl.ca":          ca,
		"bosh_url":                 boshResource.URL,
		"bosh_version":             boshResource.Version,
		"bosh_sha1":                boshResource.SHA1,
		"bpm_url":                  bpmResource.URL,
		"bpm_version":              bpmResource.Version,
		"bpm_sha1":                 bpmResource.SHA1,
		"tags":                     tags,
	}
	manifest, err = yaml.Interpolate(manifest, "", vars)
	if err != nil {
		return &CreateEnvFiles{}, err
	}
	statePath, err := writeNonEmptyTempFile(createEnvFiles.StateFileContents, "state.json")
	if err != nil {
		return &CreateEnvFiles{}, err
	}
	varsPath, err := writeNonEmptyTempFile(createEnvFiles.VarsFileContents, "vars.yml")
	if err != nil {
		return &CreateEnvFiles{}, err
	}
	manifestPath, err := writeTempFile([]byte(manifest))
	if err != nil {
		return &CreateEnvFiles{}, err
	}
	defer os.Remove(manifestPath)

	cmd := c.execCmd(c.boshPath, "create-env", "--state="+statePath, "--vars-store="+varsPath, manifestPath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()

	stateFileContents, err1 := ioutil.ReadFile(statePath)
	if err1 != nil {
		return &CreateEnvFiles{}, fmt.Errorf("Error loading state file after create-env: [%v]", err1)
	}
	varsFileContents, err1 := ioutil.ReadFile(varsPath)
	if err1 != nil {
		return &CreateEnvFiles{}, fmt.Errorf("Error loading vars file after create-env: [%v]", err1)
	}

	createEnvFiles = &CreateEnvFiles{
		StateFileContents: stateFileContents,
		VarsFileContents:  varsFileContents,
	}

	return createEnvFiles, err
}

// RunAuthenticatedCommand runs the bosh command `action` with flags `flags`
// specifying `detach` will cause the task to detach once a deployment starts
// `detach` is currently only implemented with the action `deploy`
func (c *CLI) RunAuthenticatedCommand(action, ip, password, ca string, detach bool, stdout io.Writer, flags ...string) error {
	caPath, err := writeTempFile([]byte(ca))
	if err != nil {
		return err
	}
	defer os.Remove(caPath)
	ip = fmt.Sprintf("https://%s", ip)

	authFlags := []string{"--non-interactive", "--environment", ip, "--ca-cert", caPath, "--client", "admin", "--client-secret", password, "--deployment", "concourse", action}
	flags = append(authFlags, flags...)
	if detach && action == "deploy" {
		return c.detachedBoshCommand(stdout, flags...)
	}
	return c.boshCommand(stdout, flags...)
}

func (c *CLI) boshCommand(stdout io.Writer, flags ...string) error {
	cmd := c.execCmd(c.boshPath, flags...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = stdout
	return cmd.Run()
}

func (c *CLI) detachedBoshCommand(stdout io.Writer, flags ...string) error {
	cmd := c.execCmd(c.boshPath, flags...)
	cmd.Stderr = os.Stderr

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)

	if err := cmd.Start(); err != nil {
		return err
	}

	for scanner.Scan() {
		text := scanner.Text()
		if _, err := stdout.Write([]byte(fmt.Sprintf("%s\n", text))); err != nil {
			return err
		}
		if strings.Contains(text, "Preparing deployment") {
			stdout.Write([]byte("Task started, detaching output\n"))
			return nil
		}
	}

	return fmt.Errorf("Didn't detect successful task start in BOSH comand: bosh-cli %s", strings.Join(flags, " "))
}

// If data is empty, return a path to where one could put a temp file
func writeNonEmptyTempFile(data []byte, filename string) (string, error) {
	if len(data) == 0 && filename != "" {
		dir, err := ioutil.TempDir("", "control-tower")
		if err != nil {
			return "", fmt.Errorf("Error generating temp directory: %v", err)
		}
		return filepath.Join(dir, filename), nil
	}

	return writeTempFile(data)
}

func writeTempFile(data []byte) (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	name := f.Name()
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.Remove(name)
	}
	return name, err
}
