package bosh

import (
	"io"

	"github.com/EngineerBetter/control-tower/bosh/internal/boshcli"
	"github.com/EngineerBetter/control-tower/bosh/internal/workingdir"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/terraform"
)

//GCPClient is an GCP specific implementation of IClient
type GCPClient struct {
	config      config.ConfigView
	outputs     terraform.Outputs
	workingdir  workingdir.IClient
	stdout      io.Writer
	stderr      io.Writer
	provider    iaas.Provider
	boshCLI     boshcli.ICLI
	versionFile []byte
}

//NewGCPClient returns a GCP specific implementation of IClient
func NewGCPClient(config config.ConfigView, outputs terraform.Outputs, workingdir workingdir.IClient, stdout, stderr io.Writer, provider iaas.Provider, boshCLI boshcli.ICLI, versionFile []byte) (IClient, error) {
	return &GCPClient{
		config:     config,
		outputs:    outputs,
		workingdir: workingdir,
		stdout:     stdout,
		stderr:     stderr,
		provider:   provider,
		boshCLI:    boshCLI,
		versionFile: versionFile,
	}, nil
}

//Cleanup is GCP specific implementation of Cleanup
func (client *GCPClient) Cleanup() error {
	return client.workingdir.Cleanup()
}
