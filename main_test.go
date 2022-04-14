package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("control-tower", func() {
	It("displays usage instructions on --help", func() {
		output, err := exec.Command("go", "run", "github.com/EngineerBetter/control-tower", "--help").CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		outputStr := string(output)
		Expect(outputStr).To(ContainSubstring("Control-Tower - A CLI tool to deploy Concourse CI"), outputStr)
		Expect(outputStr).To(ContainSubstring("deploy, d    Deploys or updates a Concourse"), outputStr)
		Expect(outputStr).To(ContainSubstring("destroy, x   Destroys a Concourse"), outputStr)
		Expect(outputStr).To(ContainSubstring("info, i      Fetches information on a deployed environment"), outputStr)
		Expect(outputStr).To(ContainSubstring("maintain, m  Handles maintenance operations in control-tower"), outputStr)
	})
})
