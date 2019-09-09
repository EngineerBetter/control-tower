package util

import (
	"encoding/json"
	"runtime"

	"github.com/EngineerBetter/control-tower/util/bincache"
)

// Resource safely exposes the json parameters of a resource
type Resource struct {
	URL     string `json:"url"`
	Version string `json:"version"`
	SHA1    string `json:"sha1"`
}

type BinaryPaths struct {
	Mac   string `json:"mac"`
	Linux string `json:"linux"`
}

func ParseVersionResources(versionFile []byte) map[string]Resource {
	var resources map[string]Resource

	err := json.Unmarshal(versionFile, &resources)
	if err != nil {
		panic(err)
	}

	return resources
}

func GetResource(name string, resources map[string]Resource) Resource {
	r, ok := resources[name]
	if !ok {
		panic("resource " + name + " not found")
	}
	return r
}

func (p BinaryPaths) path() string {
	switch runtime.GOOS {
	case "darwin":
		return p.Mac
	case "linux":
		return p.Linux
	default:
		panic("OS not supported")
	}
}

// DownloadBOSHCLI returns the path of the downloaded bosh-cli
func DownloadBOSHCLI(binaries map[string]BinaryPaths) (string, error) {
	p := binaries["bosh-cli"].path()
	return bincache.Download(p)
}

// DownloadTerraformCLI returns the path of the downloaded terraform-cli
func DownloadTerraformCLI(binaries map[string]BinaryPaths) (string, error) {
	p := binaries["terraform"].path()
	return bincache.Download(p)
}
