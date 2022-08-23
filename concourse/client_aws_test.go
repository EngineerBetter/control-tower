package concourse_test

import (
	_ "embed"
	"errors"
	"fmt"
	"io"

	"github.com/go-acme/lego/v4/lego"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/EngineerBetter/control-tower/bosh"
	"github.com/EngineerBetter/control-tower/bosh/boshfakes"
	"github.com/EngineerBetter/control-tower/certs"
	"github.com/EngineerBetter/control-tower/certs/certsfakes"
	"github.com/EngineerBetter/control-tower/commands/deploy"
	"github.com/EngineerBetter/control-tower/concourse"
	"github.com/EngineerBetter/control-tower/concourse/concoursefakes"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/config/configfakes"
	"github.com/EngineerBetter/control-tower/credhub"
	"github.com/EngineerBetter/control-tower/credhub/credhubfakes"
	"github.com/EngineerBetter/control-tower/fly"
	"github.com/EngineerBetter/control-tower/fly/flyfakes"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/iaas/iaasfakes"
	"github.com/EngineerBetter/control-tower/terraform"
	"github.com/EngineerBetter/control-tower/terraform/terraformfakes"
)

//go:embed fixtures/director-state.json
var directorStateFixture []byte

//go:embed fixtures/director-creds.yml
var directorCredsFixture []byte

//go:embed fixtures/private-key.pem
var privateKeyFixture string

var _ = Describe("client", func() {
	var buildClient func() concourse.IClient
	var actions []string
	var stdout *gbytes.Buffer
	var stderr *gbytes.Buffer
	var args *deploy.Args
	var configInBucket, configAfterLoad, configAfterCreateEnv config.Config
	var ipChecker func() (string, error)
	var tfInputVarsFactory *concoursefakes.FakeTFInputVarsFactory
	var flyClient *flyfakes.FakeIClient
	var terraformCLI *terraformfakes.FakeCLIInterface
	var configClient *configfakes.FakeIClient
	var boshClient *boshfakes.FakeIClient
	var credhubClient *credhubfakes.FakeIClient

	var setupFakeAwsProvider = func() *iaasfakes.FakeProvider {
		provider := &iaasfakes.FakeProvider{}
		provider.DBTypeReturns("db.t3.small")
		provider.RegionReturns("eu-west-1")
		provider.IAASReturns(iaas.AWS)
		provider.CheckForWhitelistedIPStub = func(ip, securityGroup string) (bool, error) {
			actions = append(actions, "checking security group for IP")
			if ip == "1.2.3.4" {
				return false, nil
			}
			return true, nil
		}
		provider.DeleteVMsInVPCStub = func(vpcID string) ([]string, error) {
			actions = append(actions, fmt.Sprintf("deleting vms in %s", vpcID))
			return nil, nil
		}
		provider.FindLongestMatchingHostedZoneStub = func(subdomain string) (string, string, error) {
			if subdomain == "ci.google.com" {
				return "google.com", "ABC123", nil
			}

			return "", "", errors.New("hosted zone not found")
		}
		return provider
	}

	var setupFakeTfInputVarsFactory = func() *concoursefakes.FakeTFInputVarsFactory {
		tfInputVarsFactory = &concoursefakes.FakeTFInputVarsFactory{}

		provider, err := iaas.New(iaas.AWS, "eu-west-1")
		Expect(err).ToNot(HaveOccurred())
		awsInputVarsFactory, err := concourse.NewTFInputVarsFactory(provider)
		Expect(err).ToNot(HaveOccurred())
		tfInputVarsFactory.NewInputVarsStub = func(i config.ConfigView) terraform.InputVars {
			actions = append(actions, "converting config.Config to TFInputVars")
			return awsInputVarsFactory.NewInputVars(i)
		}
		return tfInputVarsFactory
	}

	var setupFakeConfigClient = func() *configfakes.FakeIClient {
		configClient = &configfakes.FakeIClient{}
		configClient.LoadStub = func() (config.Config, error) {
			actions = append(actions, "loading config file")
			return configInBucket, nil
		}
		configClient.UpdateStub = func(config config.Config) error {
			actions = append(actions, "updating config file")
			return nil
		}
		configClient.StoreAssetStub = func(filename string, contents []byte) error {
			actions = append(actions, fmt.Sprintf("storing config asset: %s", filename))
			return nil
		}
		configClient.DeleteAllStub = func(config config.ConfigView) error {
			actions = append(actions, "deleting config")
			return nil
		}
		configClient.ConfigExistsStub = func() (bool, error) {
			actions = append(actions, "checking to see if config exists")
			return true, nil
		}
		return configClient
	}

	var setupFakeTerraformCLI = func(terraformOutputs terraform.AWSOutputs) *terraformfakes.FakeCLIInterface {
		terraformCLI = &terraformfakes.FakeCLIInterface{}
		terraformCLI.ApplyStub = func(inputVars terraform.InputVars) error {
			actions = append(actions, "applying terraform")
			return nil
		}
		terraformCLI.DestroyStub = func(conf terraform.InputVars) error {
			actions = append(actions, "destroying terraform")
			return nil
		}
		terraformCLI.BuildOutputStub = func(conf terraform.InputVars) (terraform.Outputs, error) {
			actions = append(actions, "initializing terraform outputs")
			return &terraformOutputs, nil
		}
		return terraformCLI
	}

	BeforeEach(func() {
		certGenerator := func(c func(u *certs.User) (*lego.Client, error), caName string, provider iaas.Provider, ip ...string) (*certs.Certs, error) {
			actions = append(actions, fmt.Sprintf("generating cert ca: %s, cn: %s", caName, ip))
			return &certs.Certs{
				CACert: []byte("----EXAMPLE CERT----"),
			}, nil
		}

		awsClient := setupFakeAwsProvider()
		tfInputVarsFactory = setupFakeTfInputVarsFactory()
		configClient = setupFakeConfigClient()

		flyClient = &flyfakes.FakeIClient{}
		flyClient.SetDefaultPipelineStub = func(config config.ConfigView, allowFlyVersionDiscrepancy bool) error {
			actions = append(actions, "setting default pipeline")
			return nil
		}
		credhubClient = &credhubfakes.FakeIClient{}

		args = &deploy.Args{
			AllowIPs:    "0.0.0.0/0",
			DBSize:      "small",
			DBSizeIsSet: false,
		}

		terraformOutputs := terraform.AWSOutputs{
			ATCPublicIP:              terraform.MetadataStringValue{Value: "77.77.77.77"},
			ATCSecurityGroupID:       terraform.MetadataStringValue{Value: "sg-999"},
			BlobstoreBucket:          terraform.MetadataStringValue{Value: "blobs.aws.com"},
			BlobstoreSecretAccessKey: terraform.MetadataStringValue{Value: "abc123"},
			BlobstoreUserAccessKeyID: terraform.MetadataStringValue{Value: "abc123"},
			BoshDBAddress:            terraform.MetadataStringValue{Value: "rds.aws.com"},
			BoshDBPort:               terraform.MetadataStringValue{Value: "5432"},
			BoshSecretAccessKey:      terraform.MetadataStringValue{Value: "abc123"},
			BoshUserAccessKeyID:      terraform.MetadataStringValue{Value: "abc123"},
			DirectorKeyPair:          terraform.MetadataStringValue{Value: "-- KEY --"},
			DirectorPublicIP:         terraform.MetadataStringValue{Value: "99.99.99.99"},
			DirectorSecurityGroupID:  terraform.MetadataStringValue{Value: "sg-123"},
			NatGatewayIP:             terraform.MetadataStringValue{Value: "88.88.88.88"},
			PrivateSubnetID:          terraform.MetadataStringValue{Value: "sn-private-123"},
			PublicSubnetID:           terraform.MetadataStringValue{Value: "sn-public-123"},
			VMsSecurityGroupID:       terraform.MetadataStringValue{Value: "sg-456"},
			VPCID:                    terraform.MetadataStringValue{Value: "vpc-112233"},
		}

		actions = []string{}
		configInBucket = config.Config{
			AvailabilityZone:         "eu-west-1a",
			ConcoursePassword:        "s3cret",
			ConcourseUsername:        "admin",
			ConcourseWebSize:         "medium",
			ConcourseWorkerCount:     1,
			ConcourseWorkerSize:      "large",
			Deployment:               "control-tower-happymeal",
			DirectorHMUserPassword:   "original-password",
			DirectorMbusPassword:     "original-password",
			DirectorNATSPassword:     "original-password",
			DirectorPassword:         "secret123",
			DirectorRegistryPassword: "original-password",
			DirectorUsername:         "admin",
			EncryptionKey:            "123456789a123456789b123456789c",
			IAAS:                     "AWS",
			PrivateKey:               privateKeyFixture,
			Project:                  "happymeal",
			PublicKey:                "example-public-key",
			RDSDefaultDatabaseName:   "bosh_abcdefgh",
			RDSInstanceClass:         "db.t3.medium",
			RDSPassword:              "s3cret",
			RDSUsername:              "admin",
			Region:                   "eu-west-1",
			Spot:                     true,
			TFStatePath:              "example-path",
			//These come from fixtures/director-creds.yml
			CredhubUsername:          "credhub-cli",
			CredhubPassword:          "f4b12bc0166cad1bc02b050e4e79ac4c",
			CredhubAdminClientSecret: "hxfgb56zny2yys6m9wjx",
			CredhubCACert:            "-----BEGIN CERTIFICATE-----\nMIIEXTCCAsWgAwIBAgIQSmhcetyHDHLOYGaqMnJ0QTANBgkqhkiG9w0BAQsFADA4\nMQwwCgYDVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxEDAOBgNVBAMM\nB2Jvc2hfY2EwHhcNMTkwMjEzMTAyNTM0WhcNMjAwMjEzMTAyNTM0WjA4MQwwCgYD\nVQQGEwNVU0ExFjAUBgNVBAoTDUNsb3VkIEZvdW5kcnkxEDAOBgNVBAMMB2Jvc2hf\nY2EwggGiMA0GCSqGSIb3DQEBAQUAA4IBjwAwggGKAoIBgQC+0bA9T4awlJYSn6aq\nun6Hylu47b2UiZpFZpvPomKWPay86QaJ0vC9SK8keoYI4gWwsZSAMXp2mSCkXKRi\n+rVc+sKnzv9VgPoVY5eYIYCtJvl7KCJQE02dGoxuGOaWlBiHuD6TzY6lI9fNxkAW\neMGR3UylJ7ET0NvgAZWS1daov2GfiKkaYUCdbY8DtfhMyFhJ381VNHwoP6xlZbSf\nTInO/2TS8xpW2BcMNhFAu9MJVtC5pDHtJtkXHXep027CkrPjtFQWpzvIMvPAtZ68\n9t46nS9Ix+RmeN3v+sawNzbZscnsslhB+m4GrpL9M8g8sbweMw9yxf241z1qkiNJ\nto3HRqqyNyGsvI9n7OUrZ4D5oAfY7ze1TF+nxnkmJp14y21FEdG7t76N0J5dn6bJ\n/lroojig/PqabRsyHbmj6g8N832PEQvwsPptihEwgrRmY6fcBbMUaPCpNuVTJVa5\ng0KdBGDYDKTMlEn4xaj8P1wRbVjtXVMED2l4K4tS/UiDIb8CAwEAAaNjMGEwDgYD\nVR0PAQH/BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFHii4fiqAwJS\nnNhi6C+ibr/4OOTyMB8GA1UdIwQYMBaAFHii4fiqAwJSnNhi6C+ibr/4OOTyMA0G\nCSqGSIb3DQEBCwUAA4IBgQAGXDTlsQWIJHfvU3zy9te35adKOUeDwk1lSe4NYvgW\nFJC0w2K/1ZldmQ2leHmiXSukDJAYmROy9Y1qkUazTzjsdvHGhUF2N1p7fIweNj8e\ncsR+T21MjPEwD99m5+xLvnMRMuqzH9TqVbFIM3lmCDajh8n9cp4KvGkQmB+X7DE1\nR6AXG4EN9xn91TFrqmFFNOrFtoAjtag05q/HoqMhFFVeg+JTpsPshFjlWIkzwqKx\npn68KG2ztgS0KeDraGKwItTKengTCr/VkgorXnhKcI1C6C5iRXZp3wREu8RO+wRe\nKSGbsYIHaFxd3XwW4JnsW+hes/W5MZX01wkwOLrktf85FjssBZBavxBbyFag/LvS\n8oULOZRLYUkuElM+0Wzf8ayB574Fd97gzCVzWoD0Ei982jAdbEfk77PV1TvMNmEn\n3M6ktB7GkjuD9OL12iNzxmbQe7p1WkYYps9hK4r0pbyxZPZlPMmNNZo579rywDjF\nwEW5QkylaPEkbVDhJWeR1I8=\n-----END CERTIFICATE-----\n",
			VMProvisioningType:       "spot",
			WorkerType:               "m4",
		}

		//Mutations we expect to have been done after load
		configAfterLoad = configInBucket
		configAfterLoad.AllowIPs = `"0.0.0.0/0"`
		configAfterLoad.AllowIPsUnformatted = "0.0.0.0/0"
		configAfterLoad.SourceAccessIP = "192.0.2.0"
		configAfterLoad.NetworkCIDR = "10.0.0.0/16"
		configAfterLoad.PublicCIDR = "10.0.0.0/24"
		configAfterLoad.PrivateCIDR = "10.0.1.0/24"
		configAfterLoad.RDS1CIDR = "10.0.4.0/24"
		configAfterLoad.RDS2CIDR = "10.0.5.0/24"

		//Mutations we expect to have been done after Deploy
		configAfterCreateEnv = configAfterLoad
		configAfterCreateEnv.ConcourseCACert = "----EXAMPLE CERT----"
		configAfterCreateEnv.DirectorCACert = "----EXAMPLE CERT----"
		configAfterCreateEnv.DirectorPublicIP = "99.99.99.99"
		configAfterCreateEnv.Domain = "77.77.77.77"
		configAfterCreateEnv.Tags = []string{"control-tower-version=some version"}
		configAfterCreateEnv.Version = "some version"

		terraformCLI = setupFakeTerraformCLI(terraformOutputs)

		boshClientFactory := func(config config.ConfigView, outputs terraform.Outputs, stdout, stderr io.Writer, provider iaas.Provider, versionFile []byte) (bosh.IClient, error) {
			boshClient = &boshfakes.FakeIClient{}
			boshClient.DeployStub = func(stateFileBytes, credsFileBytes []byte, detach bool) ([]byte, []byte, error) {
				if detach {
					actions = append(actions, "deploying director in self-update mode")
				} else {
					actions = append(actions, "deploying director")
				}
				return directorStateFixture, directorCredsFixture, nil
			}
			boshClient.CleanupStub = func() error {
				actions = append(actions, "cleaning up bosh init")
				return nil
			}
			boshClient.InstancesStub = func() ([]bosh.Instance, error) {
				actions = append(actions, "listing bosh instances")
				return nil, nil
			}

			return boshClient, nil
		}

		ipChecker = func() (string, error) {
			return "192.0.2.0", nil
		}

		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()

		versionFile := []byte("some versions")

		buildClient = func() concourse.IClient {
			return concourse.NewClient(
				awsClient,
				terraformCLI,
				tfInputVarsFactory,
				boshClientFactory,
				func(iaas.Provider, fly.Credentials, io.Writer, io.Writer, []byte) (fly.IClient, error) {
					return flyClient, nil
				},
				certGenerator,
				configClient,
				args,
				stdout,
				stderr,
				ipChecker,
				certsfakes.NewFakeAcmeClient,
				func(size int) string { return fmt.Sprintf("generatedPassword%d", size) },
				func() string { return "8letters" },
				func() ([]byte, []byte, string, error) { return []byte("private"), []byte("public"), "fingerprint", nil },
				"some version",
				versionFile,
				func(server, id, secret, cert string) (credhub.IClient, error) {
					return credhubClient, nil
				},
			)
		}
	})

	Describe("Destroy", func() {
		It("Loads the config file", func() {
			Expect(buildClient().Destroy()).To(Succeed())
			Expect(actions).To(ContainElement("loading config file"))
		})

		It("Builds IAAS environment", func() {
			Expect(buildClient().Destroy()).To(Succeed())
			Expect(tfInputVarsFactory.NewInputVarsCallCount()).To(Equal(1))
			Expect(tfInputVarsFactory.NewInputVarsArgsForCall(0)).To(Equal(configInBucket))
		})

		It("Loads terraform output", func() {
			Expect(buildClient().Destroy()).To(Succeed())
			Expect(actions).To(ContainElement("initializing terraform outputs"))
		})

		It("Deletes the vms in the vpcs", func() {
			Expect(buildClient().Destroy()).To(Succeed())
			Expect(actions).To(ContainElement("deleting vms in vpc-112233"))
		})

		It("Destroys the terraform infrastructure", func() {
			Expect(buildClient().Destroy()).To(Succeed())
			Expect(actions).To(ContainElement("destroying terraform"))
		})

		It("Deletes the config", func() {
			Expect(buildClient().Destroy()).To(Succeed())
			Expect(actions).To(ContainElement("deleting config"))
		})

		It("Prints a destroy success message", func() {
			Expect(buildClient().Destroy()).To(Succeed())
			Eventually(stdout).Should(gbytes.Say("DESTROY SUCCESSFUL"))
		})
	})

	Describe("FetchInfo", func() {
		BeforeEach(func() {
			configClient.HasAssetReturnsOnCall(0, true, nil)
			configClient.LoadAssetReturnsOnCall(0, directorCredsFixture, nil)
		})

		It("Loads the config file", func() {
			_, err := buildClient().FetchInfo()
			Expect(err).NotTo(HaveOccurred())
			Expect(actions).To(ContainElement("loading config file"))
		})

		It("calls TFInputVarsFactory, having populated AllowIPs and SourceAccessIPs", func() {
			Expect(buildClient().Deploy()).To(Succeed())
			Expect(tfInputVarsFactory.NewInputVarsCallCount()).To(Equal(1))
			Expect(tfInputVarsFactory.NewInputVarsArgsForCall(0)).To(Equal(configAfterLoad))
		})

		It("Loads terraform output", func() {
			_, err := buildClient().FetchInfo()
			Expect(err).NotTo(HaveOccurred())
			Expect(actions).To(ContainElement("initializing terraform outputs"))
		})

		It("Checks that the IP is whitelisted", func() {
			_, err := buildClient().FetchInfo()
			Expect(err).NotTo(HaveOccurred())
			Expect(actions).To(ContainElement("checking security group for IP"))
		})

		It("Retrieves the BOSH instances", func() {
			_, err := buildClient().FetchInfo()
			Expect(err).NotTo(HaveOccurred())
			Expect(actions).To(ContainElement("listing bosh instances"))
		})

		Context("When the IP address isn't properly whitelisted", func() {
			BeforeEach(func() {
				ipChecker = func() (string, error) {
					return "1.2.3.4", nil
				}
			})

			It("Returns a meaningful error", func() {
				_, err := buildClient().FetchInfo()
				Expect(err).To(MatchError("Do you need to add your IP 1.2.3.4 to the control-tower-happymeal-director security group/source range entry for director firewall (for ports 22, 6868, and 25555)?"))
			})
		})
	})
})
