package fly_test

import (
	_ "embed"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/util"
)

//go:embed fixtures/gcp-self-update-pipeline.yaml
var expectedGCP string

var _ = Describe("GCPPipeline", func() {
	Describe("Generating a pipeline YAML", func() {
		It("Generates something sensible", func() {
			pipeline := NewGCPPipeline()

			params, err := pipeline.BuildPipelineParams("my-deployment", "prod", "europe-west1", "ci.engineerbetter.com", "10.0.0.0", "GCP")
			Expect(err).ToNot(HaveOccurred())

			yamlBytes, err := util.RenderTemplate("self-update pipeline", pipeline.GetConfigTemplate(), params)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(yamlBytes)).To(Equal(expectedGCP))
		})
	})
})
