package fly

import (
	"io/ioutil"
	"strings"
)

// GCPPipeline is GCP specific implementation of Pipeline interface
type GCPPipeline struct {
	PipelineTemplateParams
	GCPCreds string
}

// NewGCPPipeline return GCPPipeline
func NewGCPPipeline(credsPath string) (Pipeline, error) {
	creds, err := readFileContents(credsPath)
	if err != nil {
		return nil, err
	}
	return GCPPipeline{
		GCPCreds: creds,
	}, nil
}

//BuildPipelineParams builds params for AWS control-tower self update pipeline
func (a GCPPipeline) BuildPipelineParams(deployment, namespace, region, domain, allowIps, iaas string) (Pipeline, error) {
	return GCPPipeline{
		PipelineTemplateParams: PipelineTemplateParams{
			ControlTowerVersion: ControlTowerVersion,
			Deployment:          strings.TrimPrefix(deployment, "control-tower-"),
			AllowIPs:            allowIps,
			Domain:              domain,
			Namespace:           namespace,
			Region:              region,
			IaaS:                iaas,
		},
		GCPCreds: a.GCPCreds,
	}, nil
}

// GetConfigTemplate returns template for AWS Control-Tower self update pipeline
func (a GCPPipeline) GetConfigTemplate() string {
	return gcpPipelineTemplate

}

func readFileContents(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

const gcpPipelineTemplate = `
---` + selfUpdateResources + `
jobs:
- name: self-update
  serial_groups: [cup]
  serial: true
  plan:
  - get: control-tower-release
    trigger: true
  - task: update
    params:
      AWS_REGION: "{{ .Region }}"
      DEPLOYMENT: "{{ .Deployment }}"
      GCPCreds: '{{ .GCPCreds }}'
      IAAS: "{{ .IaaS }}"
      NAMESPACE: "{{ .Namespace }}"
      ALLOW_IPS: "{{ .AllowIPs }}"
      SELF_UPDATE: true
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: engineerbetter/pcf-ops
      inputs:
      - name: control-tower-release
      run:
        path: bash
        args:
        - -c
        - |
          cd control-tower-release
          echo "${GCPCreds}" > googlecreds.json
          export GOOGLE_APPLICATION_CREDENTIALS=$PWD/googlecreds.json
          set -eux
          chmod +x control-tower-linux-amd64
          ./control-tower-linux-amd64 deploy $DEPLOYMENT
- name: renew-https-cert
  serial_groups: [cup]
  serial: true
  plan:
  - get: control-tower-release
    version: {tag: "{{ .ControlTowerVersion }}" }
  - get: every-day
    trigger: true
  - task: update
    params:
      AWS_REGION: "{{ .Region }}"
      DEPLOYMENT: "{{ .Deployment }}"
      GCPCreds: '{{ .GCPCreds }}'
      IAAS: "{{ .IaaS }}"
      NAMESPACE: "{{ .Namespace }}"
      ALLOW_IPS: "{{ .AllowIPs }}"
      SELF_UPDATE: true
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: engineerbetter/pcf-ops
      inputs:
      - name: control-tower-release
      run:
        path: bash
        args:
        - -c
        - |
          echo "${GCPCreds}" > googlecreds.json
          export GOOGLE_APPLICATION_CREDENTIALS=$PWD/googlecreds.json
          set -euxo pipefail
          cd control-tower-release
          chmod +x control-tower-linux-amd64
` + renewCertsDateCheck + `
          echo Certificates expire in $days_until_expiry days, redeploying to renew them
          ./control-tower-linux-amd64 deploy $DEPLOYMENT
`
