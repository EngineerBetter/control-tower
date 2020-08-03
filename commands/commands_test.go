package commands

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var (
	cliPath string
)

var _ = Describe("commands", func() {
	BeforeSuite(func() {
		var err error
		cliPath, err = Build("github.com/EngineerBetter/control-tower")
		Expect(err).ToNot(HaveOccurred(), "Error building source")
	})

	AfterSuite(func() {
		CleanupBuildArtifacts()
	})

	Describe("deploy", func() {
		Context("When using --help", func() {
			It("should display usage details", func() {
				command := exec.Command(cliPath, "deploy", "--help")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred(), "Error running CLI: "+cliPath)
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say("control-tower deploy - Deploys or updates a Concourse"))
				Expect(session.Out).To(Say("--region value"))
				Expect(session.Out).To(Say("--domain value"))
				Expect(session.Out).To(Say("--tls-cert value"))
				Expect(session.Out).To(Say("--tls-key value"))
				Expect(session.Out).To(Say("--db-size value"))
				Expect(session.Out).To(Say("--vpc-network-range value\\s+\\(optional\\) VPC network CIDR to deploy into, only required if IAAS is AWS"))
				Expect(session.Out).To(Say("--public-subnet-range value\\s+\\(optional\\) public network CIDR \\(if IAAS is AWS must be within --vpc-network-range\\)"))
				Expect(session.Out).To(Say("--private-subnet-range value\\s+\\(optional\\) private network CIDR \\(if IAAS is AWS must be within --vpc-network-range\\)"))
				Expect(session.Out).To(Say("--rds-subnet-range1 value\\s+\\(optional\\) first rds network CIDR \\(if IAAS is AWS must be within --vpc-network-range\\)"))
				Expect(session.Out).To(Say("--rds-subnet-range2 value\\s+\\(optional\\) second rds network CIDR \\(if IAAS is AWS must be within --vpc-network-range\\)"))
			})
		})

		Context("When the IAAS is not specified", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				// Say takes a regexp so `[` and `]` need to be escaped
				Expect(session.Err).To(Say("Error validating args on deploy: \\[failed to validate Deploy flags: \\[--iaas flag not set\\]\\]"))
			})
		})

		Context("When no name is passed in", func() {
			It("should display correct usage", func() {
				command := exec.Command(cliPath, "deploy", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Usage is `control-tower deploy <name>`"))
			})
		})

		Context("When there is a key but no cert", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--domain", "abc.engineerbetter.com", "--tls-key", "-- BEGIN RSA PRIVATE KEY --", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("--tls-key requires --tls-cert to also be provided"))
			})
		})

		Context("When there is a cert but no key", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--domain", "abc.engineerbetter.com", "--tls-cert", "-- BEGIN CERTIFICATE --", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("--tls-cert requires --tls-key to also be provided"))
			})
		})

		Context("When there is a cert and key but no domain", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--tls-key", "-- BEGIN RSA PRIVATE KEY --", "--tls-cert", "-- BEGIN RSA PRIVATE KEY --", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("custom certificates require --domain to be provided"))
			})
		})

		Context("When an invalid worker count is provided", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--workers", "0", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("minimum number of workers is 1"))
			})
		})

		Context("When an invalid worker size is provided", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--worker-size", "small", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("unknown worker size"))
			})
		})

		Context("When an invalid web size is provided", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--web-size", "tiny", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("unknown web node size"))
			})
		})

		Context("When an invalid db size is provided", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "deploy", "abc", "--db-size", "huge", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("unknown DB size"))
			})
		})
	})

	Describe("destroy", func() {
		Context("When using --help", func() {
			It("should display usage details", func() {
				command := exec.Command(cliPath, "destroy", "--help", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred(), "Error running CLI: "+cliPath)
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say("control-tower destroy - Destroys a Concourse"))
			})
		})

		Context("When the IAAS is not specified", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "destroy", "abc")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				// Say takes a regexp so `[` and `]` need to be escaped
				Expect(session.Err).To(Say("Error validating args on destroy: \\[failed to validate Destroy flags: \\[--iaas flag not set\\]\\]"))
			})
		})

		Context("When no name is passed in", func() {
			It("should display correct usage", func() {
				command := exec.Command(cliPath, "destroy", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Usage is `control-tower destroy <name>`"))
			})
		})
	})

	Describe("info", func() {
		Context("When using --help", func() {
			It("should display usage details", func() {
				command := exec.Command(cliPath, "info", "--help")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred(), "Error running CLI: "+cliPath)
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say("control-tower info - Fetches information on a deployed environment"))
			})
		})

		Context("When the IAAS is not specified", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "info", "abc")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				// Say takes a regexp so `[` and `]` need to be escaped
				Expect(session.Err).To(Say("Error validating args on info: \\[failed to validate Info flags: \\[--iaas flag not set\\]\\]"))
			})
		})

		Context("When no name is passed in", func() {
			It("should display correct usage", func() {
				command := exec.Command(cliPath, "info", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Usage is `control-tower info <name>`"))
			})
		})
	})

	Describe("maintain", func() {
		Context("When using --help", func() {
			It("should display usage details", func() {
				command := exec.Command(cliPath, "maintain", "--help", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred(), "Error running CLI: "+cliPath)
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say("control-tower maintain - Handles maintenance operations in control-tower"))
			})
		})

		Context("When the IAAS is not specified", func() {
			It("Should show a meaningful error", func() {
				command := exec.Command(cliPath, "maintain", "abc")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				// Say takes a regexp so `[` and `]` need to be escaped
				Expect(session.Err).To(Say("Error validating args on maintain: \\[failed to validate Maintain flags: \\[--iaas flag not set\\]\\]"))
			})
		})

		Context("When no name is passed in", func() {
			It("should display correct usage", func() {
				command := exec.Command(cliPath, "maintain", "--iaas", "AWS")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Usage is `control-tower maintain <name>`"))
			})
		})
	})
})
