package bosh

import (
	"fmt"
	"io"
	"net"

	"github.com/EngineerBetter/control-tower/bosh/internal/boshcli"
	"github.com/EngineerBetter/control-tower/bosh/internal/workingdir"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/terraform"
	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
)

//AWSClient is an AWS specific implementation of IClient
type AWSClient struct {
	config      config.ConfigView
	outputs     terraform.Outputs
	workingdir  workingdir.IClient
	db          Opener
	stdout      io.Writer
	stderr      io.Writer
	provider    iaas.Provider
	boshCLI     boshcli.ICLI
	versionFile []byte
}

//NewAWSClient returns a AWS specific implementation of IClient
func NewAWSClient(config config.ConfigView, outputs terraform.Outputs, workingdir workingdir.IClient, stdout, stderr io.Writer, provider iaas.Provider, boshCLI boshcli.ICLI, versionFile []byte) (IClient, error) {
	directorPublicIP, err := outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, fmt.Errorf("failed to get DirectorPublicIP from terraform outputs: [%v]", err)
	}
	addr := net.JoinHostPort(directorPublicIP, "22")
	key, err := ssh.ParsePrivateKey([]byte(config.GetPrivateKey()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key for bosh: [%v]", err)
	}
	conf := &ssh.ClientConfig{
		User:            "vcap",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
	}
	var boshDBAddress, boshDBPort string

	boshDBAddress, err = outputs.Get("BoshDBAddress")
	if err != nil {
		return nil, fmt.Errorf("failed to get BoshDBAddress from terraform outputs: [%v]", err)
	}
	boshDBPort, err = outputs.Get("BoshDBPort")
	if err != nil {
		return nil, fmt.Errorf("failed to get BoshDBPort from terraform outputs: [%v]", err)
	}

	db, err := newProxyOpener(addr, conf, &pq.Driver{},
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
			config.GetRDSUsername(),
			config.GetRDSPassword(),
			boshDBAddress,
			boshDBPort,
			config.GetRDSDefaultDatabaseName(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create db proxyOpener: [%v]", err)
	}

	return &AWSClient{
		config:      config,
		outputs:     outputs,
		workingdir:  workingdir,
		db:          db,
		stdout:      stdout,
		stderr:      stderr,
		provider:    provider,
		boshCLI:     boshCLI,
		versionFile: versionFile,
	}, nil
}

//Cleanup is AWS specific implementation of Cleanup
func (client *AWSClient) Cleanup() error {
	return client.workingdir.Cleanup()
}
