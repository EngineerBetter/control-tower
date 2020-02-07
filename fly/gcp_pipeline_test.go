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
  source:
    user: engineerbetter
    repository: control-tower
    pre_release: true

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
`

		It("Generates something sensible", func() {
			tempFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			defer os.Remove(tempFile.Name()) // clean up

			_, err = tempFile.Write([]byte("creds-content"))
			Expect(err).ToNot(HaveOccurred())

			pipeline, err := NewGCPPipeline(tempFile.Name())
			Expect(err).ToNot(HaveOccurred())

			params, err := pipeline.BuildPipelineParams("my-deployment", "prod", "europe-west1", "ci.engineerbetter.com", "GCP")
			Expect(err).ToNot(HaveOccurred())

			yamlBytes, err := util.RenderTemplate("self-update pipeline", pipeline.GetConfigTemplate(), params)
			Expect(err).ToNot(HaveOccurred())

			actual := string(yamlBytes)
			Expect(actual).To(Equal(expected))
		})
	})
})
