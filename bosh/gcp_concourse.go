package bosh

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func (client *GCPClient) deployConcourse(creds []byte, detach bool) ([]byte, error) {

	err := saveFilesToWorkingDir(client.workingdir, client.provider, creds)
	if err != nil {
		return nil, fmt.Errorf("failed saving files to working directory in deployConcourse: [%v]", err)
	}

	uaaCertPath, err := client.workingdir.SaveFileToWorkingDir(uaaCertFilename, uaaCert)
	if err != nil {
		return []byte{}, err
	}

	boshDBAddress, err := client.outputs.Get("BoshDBAddress")
	if err != nil {
		return []byte{}, err
	}
	atcPublicIP, err := client.outputs.Get("ATCPublicIP")
	if err != nil {
		return []byte{}, err
	}

	networkName, err := client.outputs.Get("Network")
	if err != nil {
		return []byte{}, err
	}

	SQLServerCert, err := client.outputs.Get("SQLServerCert")
	if err != nil {
		return []byte{}, err
	}

	vmap := map[string]interface{}{
		"deployment_name":          concourseDeploymentName,
		"domain":                   client.config.GetDomain(),
		"project":                  client.config.GetProject(),
		"web_network_name":         "public",
		"worker_network_name":      "private",
		"postgres_host":            boshDBAddress,
		"postgres_role":            client.config.GetRDSUsername(),
		"postgres_port":            "5432",
		"postgres_password":        client.config.GetRDSPassword(),
		"postgres_ca_cert":         SQLServerCert,
		"web_vm_type":              "concourse-web-" + client.config.GetConcourseWebSize(),
		"worker_vm_type":           "concourse-" + client.config.GetConcourseWorkerSize(),
		"worker_count":             client.config.GetConcourseWorkerCount(),
		"atc_eip":                  atcPublicIP,
		"external_tls.certificate": client.config.GetConcourseCert(),
		"external_tls.private_key": client.config.GetConcourseKey(),
		"atc_encryption_key":       client.config.GetEncryptionKey(),
		"network_name":             networkName,
	}

	flagFiles := []string{
		client.workingdir.PathInWorkingDir(concourseManifestFilename),
		"--vars-store",
		client.workingdir.PathInWorkingDir(credsFilename),
		"--ops-file",
		client.workingdir.PathInWorkingDir(concourseVersionsFilename),
		"--ops-file",
		client.workingdir.PathInWorkingDir(concourseSHAsFilename),
		"--ops-file",
		client.workingdir.PathInWorkingDir(concourseCompatibilityFilename),
		"--ops-file",
		uaaCertPath,
		"--vars-file",
		client.workingdir.PathInWorkingDir(concourseGrafanaFilename),
	}

	if client.config.GetConcoursePassword() != "" {
		vmap["atc_password"] = client.config.GetConcoursePassword()
	}

	if client.config.IsGithubAuthSet() {
		vmap["github_client_id"] = client.config.GetGithubClientID()
		vmap["github_client_secret"] = client.config.GetGithubClientSecret()
		flagFiles = append(flagFiles, "--ops-file", client.workingdir.PathInWorkingDir(concourseGitHubAuthFilename))
	}

	t, err1 := client.buildTagsYaml(vmap["project"], "concourse")
	if err1 != nil {
		return nil, err
	}
	vmap["tags"] = t
	flagFiles = append(flagFiles, "--ops-file", client.workingdir.PathInWorkingDir(extraTagsFilename))

	vs := vars(vmap)

	directorPublicIP, err := client.outputs.Get("DirectorPublicIP")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve director IP: [%v]", err)
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
		return nil, fmt.Errorf("failed to run bosh deploy with commands %+v: [%v]", flagFiles, err)
	}

	return ioutil.ReadFile(client.workingdir.PathInWorkingDir(credsFilename))
}

func (client *GCPClient) buildTagsYaml(project interface{}, component string) (string, error) {
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
