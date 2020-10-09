package fly_test

import (
	. "github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWSPipeline", func() {
	Describe("Generating a pipeline YAML", func() {
		var expected = `
---
resources:
- name: control-tower-release
  type: github-release
  icon: github
  source:
    user: engineerbetter
    repository: control-tower
    pre_release: true
- name: every-day
  type: time
  icon: clock
  source: {interval: 24h}

jobs:
- name: self-update
  serial_groups: [cup]
  serial: true
  plan:
  - get: control-tower-release
    trigger: true
  - task: update
    params:
      AWS_ACCESS_KEY_ID: "access-key"
      AWS_REGION: "eu-west-1"
      AWS_SECRET_ACCESS_KEY: "secret-key"
      DEPLOYMENT: "my-deployment"
      IAAS: "AWS"
      NAMESPACE: "prod"
      ALLOW_IPS: "10.0.0.0"
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
    version: {tag: COMPILE_TIME_VARIABLE_fly_control_tower_version }
  - get: every-day
    trigger: true
  - task: update
    params:
      AWS_ACCESS_KEY_ID: "access-key"
      AWS_REGION: "eu-west-1"
      AWS_SECRET_ACCESS_KEY: "secret-key"
      DEPLOYMENT: "my-deployment"
      IAAS: "AWS"
      NAMESPACE: "prod"
      ALLOW_IPS: "10.0.0.0"
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

          now_seconds=$(date +%s)
          not_after=$(echo | openssl s_client -connect ci.engineerbetter.com:443 2>/dev/null | openssl x509 -noout -enddate)
          expires_on=${not_after#'notAfter='}
          expires_on_seconds=$(date --date="$expires_on" +%s)
          let "seconds_until_expiry = $expires_on_seconds - $now_seconds"
          let "days_until_expiry = $seconds_until_expiry / 60 / 60 / 24"
          if [ $days_until_expiry -gt 2 ]; then
            echo Not renewing HTTPS cert, as they do not expire in the next two days.
            exit 0
          fi

          echo Certificates expire in $days_until_expiry days, redeploying to renew them
          ./control-tower-linux-amd64 deploy $DEPLOYMENT
`

		It("Generates something sensible", func() {
			fakeCredsGetter := func() (string, string, error) {
				return "access-key", "secret-key", nil
			}

			pipeline := NewAWSPipeline(fakeCredsGetter)

			params, err := pipeline.BuildPipelineParams("my-deployment", "prod", "eu-west-1", "ci.engineerbetter.com", "10.0.0.0", "AWS")
			Expect(err).ToNot(HaveOccurred())

			yamlBytes, err := util.RenderTemplate("self-update pipeline", pipeline.GetConfigTemplate(), params)
			Expect(err).ToNot(HaveOccurred())

			actual := string(yamlBytes)
			Expect(actual).To(Equal(expected))
		})
	})
})
