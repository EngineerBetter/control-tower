package fly

import (
	"strings"
)

// AWSPipeline is AWS specific implementation of Pipeline interface
type AWSPipeline struct {
	PipelineTemplateParams
}

// NewAWSPipeline return AWSPipeline
func NewAWSPipeline() Pipeline {
	return AWSPipeline{}
}

//BuildPipelineParams builds params for AWS control-tower self update pipeline
func (a AWSPipeline) BuildPipelineParams(deployment, namespace, region, domain, allowIps, iaas string) (Pipeline, error) {
	return AWSPipeline{
		PipelineTemplateParams: PipelineTemplateParams{
			ControlTowerVersion: ControlTowerVersion,
			Deployment:          strings.TrimPrefix(deployment, "control-tower-"),
			Domain:              domain,
			AllowIPs:            allowIps,
			Namespace:           namespace,
			Region:              region,
			IaaS:                iaas,
		},
	}, nil
}

// GetConfigTemplate returns template for AWS Control-Tower self update pipeline
func (a AWSPipeline) GetConfigTemplate() string {
	return awsPipelineTemplate

}

const awsPipelineTemplate = `
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
      AWS_ACCESS_KEY_ID: ((aws_access_key_id))
      AWS_REGION: "{{ .Region }}"
      AWS_SECRET_ACCESS_KEY: ((aws_secret_access_key))
      DEPLOYMENT: "{{ .Deployment }}"
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
          set -eux

          cd control-tower-release
          chmod +x control-tower-linux-amd64
          ./control-tower-linux-amd64 deploy $DEPLOYMENT
- name: renew-https-cert
  serial_groups: [cup]
  serial: true
  plan:
  - get: control-tower-release
    version: {tag: {{ .ControlTowerVersion }} }
  - get: every-day
    trigger: true
  - task: update
    params:
      AWS_ACCESS_KEY_ID: ((aws_access_key_id))
      AWS_REGION: "{{ .Region }}"
      AWS_SECRET_ACCESS_KEY: ((aws_secret_access_key))
      DEPLOYMENT: "{{ .Deployment }}"
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
          set -euxo pipefail
          cd control-tower-release
          chmod +x control-tower-linux-amd64
` + renewCertsDateCheck + `
          echo Certificates expire in $days_until_expiry days, redeploying to renew them
          ./control-tower-linux-amd64 deploy $DEPLOYMENT
`
