package fly_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFly(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fly Suite")
}
