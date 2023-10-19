package boshcli

import (
	"io/ioutil"
	"testing"
	"text/template"

	"github.com/EngineerBetter/control-tower/resource"
	"github.com/onsi/gomega/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func getFixture(filename string) string {
	contents, _ := ioutil.ReadFile(filename)
	return string(contents)
}

var _ = Describe("GCPEnvironment", func() {
	Describe("ConfigureDirectorCloudConfig", func() {
		var expected string
		var environment GCPEnvironment

		BeforeEach(func() {
			environment = GCPEnvironment{
				Zone:                "zone",
				PublicSubnetwork:    "public_subnetwork",
				PrivateSubnetwork:   "private_subnetwork",
				Spot:                false,
				Network:             "network",
				PublicCIDR:          "public_cidr",
				PublicCIDRGateway:   "public_cidr_gateway",
				PublicCIDRStatic:    "public_cidr_static",
				PublicCIDRReserved:  "public_cidr_reserved",
				PrivateCIDR:         "private_cidr",
				PrivateCIDRGateway:  "private_cidr_gateway",
				PrivateCIDRReserved: "private_cidr_reserved",
			}

			format.TruncatedDiff = false
		})

		Context("when spot instances are not requested", func() {
			BeforeEach(func() {
				expected = getFixture("../fixtures/gcp_cloud_config_no_spot.yml")
			})

			It("renders the expected YAML", func() {
				actual, err := environment.ConfigureDirectorCloudConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})
		})

		Context("when spot instances are requested", func() {
			BeforeEach(func() {
				expected = getFixture("../fixtures/gcp_cloud_config_spot.yml")
				environment.Spot = true
			})

			It("renders the expected YAML", func() {
				actual, err := environment.ConfigureDirectorCloudConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(Equal(expected))
			})
		})
	})
})

func TestGCPEnvironment_ConfigureConcourseStemcell(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
		fixture string
	}{
		{
			name:    "parse versions and provide a valid stemcell url",
			want:    "https://storage.googleapis.com/bosh-gce-light-stemcells/1.260/light-bosh-stemcell-1.260-google-kvm-ubuntu-jammy-go_agent.tgz",
			wantErr: false,
			fixture: "stemcell_version",
		},
		{
			name:    "parse versions and indicate no stemcell was found",
			want:    "",
			wantErr: true,
			fixture: "invalid_stemcell_version",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := GCPEnvironment{}
			resource.GCPReleaseVersions = getStemcellFixture(tt.fixture)
			got, err := e.ConcourseStemcellURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("Environment.ConcourseStemcellURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Environment.ConcourseStemcellURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GCPCloudConfigStructureTest(t *testing.T) {
	t.Run("validating structure", func(t *testing.T) {
		templ, err := template.New("template").Option("missingkey=error").Parse(resource.GCPDirectorCloudConfig)
		if err != nil {
			t.Errorf("cannot parse the template")
		}
		emptyGcpCloudConfigParams := gcpCloudConfigParams{}
		for k, v := range matchStructFields(emptyGcpCloudConfigParams, listTemplFields(templ)) {
			if v < 2 {
				t.Errorf("Field with key name %s is not mapped properly", k)
			}
		}
	})
}
