package commands_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("commands", func() {
	Describe("deploy", func() {
		When("using --help", func() {
			It("displays usage details", func() {
				output, err := controlTowerCommand("deploy", "--help").CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("control-tower deploy - Deploys or updates a Concourse"))
				Expect(string(output)).To(ContainSubstring("--region value"))
				Expect(string(output)).To(ContainSubstring("--domain value"))
				Expect(string(output)).To(ContainSubstring("--tls-cert value"))
				Expect(string(output)).To(ContainSubstring("--tls-key value"))
				Expect(string(output)).To(ContainSubstring("--db-size value"))
				Expect(string(output)).To(MatchRegexp(`--vpc-network-range value\s+\(optional\) VPC network CIDR to deploy into, only required if IAAS is AWS`))
				Expect(string(output)).To(MatchRegexp(`--public-subnet-range value\s+\(optional\) public network CIDR \(if IAAS is AWS must be within --vpc-network-range\)`))
				Expect(string(output)).To(MatchRegexp(`--private-subnet-range value\s+\(optional\) private network CIDR \(if IAAS is AWS must be within --vpc-network-range\)`))
				Expect(string(output)).To(MatchRegexp(`--rds-subnet-range1 value\s+\(optional\) first rds network CIDR \(if IAAS is AWS must be within --vpc-network-range\)`))
				Expect(string(output)).To(MatchRegexp(`--rds-subnet-range2 value\s+\(optional\) second rds network CIDR \(if IAAS is AWS must be within --vpc-network-range\)`))
			})
		})

		When("the IAAS is not specified", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("deploy", "abc").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(output).To(MatchRegexp(`Error validating args on deploy: \[failed to validate Deploy flags: \[--iaas flag not set\]\]`))
			})
		})

		When("no name is passed in", func() {
			It("displays correct usage", func() {
				output, err := controlTowerCommand("deploy", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("Usage is `control-tower deploy <name>`"))
			})
		})

		When("there is a key but no cert", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("deploy", "abc", "--domain", "abc.engineerbetter.com", "--tls-key", "-- BEGIN RSA PRIVATE KEY --", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("--tls-key requires --tls-cert to also be provided"))
			})
		})

		When("there is a cert but no key", func() {
			It("Should show a meaningful error", func() {
				output, err := controlTowerCommand("deploy", "abc", "--domain", "abc.engineerbetter.com", "--tls-cert", "-- BEGIN CERTIFICATE --", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("--tls-cert requires --tls-key to also be provided"))
			})
		})

		When("there is a cert and key but no domain", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("deploy", "abc", "--tls-key", "-- BEGIN RSA PRIVATE KEY --", "--tls-cert", "-- BEGIN RSA PRIVATE KEY --", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("custom certificates require --domain to be provided"))
			})
		})

		When("an invalid worker count is provided", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("deploy", "abc", "--workers", "0", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("minimum number of workers is 1"))
			})
		})

		When("an invalid worker size is provided", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("deploy", "abc", "--worker-size", "small", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Eventually(string(output)).Should(ContainSubstring("unknown worker size"))
			})
		})

		When("an invalid web size is provided", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("deploy", "abc", "--web-size", "tiny", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Eventually(string(output)).Should(ContainSubstring("unknown web node size"))
			})
		})

		When("an invalid db size is provided", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("deploy", "abc", "--db-size", "huge", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Eventually(string(output)).Should(ContainSubstring("unknown DB size"))
			})
		})
	})

	Describe("destroy", func() {
		When("using --help", func() {
			It("displays usage details", func() {
				output, err := controlTowerCommand("destroy", "--help", "--iaas", "AWS").CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("control-tower destroy - Destroys a Concourse"))
			})
		})

		When("the IAAS is not specified", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("destroy", "abc").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(MatchRegexp(`Error validating args on destroy: \[failed to validate Destroy flags: \[--iaas flag not set\]\]`))
			})
		})

		When("no name is passed in", func() {
			It("displays correct usage", func() {
				output, err := controlTowerCommand("destroy", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("Usage is `control-tower destroy <name>`"))
			})
		})
	})

	Describe("info", func() {
		When("using --help", func() {
			It("displays usage details", func() {
				output, err := controlTowerCommand("info", "--help").CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("control-tower info - Fetches information on a deployed environment"))
			})
		})

		When("the IAAS is not specified", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("info", "abc").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(MatchRegexp(`Error validating args on info: \[failed to validate Info flags: \[--iaas flag not set\]\]`))
			})
		})

		When("no name is passed in", func() {
			It("displays correct usage", func() {
				output, err := controlTowerCommand("info", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("Usage is `control-tower info <name>`"))
			})
		})
	})

	Describe("maintain", func() {
		When("using --help", func() {
			It("displays usage details", func() {
				output, err := controlTowerCommand("maintain", "--help", "--iaas", "AWS").CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("control-tower maintain - Handles maintenance operations in control-tower"))
			})
		})

		When("the IAAS is not specified", func() {
			It("shows a meaningful error", func() {
				output, err := controlTowerCommand("maintain", "abc").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(MatchRegexp(`Error validating args on maintain: \[failed to validate Maintain flags: \[--iaas flag not set\]\]`))
			})
		})

		When("no name is passed in", func() {
			It("displays correct usage", func() {
				output, err := controlTowerCommand("maintain", "--iaas", "AWS").CombinedOutput()
				Expect(err).To(HaveOccurred(), string(output))
				Expect(string(output)).To(ContainSubstring("Usage is `control-tower maintain <name>`"))
			})
		})
	})
})
