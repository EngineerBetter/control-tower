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
      AWS_ACCESS_KEY_ID: "access-key"
      AWS_REGION: "eu-west-1"
      AWS_SECRET_ACCESS_KEY: "secret-key"
      DEPLOYMENT: "my-deployment"
      IAAS: "AWS"
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
          set -eux

          cd control-tower-release
          chmod +x control-tower-linux-amd64
          ./control-tower-linux-amd64 deploy $DEPLOYMENT
`

		It("Generates something sensible", func() {
			fakeCredsGetter := func() (string, string, error) {
				return "access-key", "secret-key", nil
			}

			pipeline := NewAWSPipeline(fakeCredsGetter)

			params, err := pipeline.BuildPipelineParams("my-deployment", "prod", "eu-west-1", "ci.engineerbetter.com", "AWS")
			Expect(err).ToNot(HaveOccurred())

			yamlBytes, err := util.RenderTemplate("self-update pipeline", pipeline.GetConfigTemplate(), params)
			Expect(err).ToNot(HaveOccurred())

			actual := string(yamlBytes)
			Expect(actual).To(Equal(expected))
		})
	})
})
