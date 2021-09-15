package bosh

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/EngineerBetter/control-tower/db"
	"github.com/apparentlymart/go-cidr/cidr"
)

func (client *AWSClient) deployConcourse(creds []byte, detach bool) ([]byte, error) {

	err := saveFilesToWorkingDir(client.workingdir, client.provider, creds)
	if err != nil {
		return creds, fmt.Errorf("failed saving files to working directory in deployConcourse: [%v]", err)
	}

	boshDBAddress, err := client.outputs.Get("BoshDBAddress")
	if err != nil {
		return creds, err
	}
	boshDBPort, err := client.outputs.Get("BoshDBPort")
	if err != nil {
		return creds, err
	}
	atcPublicIP, err := client.outputs.Get("ATCPublicIP")
	if err != nil {
		return creds, err
	}

	publicCIDR := client.config.GetPublicCIDR()
	_, pubCIDR, err1 := net.ParseCIDR(publicCIDR)
	if err1 != nil {
		return creds, err
	}
	atcPrivateIP, err := cidr.Host(pubCIDR, 8)
	if err != nil {
		return creds, err
	}

	vmap := map[string]interface{}{
		"deployment_name":           concourseDeploymentName,
		"domain":                    client.config.GetDomain(),
		"project":                   client.config.GetProject(),
		"web_network_name":          "public",
		"worker_network_name":       "private",
		"postgres_host":             boshDBAddress,
		"postgres_port":             boshDBPort,
		"postgres_role":             client.config.GetRDSUsername(),
		"postgres_password":         client.config.GetRDSPassword(),
		"postgres_ca_cert":          db.RDSRootCert,
		"web_vm_type":               "concourse-web-" + client.config.GetConcourseWebSize(),
		"worker_vm_type":            "concourse-" + client.config.GetConcourseWorkerSize(),
		"worker_count":              client.config.GetConcourseWorkerCount(),
		"atc_eip":                   atcPublicIP,
		"external_tls.certificate":  client.config.GetConcourseCert(),
		"external_tls.private_key":  client.config.GetConcourseKey(),
		"atc_encryption_key":        client.config.GetEncryptionKey(),
		"web_static_ip":             atcPrivateIP.String(),
		"enable_global_resources":   client.config.GetEnableGlobalResources(),
		"enable_pipeline_instances": client.config.GetEnablePipelineInstances(),
	}

	flagFiles := []string{
		client.workingdir.PathInWorkingDir(concourseManifestFilename),
		"--vars-store",
		client.workingdir.PathInWorkingDir(credsFilename),
		"--ops-file",
		client.workingdir.PathInWorkingDir(concourseVersionsFilename),
		"--ops-file",
		client.workingdir.PathInWorkingDir(concourseSHAsFilename),
		"--vars-file",
		client.workingdir.PathInWorkingDir(concourseGrafanaFilename),
	}

	if client.config.GetConcoursePassword() != "" {
		vmap["atc_password"] = client.config.GetConcoursePassword()
	}

	if client.config.IsBitbucketAuthSet() {
		vmap["bitbucket_client_id"] = client.config.GetBitbucketClientID()
		vmap["bitbucket_client_secret"] = client.config.GetBitbucketClientSecret()
		flagFiles = append(flagFiles, "--ops-file", client.workingdir.PathInWorkingDir(concourseBitBucketAuthFilename))
	}

	if client.config.IsGithubAuthSet() {
		vmap["github_client_id"] = client.config.GetGithubClientID()
		vmap["github_client_secret"] = client.config.GetGithubClientSecret()
		flagFiles = append(flagFiles, "--ops-file", client.workingdir.PathInWorkingDir(concourseGitHubAuthFilename))
	}

	if client.config.IsMicrosoftAuthSet() {
		vmap["microsoft_client_id"] = client.config.GetMicrosoftClientID()
		vmap["microsoft_client_secret"] = client.config.GetMicrosoftClientSecret()
		vmap["microsoft_tenant"] = client.config.GetMicrosoftTenant()
		flagFiles = append(flagFiles, "--ops-file", client.workingdir.PathInWorkingDir(concourseMicrosoftAuthFilename))
	}

	if client.config.IsSpot() {
		flagFiles = append(flagFiles, "--ops-file", client.workingdir.PathInWorkingDir(concourseEphemeralWorkersFilename))
	}

	t, err1 := client.buildTagsYaml(vmap["project"], "concourse")
	if err1 != nil {
		return creds, err
	}
	vmap["tags"] = t
	flagFiles = append(flagFiles, "--ops-file", client.workingdir.PathInWorkingDir(extraTagsFilename))

	vs := vars(vmap)

	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return creds, fmt.Errorf("failed to retrieve director IP: [%v]", err)
	}

	err = client.boshCLI.RunAuthenticatedCommand(
		"deploy",
		directorPublicIP,
		client.config.GetDirectorPassword(),
		client.config.GetDirectorCACert(),
		detach,
		os.Stdout,
		append(flagFiles, vs...)...)
	if err != nil {
		return creds, fmt.Errorf("failed to run bosh deploy with commands %+v: [%v]", flagFiles, err)
	}

	return ioutil.ReadFile(client.workingdir.PathInWorkingDir(credsFilename))
}

func (client *AWSClient) buildTagsYaml(project interface{}, component string) (string, error) {
	var b strings.Builder

	for _, e := range client.config.GetTags() {
		kv := strings.Join(strings.Split(e, "="), ": ")
		_, err := fmt.Fprintf(&b, "%s,", kv)
		if err != nil {
			return "", err
		}
	}
	cProjectTag := fmt.Sprintf("control-tower-project: %v,", project)
	b.WriteString(cProjectTag)
	cComponentTag := fmt.Sprintf("control-tower-component: %s", component)
	b.WriteString(cComponentTag)
	return fmt.Sprintf("{%s}", b.String()), nil
}
