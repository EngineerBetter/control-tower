package certsfakes

import (
	"github.com/EngineerBetter/control-tower/certs"
	"github.com/xenolf/lego/lego"
)

// Not really a fake, but it seemed best to put it here
func NewFakeAcmeClient(u *certs.User) (*lego.Client, error) {
	return &lego.Client{}, nil
}
