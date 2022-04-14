package bosh_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBosh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bosh Suite")
}
