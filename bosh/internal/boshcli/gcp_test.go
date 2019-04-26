package boshcli

import (
	"fmt"
	"github.com/EngineerBetter/control-tower/resource"
	"io/ioutil"
	"testing"
	"text/template"
)

func TestGCPEnvironment_ConfigureDirectorCloudConfig(t *testing.T) {

	fullTemplateParams := GCPEnvironment{
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

	getFixture := func(f string) string {
		contents, _ := ioutil.ReadFile(f)
		return string(contents)
	}

	tests := []struct {
		name     string
		fields   GCPEnvironment
		want     string
		wantErr  bool
		init     func(GCPEnvironment) GCPEnvironment
		validate func(string, string) (bool, string)
	}{
		{
			name:    "Success- template rendered",
			fields:  fullTemplateParams,
			want:    getFixture("../fixtures/gcp_cloud_config_full.yml"),
			wantErr: false,
			init: func(e GCPEnvironment) GCPEnvironment {
				return e
			},
			validate: func(a, b string) (bool, string) {
				return a == b, fmt.Sprintf("basic rendering expected to work")
			},
		},
		{
			name:    "Success- spot instance rendered",
			fields:  fullTemplateParams,
			want:    getFixture("../fixtures/gcp_cloud_config_spot.yml"),
			wantErr: false,
			init: func(e GCPEnvironment) GCPEnvironment {
				n := e
				n.Spot = true
				return n
			},
			validate: func(a, b string) (bool, string) {
				return a == b, fmt.Sprintf("templating failed while rendering without spots")
			},
		},

		{
			name:    "Success- running with no spot",
			fields:  fullTemplateParams,
			want:    getFixture("../fixtures/gcp_cloud_config_no_spot.yml"),
			wantErr: false,
			init: func(e GCPEnvironment) GCPEnvironment {
				n := e
				n.Spot = false
				return n
			},
			validate: func(a, b string) (bool, string) {
				return a == b, fmt.Sprintf("templating failed while rendering without spots")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.init(tt.fields)
			got, err := e.ConfigureDirectorCloudConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("Environment.ConfigureDirectorCloudConfig()\nerror expected:  %v\nreceived error:  %v", tt.wantErr, err)
				return
			}
			passed, message := tt.validate(got, tt.want)
			if !passed {
				t.Errorf(message)
			}
		})
	}
}

func TestGCPEnvironment_ConfigureConcourseStemcell(t *testing.T) {
	type args struct {
		versions string
	}
	tests := []struct {
		name    string
		want    string
		wantErr bool
		fixture string
	}{
		{
			name:    "parse versions and provide a valid stemcell url",
			want:    "https://s3.amazonaws.com/bosh-gce-light-stemcells/5/light-bosh-stemcell-5-google-kvm-ubuntu-xenial-go_agent.tgz",
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
