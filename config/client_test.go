package config_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/EngineerBetter/control-tower/iaas/iaasfakes"

	. "github.com/EngineerBetter/control-tower/config"
	"github.com/EngineerBetter/control-tower/iaas"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var provider *iaasfakes.FakeProvider
	var client *Client

	BeforeEach(func() {
		provider = &iaasfakes.FakeProvider{}
		provider.RegionReturns("eu-west-1")
		provider.DBTypeReturns("db.t3.medium")
		provider.EnsureFileExistsStub = func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
			return defaultContents, true, nil
		}

		client = New(provider, "test", "")
	})

	Describe("NewConfig", func() {
		It("populates fields correctly", func() {
			conf := client.NewConfig()
			Expect(conf.ConfigBucket).To(Equal("control-tower-test-eu-west-1-config"))
			Expect(conf.Deployment).To(Equal("control-tower-test"))
			Expect(conf.Namespace).To(Equal("eu-west-1"))
			Expect(conf.Project).To(Equal("test"))
			Expect(conf.Region).To(Equal("eu-west-1"))
			Expect(conf.TFStatePath).To(Equal("terraform.tfstate"))
		})
	})

	Describe("EnsureBucketExists", func() {
		BeforeEach(func() {
			provider = &iaasfakes.FakeProvider{}
			provider.RegionReturns("eu-west-1")
			client = New(provider, "test", "")
		})

		Context("creating the client caused a BucketError", func() {
			JustBeforeEach(func() {
				client.BucketError = errors.New("Ooops")
			})

			It("returns the error", func() {
				err := client.EnsureBucketExists()
				Expect(err).To(MatchError("client failed to configure properly: [Ooops]"))
			})
		})

		Context("when the bucket exists", func() {
			JustBeforeEach(func() {
				provider.BucketExistsReturns(false, nil)
			})

			It("no-ops and returns no error", func() {
				Expect(client.EnsureBucketExists()).To(Succeed())
			})
		})

		Context("when the bucket does not exist", func() {
			JustBeforeEach(func() {
				provider.BucketExistsReturns(false, nil)
				provider.CreateBucketReturns(nil)
			})

			It("creates it", func() {
				Expect(client.EnsureBucketExists()).To(Succeed())
				// Will be 1 once we remove invocation from New()
				Expect(provider.CreateBucketCallCount()).To(Equal(1))
			})

			Context("and it cannot be created", func() {
				JustBeforeEach(func() {
					provider.BucketExistsReturns(false, nil)
					provider.CreateBucketReturns(fmt.Errorf("SOME IAAS ERROR"))
				})

				It("returns a useful error message", func() {
					err := client.EnsureBucketExists()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error creating config bucket [control-tower-test-eu-west-1-config]: [SOME IAAS ERROR]"))
				})
			})
		})
	})
})

func TestNew(t *testing.T) {
	provider := new(iaasfakes.FakeProvider)
	provider.RegionReturns("eu-west-1")
	provider.DBTypeReturns("db.t3.medium")
	provider.EnsureFileExistsStub = func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
		return defaultContents, true, nil
	}

	type args struct {
		iaas      iaas.Provider
		project   string
		namespace string
	}
	tests := []struct {
		name             string
		args             args
		want             *Client
		FakeBucketExists func(name string) (bool, error)
	}{
		{
			name: "default",
			args: args{
				iaas:      provider,
				project:   "aProject",
				namespace: "",
			},
			want: &Client{
				Iaas:         provider,
				Project:      "aProject",
				Namespace:    "eu-west-1",
				BucketName:   "control-tower-aProject-eu-west-1-config",
				BucketExists: false,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				return false, nil
			},
		},
		{
			name: "with Namespace",
			args: args{
				iaas:      provider,
				project:   "aProject",
				namespace: "someNamespace",
			},
			want: &Client{
				Iaas:         provider,
				Project:      "aProject",
				Namespace:    "someNamespace",
				BucketName:   "control-tower-aProject-someNamespace-config",
				BucketExists: false,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				return false, nil
			},
		},
		{
			name: "with Namespace and region based bucket",
			args: args{
				iaas:      provider,
				project:   "aProject",
				namespace: "someNamespace",
			},
			want: &Client{
				Iaas:         provider,
				Project:      "aProject",
				Namespace:    "someNamespace",
				BucketName:   "control-tower-aProject-eu-west-1-config",
				BucketExists: true,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				if name == "control-tower-aProject-eu-west-1-config" {
					return true, nil
				}
				return false, nil
			},
		},
		{
			name: "with Namespace and namespace based bucket",
			args: args{
				iaas:      provider,
				project:   "aProject",
				namespace: "someNamespace",
			},
			want: &Client{
				Iaas:         provider,
				Project:      "aProject",
				Namespace:    "someNamespace",
				BucketName:   "control-tower-aProject-someNamespace-config",
				BucketExists: true,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				if name == "control-tower-aProject-someNamespace-config" {
					return true, nil
				}
				return false, fmt.Errorf("an error")
			},
		},
		{
			name: "with Namespace and bucket existing and namespace == region",
			args: args{
				iaas:      provider,
				project:   "aProject",
				namespace: "eu-west-1",
			},
			want: &Client{
				Iaas:         provider,
				Project:      "aProject",
				Namespace:    "eu-west-1",
				BucketName:   "control-tower-aProject-eu-west-1-config",
				BucketExists: true,
				BucketError:  nil,
			},
			FakeBucketExists: func(name string) (bool, error) {
				return true, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider.BucketExistsStub = tt.FakeBucketExists
			if got := New(tt.args.iaas, tt.args.project, tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v,\n want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Load(t *testing.T) {
	provider := new(iaasfakes.FakeProvider)
	provider.RegionReturns("eu-west-1")
	provider.DBTypeReturns("db.t3.medium")
	provider.EnsureFileExistsStub = func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
		return defaultContents, true, nil
	}
	provider.LoadFileStub = func(bucket, path string) ([]byte, error) {
		bytes, _ := json.Marshal(Config{Spot: true})
		return bytes, nil
	}

	tests := []struct {
		name    string
		prepare func() *Client
		want    Config
		wantErr bool
	}{
		{
			name: "BucketError is raised",
			prepare: func() *Client {
				return &Client{
					Iaas:         provider,
					Project:      "",
					Namespace:    "",
					BucketName:   "",
					BucketExists: false,
					BucketError:  fmt.Errorf("some error"),
				}
			},
			want:    Config{},
			wantErr: true,
		},
		{
			name: "default",
			prepare: func() *Client {
				return &Client{
					Iaas:         provider,
					Project:      "",
					Namespace:    "",
					BucketName:   "",
					BucketExists: false,
					BucketError:  nil,
				}
			},
			want: Config{
				Spot:               true,
				VMProvisioningType: SPOT,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.prepare()
			got, err := client.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_HasConfig(t *testing.T) {
	provider := new(iaasfakes.FakeProvider)
	provider.RegionReturns("eu-west-1")
	provider.DBTypeReturns("db.t3.medium")
	provider.EnsureFileExistsStub = func(bucket, path string, defaultContents []byte) ([]byte, bool, error) {
		return defaultContents, true, nil
	}
	provider.LoadFileStub = func(bucket, path string) ([]byte, error) {
		bytes, _ := json.Marshal(Config{})
		return bytes, nil
	}

	tests := []struct {
		name    string
		prepare func() *Client
		want    Config
		wantErr bool
	}{
		{
			name: "BucketError is raised",
			prepare: func() *Client {
				return &Client{
					Iaas:         provider,
					Project:      "",
					Namespace:    "",
					BucketName:   "",
					BucketExists: false,
					BucketError:  fmt.Errorf("some error"),
				}
			},
			want:    Config{},
			wantErr: true,
		},
		{
			name: "default",
			prepare: func() *Client {
				return &Client{
					Iaas:         provider,
					Project:      "",
					Namespace:    "",
					BucketName:   "",
					BucketExists: false,
					BucketError:  nil,
				}
			},
			want: Config{
				VMProvisioningType: ON_DEMAND,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.prepare()
			got, err := client.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Load() = %v, want %v", got, tt.want)
			}
		})
	}
}
