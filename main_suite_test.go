package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestControlTower(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Control-Tower Suite")
}
