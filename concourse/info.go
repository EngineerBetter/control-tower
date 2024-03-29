package concourse

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/EngineerBetter/control-tower/iaas"

	"github.com/EngineerBetter/control-tower/bosh"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/util/yaml"
	"github.com/fatih/color"
)

// Info represents the compound fields for info templates
type Info struct {
	Terraform   TerraformInfo   `json:"terraform"`
	Config      config.Config   `json:"config"`
	Instances   []bosh.Instance `json:"instances"`
	CertExpiry  string          `json:"cert_expiry"`
	GatewayUser string
}

// TerraformInfo represents the terraform output fields needed for the info templates
type TerraformInfo struct {
	DirectorPublicIP string
	NatGatewayIP     string
}

// FetchInfo fetches and builds the info
func (client *Client) FetchInfo() (*Info, error) {
	var gatewayUser string
	conf, err := client.configClient.Load()
	if err != nil {
		return nil, err
	}

	directorCredsBytes, err := loadDirectorCreds(client.configClient)
	if err != nil {
		return nil, err
	}

	var certExpiry string
	if len(directorCredsBytes) > 0 {
		natsCA, err1 := yaml.Path(directorCredsBytes, "nats_server_tls/ca")
		if err1 != nil {
			return nil, err1
		}

		var re = regexp.MustCompile(`\n\s*`)

		openSSL := exec.Command("openssl", "x509", "-noout", "-dates")
		openSSL.Stdin = strings.NewReader(re.ReplaceAllString(natsCA, "\n"))
		var out bytes.Buffer
		openSSL.Stdout = &out
		err1 = openSSL.Run()
		if err1 != nil {
			return nil, err1
		}
		if strings.Contains(out.String(), "notAfter=") {
			certExpiry = strings.Split(out.String(), "notAfter=")[1]
		} else {
			return nil, fmt.Errorf("openssl output is not as expected. got: %s", out.String())
		}
	}

	tfInputVars := client.tfInputVarsFactory.NewInputVars(conf)

	switch client.provider.IAAS() {
	case iaas.AWS:
		gatewayUser = "vcap"
	case iaas.GCP:
		gatewayUser = "jumpbox"
	}

	tfOutputs, err := client.tfCLI.BuildOutput(tfInputVars)
	if err != nil {
		return nil, err
	}

	directorPublicIP, err := tfOutputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, err
	}

	natGatewayIP, err := tfOutputs.Get("NatGatewayIP")
	if err != nil {
		return nil, err
	}

	terraformInfo := TerraformInfo{
		DirectorPublicIP: directorPublicIP,
		NatGatewayIP:     natGatewayIP,
	}

	userIP, err1 := client.ipChecker()
	if err1 != nil {
		return nil, err1
	}

	directorSecurityGroupID, err1 := tfOutputs.Get("DirectorSecurityGroupID")
	if err1 != nil {
		return nil, err1
	}
	whitelisted, err1 := client.provider.CheckForWhitelistedIP(userIP, directorSecurityGroupID)
	if err1 != nil {
		return nil, err1
	}

	if !whitelisted {
		err1 = fmt.Errorf("Do you need to add your IP %s to the %s-director security group/source range entry for director firewall (for ports 22, 6868, and 25555)?", userIP, conf.Deployment)
		return nil, err1
	}

	boshClient, err := client.buildBoshClient(conf, tfOutputs)
	if err != nil {
		return nil, err
	}
	defer boshClient.Cleanup()

	instances, err := boshClient.Instances()
	if err != nil {
		return nil, fmt.Errorf("Error getting BOSH instances: %s", err)
	}

	return &Info{
		Terraform:   terraformInfo,
		Config:      conf,
		Instances:   instances,
		GatewayUser: gatewayUser,
		CertExpiry:  certExpiry,
	}, nil
}

const infoTemplate = `Deployment:
	Namespace: {{.Config.Namespace}}
	IAAS:      {{.Config.IAAS}}
	Region:    {{.Config.Region}}

Workers:
	Count:              {{.Config.ConcourseWorkerCount}}
	Size:               {{.Config.ConcourseWorkerSize}}
	Outbound Public IP: {{.Terraform.NatGatewayIP}}

Instances:
{{range .Instances}}
	{{.Name}} {{.IP | replace "\n" ","}} {{.State}}
{{end}}

Concourse credentials:
	username: {{.Config.ConcourseUsername}}
	password: {{.Config.ConcoursePassword}}
	URL:      https://{{.Config.Domain}}

Credhub credentials:
	username: {{.Config.CredhubUsername}}
	password: {{.Config.CredhubPassword}}
	URL:      {{.Config.CredhubURL}}
	CA Cert:
		{{ .Config.CredhubCACert | replace "\n" "\n\t\t"}}

Grafana credentials (if metrics are enabled):
	username: {{.Config.ConcourseUsername}}
	password: {{.Config.ConcoursePassword}}
	URL:      https://{{.Config.Domain}}:3000

Bosh credentials:
	username: {{.Config.DirectorUsername}}
	password: {{.Config.DirectorPassword}}
	IP:       {{.Terraform.DirectorPublicIP}}
	CA Cert:
		{{ .Config.DirectorCACert | replace "\n" "\n\t\t"}}

BOSH-generated NAT certs will expire on: {{ .CertExpiry }}

Uses Control-Tower version {{.Config.Version}}

Built by {{"EngineerBetter http://engineerbetter.com" | blue}}
`

func (info *Info) String() string {
	t := template.Must(template.New("info").Funcs(template.FuncMap{
		"replace": func(old, new, s string) string {
			return strings.Replace(s, old, new, -1)
		},
		"blue": color.New(color.FgCyan, color.Bold).Sprint,
	}).Parse(infoTemplate))
	var buf bytes.Buffer
	err := t.Execute(&buf, info)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func writeTempFile(data string) (name string, err error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	name = f.Name()
	_, err = f.WriteString(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		os.Remove(name)
	}
	return name, err
}

var envTemplate = template.Must(template.New("env").Funcs(template.FuncMap{
	"to_file": writeTempFile,
}).Parse(`
export BOSH_ENVIRONMENT={{.Terraform.DirectorPublicIP}}
export BOSH_GW_HOST={{.Terraform.DirectorPublicIP}}
export BOSH_CA_CERT='{{.Config.DirectorCACert}}'
export BOSH_DEPLOYMENT=concourse
export BOSH_CLIENT={{.Config.DirectorUsername}}
export BOSH_CLIENT_SECRET={{.Config.DirectorPassword}}
export BOSH_GW_USER={{.GatewayUser}}
export BOSH_GW_PRIVATE_KEY={{.Config.PrivateKey | to_file}}
export CREDHUB_SERVER={{.Config.CredhubURL}}
export CREDHUB_CA_CERT='{{.Config.CredhubCACert}}'
export CREDHUB_CLIENT=credhub_admin
export CREDHUB_SECRET={{.Config.CredhubAdminClientSecret}}
export NAMESPACE={{.Config.Namespace}}
`))

// Env returns a string that is suitable for a shell to evaluate that sets environment
// varibles which are used to log into bosh and credhub
func (info *Info) Env() (string, error) {
	var buf bytes.Buffer
	var i Info
	i = *info
	err := envTemplate.Execute(&buf, i)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
