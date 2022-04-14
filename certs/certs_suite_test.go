package certs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCerts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Certs Suite")
}
