package commands_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Control-Tower Commands Suite")
}

func controlTowerCommand(args ...string) *exec.Cmd {
	return exec.Command("go", append([]string{"run", "github.com/EngineerBetter/control-tower"}, args...)...)
}
