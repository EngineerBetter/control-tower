//go:build integration
// +build integration

package certs_test

import (
	. "github.com/EngineerBetter/control-tower/certs"
	"github.com/EngineerBetter/control-tower/iaas/iaasfakes"
	"github.com/EngineerBetter/control-tower/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Certs", func() {
	var constructor = certsfakes.NewFakeAcmeClient
	var provider = &iaasfakes.FakeProvider{}

	It("Generates a cert for an IP address", func() {
		certs, err := Generate(constructor, "control-tower-mole", &provider, "99.99.99.99")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(certs.CACert)).To(ContainSubstring("BEGIN CERTIFICATE"))
		Expect(string(certs.Key)).To(ContainSubstring("BEGIN RSA PRIVATE KEY"))
		Expect(string(certs.Cert)).To(ContainSubstring("BEGIN CERTIFICATE"))
	})

	It("Generates a cert for a domain", func() {
		certs, err := Generate(constructor, "control-tower-mole", &provider, "control-tower-test-"+util.GeneratePasswordWithLength(10)+".engineerbetter.com")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(certs.CACert)).To(ContainSubstring("BEGIN CERTIFICATE"))
		Expect(string(certs.Key)).To(ContainSubstring("BEGIN RSA PRIVATE KEY"))
		Expect(string(certs.Cert)).To(ContainSubstring("BEGIN CERTIFICATE"))
	})

	It("Can't generate a cert for google.com", func() {
		_, err := Generate(constructor, "control-tower-mole", &provider, "google.com")
		Expect(err).To(HaveOccurred())
	})
})
