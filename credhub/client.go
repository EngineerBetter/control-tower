package credhub

import (
	"fmt"

	ch "code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"code.cloudfoundry.org/credhub-cli/credhub/credentials/values"
	"github.com/EngineerBetter/control-tower/iaas"
	"github.com/EngineerBetter/control-tower/terraform"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . IClient
type IClient interface {
	SetSelfUpdateCreds(provider iaas.Provider, tfOutputs terraform.Outputs) error
}

type Client struct {
	credHub *ch.CredHub
}

// Creates a new CredHub client using provided details
func NewClient(server, id, secret, cert string) (IClient, error) {
	client, err := ch.New(
		server,
		ch.CaCerts(cert),
		ch.Auth(auth.UaaClientCredentials(id, secret)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create credhub client: [%q]", err)
	}
	return &Client{
		client,
	}, nil
}

func (client *Client) SetSelfUpdateCreds(provider iaas.Provider, tfOutputs terraform.Outputs) error {
	switch provider.IAAS() {
	case iaas.AWS:
		keyId, err := tfOutputs.Get("SelfUpdateUserAccessKeyID")
		if err != nil {
			return err
		}
		secretKey, err := tfOutputs.Get("SelfUpdateSecretAccessKey")
		if err != nil {
			return err
		}
		_, err = client.credHub.SetValue("/concourse/main/control-tower-self-update/aws_access_key_id", values.Value(keyId))
		if err != nil {
			return err
		}
		_, err = client.credHub.SetValue("/concourse/main/control-tower-self-update/aws_secret_access_key", values.Value(secretKey))
		if err != nil {
			return err
		}
	case iaas.GCP:
		googleCreds, err := tfOutputs.Get("SelfUpdateAccountCreds")
		if err != nil {
			return err
		}
		_, err = client.credHub.SetValue("/concourse/main/control-tower-self-update/google_self_update_credentials", values.Value(googleCreds))
		if err != nil {
			return err
		}
	}
	return nil
}
