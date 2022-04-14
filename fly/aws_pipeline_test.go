package fly_test

import (
	_ "embed"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/util"
)

//go:embed fixtures/aws-self-update-pipeline.yaml
var expectedAWS string

var _ = Describe("AWSPipeline", func() {
	Describe("Generating a pipeline YAML", func() {
		It("Generates something sensible", func() {

			pipeline := NewAWSPipeline()

			params, err := pipeline.BuildPipelineParams("my-deployment", "prod", "eu-west-1", "ci.engineerbetter.com", "10.0.0.0", "AWS")
			Expect(err).ToNot(HaveOccurred())

			yamlBytes, err := util.RenderTemplate("self-update pipeline", pipeline.GetConfigTemplate(), params)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(yamlBytes)).To(Equal(expectedAWS))
		})
	})
})
