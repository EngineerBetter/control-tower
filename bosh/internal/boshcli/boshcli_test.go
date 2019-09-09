package boshcli_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/EngineerBetter/control-tower/bosh/internal/boshcli"
	"github.com/EngineerBetter/control-tower/internal/fakeexec"
	"github.com/EngineerBetter/control-tower/util"
	"github.com/stretchr/testify/require"
)

func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Print(os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(i)
}

type mockIAASConfig struct {
}

func (c mockIAASConfig) ConfigureDirectorManifestCPI() (string, error) {
	return "a CPI", nil
}

func (c mockIAASConfig) ConfigureDirectorCloudConfig() (string, error) {
	return "a Cloud Config", nil
}

func (c mockIAASConfig) ConcourseStemcellURL() (string, error) {
	return "a Stemcell", nil
}

func (c mockIAASConfig) ExtractBOSHandBPM() (util.Resource, util.Resource, error) {
	return util.Resource{}, util.Resource{}, nil
}

func TestCLI_CreateEnv(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	c := boshcli.New("bosh", e.Cmd())
	config := mockIAASConfig{}
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "bosh", command)
		require.Equal(t, "create-env", args[0])
	})
	c.CreateEnv(&boshcli.CreateEnvFiles{}, config, "password", "cert", "key", "ca", map[string]string{})
}

func TestCLI_UpdateCloudConfig(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	c := boshcli.New("bosh", e.Cmd())
	config := mockIAASConfig{}
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "bosh", command)

		require.Equal(t, "--non-interactive", args[0])
		require.Equal(t, "--environment", args[1])
		require.Equal(t, "https://ip", args[2])
		require.Equal(t, "--client-secret", args[7])
		require.Equal(t, "password", args[8])
		require.Equal(t, "update-cloud-config", args[9])
	})
	err := c.UpdateCloudConfig(config, "ip", "password", "ca")
	require.NoError(t, err)
}

func TestCLI_UploadConcourseStemcell(t *testing.T) {
	e := fakeexec.New(t)
	defer e.Finish()
	c := boshcli.New("bosh", e.Cmd())
	config := mockIAASConfig{}
	e.ExpectFunc(func(t testing.TB, command string, args ...string) {
		require.Equal(t, "bosh", command)

		require.Equal(t, "--non-interactive", args[0])
		require.Equal(t, "--environment", args[1])
		require.Equal(t, "https://ip", args[2])
		require.Equal(t, "--client-secret", args[7])
		require.Equal(t, "password", args[8])
		require.Equal(t, "upload-stemcell", args[9])
	})
	err := c.UploadConcourseStemcell(config, "ip", "password", "ca")
	require.NoError(t, err)

}
