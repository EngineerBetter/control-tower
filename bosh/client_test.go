package bosh_test

import (
	"bytes"
	_ "embed"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/EngineerBetter/control-tower/bosh"
	"github.com/EngineerBetter/control-tower/bosh/internal/boshcli/boshclifakes"
	"github.com/EngineerBetter/control-tower/bosh/internal/workingdir/workingdirfakes"
	"github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/iaas/iaasfakes"
	"github.com/EngineerBetter/control-tower/terraform/terraformfakes"
)

//go:embed fixtures/private_key.pem
var privateKey string

var _ = Describe("Client", func() {
	var (
		buildClient      func() bosh.IClient
		boshCLI          *boshclifakes.FakeICLI
		directorClient   *workingdirfakes.FakeIClient
		configInput      config.Config
		terraformOutputs *terraformfakes.FakeOutputs
		provider         *iaasfakes.FakeProvider
		versionFile      []byte
	)

	BeforeEach(func() {
		boshCLI = new(boshclifakes.FakeICLI)
		directorClient = new(workingdirfakes.FakeIClient)
		terraformOutputs = new(terraformfakes.FakeOutputs)
		provider = new(iaasfakes.FakeProvider)

		configInput = config.Config{
			PrivateKey: privateKey,
			PublicKey:  "example-public-key",
		}
	})

	Describe("New", func() {
		When("provider is AWS", func() {
			BeforeEach(func() {
				provider = buildFakeAwsProvider()
			})

			When("Bosh CLI url is in versionFile", func() {
				BeforeEach(func() {
					versionFile = []byte(`{
						"bosh-cli": {
							"mac": "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-darwin-amd64",
							"linux": "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-linux-amd64"
						}
					}`)
				})

				It("returns an AWSClient", func() {
					client, err := bosh.New(configInput, terraformOutputs, io.Discard, io.Discard, provider, versionFile)
					Expect(err).ToNot(HaveOccurred())
					Expect(client).To(BeAssignableToTypeOf(&bosh.AWSClient{}))
				})
			})

			When("Bosh CLI url is not in versionFile", func() {
				BeforeEach(func() {
					versionFile = []byte(`{}`)
				})

				It("returns an appropriate error", func() {
					_, err := bosh.New(configInput, terraformOutputs, io.Discard, io.Discard, provider, versionFile)
					Expect(err.Error()).To(HavePrefix("failed to determine BOSH CLI path:"))
				})
			})
		})

		When("provider is GCP", func() {
			BeforeEach(func() {
				provider = buildFakeGCPProvider()
			})

			When("Bosh CLI url is in versionFile", func() {
				BeforeEach(func() {
					versionFile = []byte(`{
						"bosh-cli": {
							"mac": "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-darwin-amd64",
							"linux": "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-linux-amd64"
						}
					}`)
				})

				It("returns an AWSClient", func() {
					client, err := bosh.New(configInput, terraformOutputs, io.Discard, io.Discard, provider, versionFile)
					Expect(err).NotTo(HaveOccurred())
					Expect(client).To(BeAssignableToTypeOf(&bosh.GCPClient{}))
				})
			})

			When("Bosh CLI url is not in versionFile", func() {
				BeforeEach(func() {
					versionFile = []byte(`{}`)
				})

				It("returns an appropriate error", func() {
					_, err := bosh.New(configInput, terraformOutputs, io.Discard, io.Discard, provider, versionFile)
					Expect(err.Error()).To(HavePrefix("failed to determine BOSH CLI path:"))
				})
			})
		})

		When("provider is unknown", func() {
			BeforeEach(func() {
				versionFile = []byte(`{
					"bosh-cli": {
						"mac": "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-darwin-amd64",
						"linux": "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.0.1-linux-amd64"
					}
				}`)
			})

			It("returns an appropriate error", func() {
				_, err := bosh.New(configInput, terraformOutputs, io.Discard, io.Discard, provider, versionFile)
				Expect(err.Error()).To(HavePrefix("IAAS not supported: Unknown"))
			})
		})
	})

	Describe("Instances", func() {
		When("on AWS", func() {
			var stdout *bytes.Buffer

			BeforeEach(func() {
				provider = buildFakeAwsProvider()
				versionFile = []byte("{}")
				stdout = new(bytes.Buffer)

				buildClient = func() bosh.IClient {
					client, err := bosh.NewAWSClient(configInput, terraformOutputs, directorClient, stdout, io.Discard, provider, boshCLI, versionFile)
					Expect(err).NotTo(HaveOccurred())
					return client
				}
			})

			When("instances are found", func() {
				BeforeEach(func() {
					boshCLI.RunAuthenticatedCommandStub = func(action, ip, password, ca string, detach bool, stdout io.Writer, flags ...string) error {
						io.WriteString(stdout, `{"Tables":[{"Rows": [{"instance": "foo","ips": "1.2.3.4", "process_state": "bar"}]}]}`)
						return nil
					}
				})

				It("returns them", func() {
					expectedInstance := bosh.Instance{
						Name:  "foo",
						IP:    "1.2.3.4",
						State: "bar",
					}

					instances, err := buildClient().Instances()
					Expect(err).NotTo(HaveOccurred())
					Expect(instances).To(Equal([]bosh.Instance{expectedInstance}))
				})
			})
		})
	})
})

func buildFakeGCPProvider() *iaasfakes.FakeProvider {
	provider := &iaasfakes.FakeProvider{
		DBTypeStub: func(size string) string {
			return "db.t3." + size
		},
		CheckForWhitelistedIPStub: func(ip, securityGroup string) (bool, error) {
			if ip == "1.2.3.4" {
				return false, nil
			}
			return true, nil
		},
		FindLongestMatchingHostedZoneStub: func(subdomain string) (string, string, error) {
			if subdomain == "ci.google.com" {
				return "google.com", "ABC123", nil
			}

			return "", "", errors.New("hosted zone not found")
		},
	}

	provider.RegionReturns("eu-west-1")
	provider.IAASReturns(iaas.GCP)

	return provider
}

func buildFakeAwsProvider() *iaasfakes.FakeProvider {
	provider := &iaasfakes.FakeProvider{
		DBTypeStub: func(size string) string {
			switch size {
			case "small":
				return "db-g1-small"
			case "medium":
				return "db-custom-2-4096"
			default:
				return "big-db-is-big"
			}
		},
		CheckForWhitelistedIPStub: func(ip, securityGroup string) (bool, error) {
			if ip == "1.2.3.4" {
				return false, nil
			}
			return true, nil
		},
		FindLongestMatchingHostedZoneStub: func(subdomain string) (string, string, error) {
			if subdomain == "ci.google.com" {
				return "google.com", "ABC123", nil
			}

			return "", "", errors.New("hosted zone not found")
		},
	}

	provider.RegionReturns("eu-west-1")
	provider.IAASReturns(iaas.AWS)

	return provider
}
