package config

import (
	"encoding/json"
	"fmt"

	"github.com/EngineerBetter/control-tower/iaas"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const terraformStateFileName = "terraform.tfstate"
const configFilePath = "config.json"

//counterfeiter: generate . IClient
type IClient interface {
	Load() (Config, error)
	DeleteAll(config ConfigView) error
	Update(Config) error
	StoreAsset(filename string, contents []byte) error
	HasAsset(filename string) (bool, error)
	ConfigExists() (bool, error)
	LoadAsset(filename string) ([]byte, error)
	NewConfig() Config
	EnsureBucketExists() error
}

// Client is a client for loading the config file  from S3
type Client struct {
	Iaas         iaas.Provider
	Project      string
	Namespace    string
	BucketName   string
	BucketExists bool
	BucketError  error
}

// New instantiates a new client
func New(iaas iaas.Provider, project, namespace string) *Client {
	namespace = determineNamespace(namespace, iaas.Region())
	bucketName, exists, err := determineBucketName(iaas, namespace, project)

	return &Client{
		iaas,
		project,
		namespace,
		bucketName,
		exists,
		err,
	}
}

// StoreAsset stores an associated configuration file
func (client *Client) StoreAsset(filename string, contents []byte) error {
	return client.Iaas.WriteFile(client.configBucket(),
		filename,
		contents,
	)
}

// LoadAsset loads an associated configuration file
func (client *Client) LoadAsset(filename string) ([]byte, error) {
	return client.Iaas.LoadFile(
		client.configBucket(),
		filename,
	)
}

// HasAsset returns true if an associated configuration file exists
func (client *Client) HasAsset(filename string) (bool, error) {
	return client.Iaas.HasFile(
		client.configBucket(),
		filename,
	)
}

// ConfigExists returns true if the configuration file exists
func (client *Client) ConfigExists() (bool, error) {
	return client.HasAsset(configFilePath)
}

// Update stores the control-tower config file to S3
func (client *Client) Update(config Config) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return client.Iaas.WriteFile(client.configBucket(), configFilePath, bytes)
}

// DeleteAll deletes the entire configuration bucket
func (client *Client) DeleteAll(config ConfigView) error {
	return client.Iaas.DeleteVersionedBucket(config.GetConfigBucket())
}

// Load loads an existing config file from S3
func (client *Client) Load() (Config, error) {
	if client.BucketError != nil {
		return Config{}, client.BucketError
	}

	configBytes, err := client.Iaas.LoadFile(
		client.configBucket(),
		configFilePath,
	)
	if err != nil {
		return Config{}, err
	}

	conf := Config{}
	if err := json.Unmarshal(configBytes, &conf); err != nil {
		return Config{}, err
	}

	conf = populateMandatoryFieldsAddedSinceLastSave(conf)

	return conf, nil
}

func (client *Client) NewConfig() Config {
	return Config{
		ConfigBucket: client.configBucket(),
		Deployment:   deployment(client.Project),
		Namespace:    client.Namespace,
		Project:      client.Project,
		Region:       client.Iaas.Region(),
		TFStatePath:  terraformStateFileName,
	}
}

func (client *Client) EnsureBucketExists() error {
	if client.BucketError != nil {
		return fmt.Errorf("client failed to configure properly: [%v]", client.BucketError)
	}

	exists, err := client.Iaas.BucketExists(client.BucketName)

	if err != nil {
		return fmt.Errorf("error determining if bucket [%v] exists: [%v]", client.BucketName, err)
	}

	if !exists {
		err = client.Iaas.CreateBucket(client.BucketName)

		if err != nil {
			return fmt.Errorf("error creating config bucket [%v]: [%v]", client.BucketName, err)
		}
	}

	return nil
}

func (client *Client) configBucket() string {
	return client.BucketName
}

func deployment(project string) string {
	return fmt.Sprintf("control-tower-%s", project)
}

func createBucketName(deployment, extension string) string {
	return fmt.Sprintf("%s-%s-config", deployment, extension)
}

func determineBucketName(iaas iaas.Provider, namespace, project string) (string, bool, error) {
	regionBucketName := createBucketName(deployment(project), iaas.Region())
	namespaceBucketName := createBucketName(deployment(project), namespace)

	foundRegionNamedBucket, err := iaas.BucketExists(regionBucketName)
	var foundNamespacedBucket bool
	if err != nil {
		foundNamespacedBucket, err = iaas.BucketExists(namespaceBucketName)
		if err != nil {
			return "", false, fmt.Errorf("error looking for possible config buckets [%v] or [%v]: [%v]", regionBucketName, namespaceBucketName, err)
		}
	}

	foundOne := foundRegionNamedBucket || foundNamespacedBucket

	switch {
	case !foundRegionNamedBucket && foundNamespacedBucket:
		return namespaceBucketName, foundOne, nil
	case foundRegionNamedBucket && !foundNamespacedBucket:
		return regionBucketName, foundOne, nil
	default:
		return namespaceBucketName, foundOne, nil
	}
}

func determineNamespace(namespace, region string) string {
	if namespace == "" {
		return region
	}
	return namespace
}

// Allow new mandatory fields to be populated based on old fields that are now deprecated
func populateMandatoryFieldsAddedSinceLastSave(oldConf Config) Config {
	if oldConf.VMProvisioningType == "" {
		oldConf.VMProvisioningType = ConvertSpotBoolToVMProvisioningType(oldConf.Spot)
	}

	return oldConf
}
