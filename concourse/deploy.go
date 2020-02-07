package concourse

import (
	"fmt"
	"io"
	"text/template"

	"strings"

	"github.com/EngineerBetter/control-tower/bosh"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/terraform"
	"github.com/EngineerBetter/control-tower/util"
	"gopkg.in/yaml.v2"
)

// BoshParams represents the params used and produced by a BOSH deploy
type BoshParams struct {
	AutoCert                 bool
	CredhubPassword          string
	CredhubAdminClientSecret string
	CredhubCACert            string
	ConcourseCACert          string
	ConcourseCert            string
	ConcourseKey             string
	CredhubURL               string
	CredhubUsername          string
	ConcourseUsername        string
	ConcoursePassword        string
	GrafanaPassword          string
	DirectorUsername         string
	DirectorPassword         string
	DirectorCACert           string
}

func stripVersion(tags []string) []string {
	output := []string{}
	for _, tag := range tags {
		if !strings.HasPrefix(tag, "control-tower-version") {
			output = append(output, tag)
		}
	}
	return output
}

// Deploy deploys a concourse instance
func (client *Client) Deploy() error {
	err := client.configClient.EnsureBucketExists()
	if err != nil {
		return fmt.Errorf("error ensuring config bucket exists before deploy: [%v]", err)
	}

	conf, err := client.getInitialConfig()
	if err != nil {
		return fmt.Errorf("error getting initial config before deploy: [%v]", err)
	}

	r, err := client.checkPreTerraformConfigRequirements(conf, client.deployArgs.SelfUpdate)
	if err != nil {
		return err
	}
	conf.Region = r.Region
	conf.SourceAccessIP = r.SourceAccessIP
	conf.HostedZoneID = r.HostedZoneID
	conf.HostedZoneRecordPrefix = r.HostedZoneRecordPrefix
	conf.Domain = r.Domain

	tfInputVars := client.tfInputVarsFactory.NewInputVars(conf)

	err = client.tfCLI.Apply(tfInputVars)
	if err != nil {
		return err
	}

	tfOutputs, err := client.tfCLI.BuildOutput(tfInputVars)
	if err != nil {
		return err
	}

	err = client.configClient.Update(conf)
	if err != nil {
		return err
	}
	conf.Tags = stripVersion(conf.Tags)
	conf.Tags = append([]string{fmt.Sprintf("control-tower-version=%s", client.version)}, conf.Tags...)

	conf.Version = client.version

	cr, err := client.checkPreDeployConfigRequirements(conf, tfOutputs)
	if err != nil {
		return err
	}

	conf.AutoCert = cr.Certs.Autocert
	if client.deployArgs.TLSCert != "" {
		conf.ConcourseCert = cr.Certs.ConcourseCert
		conf.ConcourseKey = cr.Certs.ConcourseKey
	}
	conf.Domain = cr.Domain
	conf.DirectorPublicIP = cr.DirectorPublicIP

	var bp BoshParams
	if client.deployArgs.SelfUpdate {
		bp, err = client.updateBoshAndPipeline(conf, tfOutputs)
	} else {
		bp, err = client.deployBoshAndPipeline(conf, tfOutputs)
	}

	conf.CredhubPassword = bp.CredhubPassword
	conf.CredhubAdminClientSecret = bp.CredhubAdminClientSecret
	conf.CredhubCACert = bp.CredhubCACert
	conf.CredhubURL = bp.CredhubURL
	conf.CredhubUsername = bp.CredhubUsername
	conf.ConcourseUsername = bp.ConcourseUsername
	conf.ConcoursePassword = bp.ConcoursePassword
	conf.GrafanaPassword = bp.GrafanaPassword
	conf.DirectorUsername = bp.DirectorUsername
	conf.DirectorPassword = bp.DirectorPassword
	conf.DirectorCACert = bp.DirectorCACert

	if !conf.AutoCert && client.deployArgs.TLSCert == "" {
		conf.ConcourseCert = bp.ConcourseCert
		conf.ConcourseKey = bp.ConcourseKey
		conf.ConcourseCACert = bp.ConcourseCACert
	}

	err1 := client.configClient.Update(conf)
	if err == nil {
		err = err1
	}
	return err
}

func (client *Client) deployBoshAndPipeline(c config.ConfigView, tfOutputs terraform.Outputs) (BoshParams, error) {
	// When we are deploying for the first time rather than updating
	// ensure that the pipeline is set _after_ the concourse is deployed

	bp, err := client.deployBosh(c, tfOutputs, false)
	if err != nil {
		return bp, err
	}

	flyClient, err := client.flyClientFactory(client.provider, fly.Credentials{
		Target:   c.GetDeployment(),
		API:      fmt.Sprintf("https://%s", c.GetDomain()),
		Username: bp.ConcourseUsername,
		Password: bp.ConcoursePassword,
	},
		client.stdout,
		client.stderr,
		client.versionFile,
	)
	if err != nil {
		return bp, err
	}
	defer flyClient.Cleanup()

	if err := flyClient.SetDefaultPipeline(c, false); err != nil {
		return bp, err
	}

	params := deployMessageParams{
		ConcoursePassword:         bp.ConcoursePassword,
		ConcourseUsername:         bp.ConcourseUsername,
		ConcourseUserProvidedCert: client.deployArgs.TLSCertIsSet && client.deployArgs.TLSKeyIsSet,
		Domain:                    c.GetDomain(),
		IAAS:                      c.GetIAAS(),
		Namespace:                 c.GetNamespace(),
		Project:                   c.GetProject(),
		Region:                    c.GetRegion(),
	}

	return bp, writeDeploySuccessMessage(params, client.stdout)
}

func (client *Client) updateBoshAndPipeline(c config.ConfigView, tfOutputs terraform.Outputs) (BoshParams, error) {
	// If concourse is already running this is an update rather than a fresh deploy
	// When updating we need to deploy the BOSH as the final step in order to
	// Detach from the update, so the update job can exit

	bp := BoshParams{
		CredhubPassword:          c.GetCredhubPassword(),
		CredhubAdminClientSecret: c.GetCredhubAdminClientSecret(),
		CredhubCACert:            c.GetCredhubCACert(),
		CredhubURL:               c.GetCredhubURL(),
		CredhubUsername:          c.GetCredhubUsername(),
		ConcourseUsername:        c.GetConcourseUsername(),
		ConcoursePassword:        c.GetConcoursePassword(),
		GrafanaPassword:          c.GetGrafanaPassword(),
		DirectorUsername:         c.GetDirectorUsername(),
		DirectorPassword:         c.GetDirectorPassword(),
		DirectorCACert:           c.GetDirectorCACert(),
	}

	flyClient, err := client.flyClientFactory(client.provider, fly.Credentials{
		Target:   c.GetDeployment(),
		API:      fmt.Sprintf("https://%s", c.GetDomain()),
		Username: c.GetConcourseUsername(),
		Password: c.GetConcoursePassword(),
	},
		client.stdout,
		client.stderr,
		client.versionFile,
	)
	if err != nil {
		return bp, err
	}
	defer flyClient.Cleanup()

	concourseAlreadyRunning, err := flyClient.CanConnect()
	if err != nil {
		return bp, err
	}

	if !concourseAlreadyRunning {
		return bp, fmt.Errorf("In detach mode but it seems that concourse is not currently running")
	}

	// Allow a fly version discrepancy since we might be targetting an older Concourse
	if err = flyClient.SetDefaultPipeline(c, true); err != nil {
		return bp, err
	}

	bp, err = client.deployBosh(c, tfOutputs, true)
	if err != nil {
		return bp, err
	}

	_, err = client.stdout.Write([]byte("\nUPGRADE RUNNING IN BACKGROUND\n\n"))

	return bp, err
}

// TerraformRequirements represents the required values for running terraform
type TerraformRequirements struct {
	Region                 string
	SourceAccessIP         string
	HostedZoneID           string
	HostedZoneRecordPrefix string
	Domain                 string
}

func (client *Client) checkPreTerraformConfigRequirements(conf config.ConfigView, selfUpdate bool) (TerraformRequirements, error) {
	r := TerraformRequirements{
		Region:                 conf.GetRegion(),
		SourceAccessIP:         conf.GetSourceAccessIP(),
		HostedZoneID:           conf.GetHostedZoneID(),
		HostedZoneRecordPrefix: conf.GetHostedZoneRecordPrefix(),
		Domain:                 conf.GetDomain(),
	}

	region := client.provider.Region()
	if conf.GetRegion() != "" {
		if conf.GetRegion() != region {
			return r, fmt.Errorf("found previous deployment in %s. Refusing to deploy to %s as changing regions for existing deployments is not supported", conf.GetRegion(), region)
		}
	}

	r.Region = region

	// When in self-update mode do not override the user IP, since we already have access to the worker
	if !selfUpdate {
		var err error
		r.SourceAccessIP, err = client.setUserIP(conf)
		if err != nil {
			return r, err
		}
	}

	zone, err := client.setHostedZone(conf, conf.GetDomain())
	if err != nil {
		return r, err
	}
	r.HostedZoneID = zone.HostedZoneID
	r.HostedZoneRecordPrefix = zone.HostedZoneRecordPrefix
	r.Domain = zone.Domain

	return r, nil
}

// Certs represents the certificate of a Concourse
type Certs struct {
	Autocert        bool
	ConcourseCert   string
	ConcourseKey    string
	ConcourseCACert string
}

// Requirements represents the pre deployment requirements of a Concourse
type Requirements struct {
	Domain           string
	DirectorPublicIP string
	Certs            Certs
}

func (client *Client) checkPreDeployConfigRequirements(cfg config.ConfigView, tfOutputs terraform.Outputs) (Requirements, error) {
	cr := Requirements{
		Domain:           cfg.GetDomain(),
		DirectorPublicIP: cfg.GetDirectorPublicIP(),
	}

	if cfg.GetDomain() == "" {
		domain, err := tfOutputs.Get("ATCPublicIP")
		if err != nil {
			return cr, err
		}
		cr.Domain = domain
	}

	cc := Certs{
		Autocert:        cfg.GetAutoCert(),
		ConcourseCert:   cfg.GetConcourseCert(),
		ConcourseKey:    cfg.GetConcourseKey(),
		ConcourseCACert: cfg.GetConcourseCACert(),
	}

	cc, err := client.ensureConcourseCerts(cc, cr.Domain)
	if err != nil {
		return cr, err
	}

	cr.Certs = cc

	cr.DirectorPublicIP, err = tfOutputs.Get("DirectorPublicIP")
	if err != nil {
		return cr, err
	}

	return cr, nil
}

func (client *Client) ensureConcourseCerts(cc Certs, domain string) (Certs, error) {
	certs := cc

	if client.deployArgs.TLSCert != "" {
		certs.ConcourseCert = client.deployArgs.TLSCert
		certs.ConcourseKey = client.deployArgs.TLSKey
		return certs, nil
	}

	// If a TLS cert has not been provided by the user
	// and a domain (non-IP) has been provided
	// use lets encrypt to generate cert
	if !util.IsIP(domain) {
		certs.Autocert = true
	}

	return certs, nil
}

func (client *Client) deployBosh(config config.ConfigView, tfOutputs terraform.Outputs, detach bool) (BoshParams, error) {
	bp := BoshParams{
		AutoCert:                 config.GetAutoCert(),
		CredhubPassword:          config.GetCredhubPassword(),
		CredhubAdminClientSecret: config.GetCredhubAdminClientSecret(),
		CredhubCACert:            config.GetCredhubCACert(),
		CredhubURL:               config.GetCredhubURL(),
		CredhubUsername:          config.GetCredhubUsername(),
		ConcourseUsername:        config.GetConcourseUsername(),
		ConcoursePassword:        config.GetConcoursePassword(),
		ConcourseCACert:          config.GetConcourseCACert(),
		ConcourseCert:            config.GetConcourseCert(),
		ConcourseKey:             config.GetConcourseKey(),
		GrafanaPassword:          config.GetGrafanaPassword(),
		DirectorUsername:         config.GetDirectorUsername(),
		DirectorPassword:         config.GetDirectorPassword(),
		DirectorCACert:           config.GetDirectorCACert(),
	}

	boshClient, err := client.buildBoshClient(config, tfOutputs)
	if err != nil {
		return bp, err
	}
	defer boshClient.Cleanup()

	boshStateBytes, err := loadDirectorState(client.configClient)
	if err != nil {
		return bp, err
	}
	boshCredsBytes, err := loadDirectorCreds(client.configClient)
	if err != nil {
		return bp, err
	}

	boshStateBytes, boshAndConcourseCredsBytes, err := boshClient.Deploy(boshStateBytes, boshCredsBytes, detach)

	err1 := client.configClient.StoreAsset(bosh.StateFilename, boshStateBytes)
	if err == nil {
		err = err1
	}
	err1 = client.configClient.StoreAsset(bosh.CredsFilename, boshAndConcourseCredsBytes)
	if err == nil {
		err = err1
	}
	if err != nil {
		return bp, err
	}

	var cc struct {
		CredhubPassword          string `yaml:"credhub_cli_password"`
		CredhubAdminClientSecret string `yaml:"credhub_admin_client_secret"`
		InternalTLS              struct {
			CA string `yaml:"ca"`
		} `yaml:"internal_tls"`
		AtcPassword string `yaml:"atc_password"`
		DirectorCA  struct {
			Cert string `yaml:"certificate"`
		} `yaml:"default_ca"`
		ConcourseCA struct {
			Cert string `yaml:"certificate"`
		} `yaml:"ca"`
		ConcourseTLS struct {
			Cert string `yaml:"certificate"`
			Key  string `yaml:"private_key"`
		} `yaml:"external_tls"`
	}

	err = yaml.Unmarshal(boshAndConcourseCredsBytes, &cc)
	if err != nil {
		return bp, err
	}

	bp.CredhubPassword = cc.CredhubPassword
	bp.CredhubAdminClientSecret = cc.CredhubAdminClientSecret
	bp.CredhubCACert = cc.InternalTLS.CA
	bp.CredhubURL = fmt.Sprintf("https://%s:8844/", config.GetDomain())
	bp.CredhubUsername = "credhub-cli"
	bp.ConcourseUsername = "admin"
	bp.DirectorCACert = cc.DirectorCA.Cert
	if len(cc.AtcPassword) > 0 {
		bp.ConcoursePassword = cc.AtcPassword
		bp.GrafanaPassword = cc.AtcPassword
	}
	if !bp.AutoCert {
		bp.ConcourseCACert = cc.ConcourseCA.Cert
		bp.ConcourseCert = cc.ConcourseTLS.Cert
		bp.ConcourseKey = cc.ConcourseTLS.Key
	}

	return bp, nil
}

func (client *Client) setUserIP(c config.ConfigView) (string, error) {
	sourceAccessIP := c.GetSourceAccessIP()
	userIP, err := client.ipChecker()
	if err != nil {
		return sourceAccessIP, err
	}

	if sourceAccessIP != userIP {
		sourceAccessIP = userIP
		_, err = client.stderr.Write([]byte(fmt.Sprintf(
			"\nWARNING: allowing access from local machine (address: %s)\n\n", userIP)))
		if err != nil {
			return sourceAccessIP, err
		}
	}

	return sourceAccessIP, nil
}

// HostedZone represents a DNS hosted zone
type HostedZone struct {
	HostedZoneID           string
	HostedZoneRecordPrefix string
	Domain                 string
}

func (client *Client) setHostedZone(c config.ConfigView, domain string) (HostedZone, error) {
	zone := HostedZone{
		HostedZoneID:           c.GetHostedZoneID(),
		HostedZoneRecordPrefix: c.GetHostedZoneRecordPrefix(),
		Domain:                 c.GetDomain(),
	}
	if domain == "" {
		return zone, nil
	}

	hostedZoneName, hostedZoneID, err := client.provider.FindLongestMatchingHostedZone(domain)
	if err != nil {
		return zone, err
	}
	zone.HostedZoneID = hostedZoneID

	if domain == hostedZoneName {
		zone.HostedZoneRecordPrefix = ""
	} else {
		zone.HostedZoneRecordPrefix = strings.TrimSuffix(domain, fmt.Sprintf(".%s", hostedZoneName))
		if c.GetIAAS() == "GCP" {
			zone.HostedZoneRecordPrefix = fmt.Sprintf("%s.", zone.HostedZoneRecordPrefix)
		}
	}
	zone.Domain = domain

	_, err = client.stderr.Write([]byte(fmt.Sprintf(
		"\nWARNING: adding record %s to DNS zone %s with name %s\n\n", domain, hostedZoneName, hostedZoneID)))
	if err != nil {
		return zone, err
	}
	return zone, err
}

const deployMsg = `DEPLOY SUCCESSFUL. Log in with:
fly --target {{.Project}} login{{if not .ConcourseUserProvidedCert}} --insecure{{end}} --concourse-url https://{{.Domain}} --username {{.ConcourseUsername}} --password {{.ConcoursePassword}}

Metrics available at https://{{.Domain}}:3000 using the same username and password

Log into credhub with:
eval "$(control-tower info --region {{.Region}} {{ if ne .Namespace .Region }} --namespace {{ .Namespace }} {{ end }} --iaas {{ .IAAS }} --env {{.Project}})"

Please complete our quick 7-question survey so that we can learn how & why you use Control Tower! http://bit.ly/eb-ctower
`

type deployMessageParams struct {
	ConcoursePassword         string
	ConcourseUsername         string
	ConcourseUserProvidedCert bool
	Domain                    string
	IAAS                      string
	Namespace                 string
	Project                   string
	Region                    string
}

func writeDeploySuccessMessage(params deployMessageParams, stdout io.Writer) error {
	t := template.Must(template.New("deploy").Parse(deployMsg))
	return t.Execute(stdout, params)
}

func writeConfigLoadedSuccessMessage(stdout io.Writer) {
	stdout.Write([]byte("\nUSING PREVIOUS DEPLOYMENT CONFIG\n"))
}

func loadDirectorState(configClient config.IClient) ([]byte, error) {
	hasState, err := configClient.HasAsset(bosh.StateFilename)
	if err != nil {
		return nil, err
	}

	if !hasState {
		return nil, nil
	}

	return configClient.LoadAsset(bosh.StateFilename)
}
func loadDirectorCreds(configClient config.IClient) ([]byte, error) {
	hasCreds, err := configClient.HasAsset(bosh.CredsFilename)
	if err != nil {
		return nil, err
	}

	if !hasCreds {
		return nil, nil
	}

	return configClient.LoadAsset(bosh.CredsFilename)
}
