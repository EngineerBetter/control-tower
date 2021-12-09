package fly_test

import (
	"io/ioutil"
	"os"

	. "github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPPipeline", func() {
	Describe("Generating a pipeline YAML", func() {
		var expected = `
---
resources:
- name: control-tower-release
  type: github-release
  icon: github
  check_every: 30m
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
      AWS_REGION: "europe-west1"
      DEPLOYMENT: "my-deployment"
      GCPCreds: 'creds-content'
      IAAS: "GCP"
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
    version: {tag: "COMPILE_TIME_VARIABLE_fly_control_tower_version" }
  - get: every-day
    trigger: true
  - task: update
    params:
      AWS_REGION: "europe-west1"
      DEPLOYMENT: "my-deployment"
      GCPCreds: 'creds-content'
      IAAS: "GCP"
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
          echo "${GCPCreds}" > googlecreds.json
          export GOOGLE_APPLICATION_CREDENTIALS=$PWD/googlecreds.json
          set -euxo pipefail
          cd control-tower-release
          chmod +x control-tower-linux-amd64

          now_seconds=$(date +%s)
          not_after=$(echo | openssl s_client -connect ci.engineerbetter.com:443 2>/dev/null | openssl x509 -noout -enddate)
          expires_on=${not_after#'notAfter='}
          expires_on_seconds=$(date --date="$expires_on" +%s)

          # "let" fails when the result is 0
          set +e
          let "seconds_until_expiry = $expires_on_seconds - $now_seconds"
          let "days_until_expiry = $seconds_until_expiry / 60 / 60 / 24"
          set -e

          if [ $days_until_expiry -gt 2 ]; then
            echo Not renewing HTTPS cert, as they do not expire in the next two days.
            exit 0
          fi

          echo Certificates expire in $days_until_expiry days, redeploying to renew them
          ./control-tower-linux-amd64 deploy $DEPLOYMENT
`

		It("Generates something sensible", func() {
			tempFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			defer os.Remove(tempFile.Name()) // clean up

			_, err = tempFile.Write([]byte("creds-content"))
			Expect(err).ToNot(HaveOccurred())

			pipeline, err := NewGCPPipeline(tempFile.Name())
			Expect(err).ToNot(HaveOccurred())

			params, err := pipeline.BuildPipelineParams("my-deployment", "prod", "europe-west1", "ci.engineerbetter.com", "10.0.0.0", "GCP")
			Expect(err).ToNot(HaveOccurred())

			yamlBytes, err := util.RenderTemplate("self-update pipeline", pipeline.GetConfigTemplate(), params)
			Expect(err).ToNot(HaveOccurred())

			actual := string(yamlBytes)
			Expect(actual).To(Equal(expected))
		})
	})
})
