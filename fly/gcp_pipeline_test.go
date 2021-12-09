package fly_test

import (
	_ "embed"
	"io"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/util"
)

//go:embed fixtures/gcp-self-update-pipeline.yaml
var expectedGCP string

var _ = Describe("GCPPipeline", func() {
	Describe("Generating a pipeline YAML", func() {
		It("Generates something sensible", func() {
			tempFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			defer os.Remove(tempFile.Name()) // clean up

			_, err = io.WriteString(tempFile, "creds-content")
			Expect(err).ToNot(HaveOccurred())

			pipeline, err := NewGCPPipeline(tempFile.Name())
			Expect(err).ToNot(HaveOccurred())

			params, err := pipeline.BuildPipelineParams("my-deployment", "prod", "europe-west1", "ci.engineerbetter.com", "10.0.0.0", "GCP")
			Expect(err).ToNot(HaveOccurred())

			yamlBytes, err := util.RenderTemplate("self-update pipeline", pipeline.GetConfigTemplate(), params)
			Expect(err).ToNot(HaveOccurred())

			Expect(string(yamlBytes)).To(Equal(expectedGCP))
		})
	})
})
