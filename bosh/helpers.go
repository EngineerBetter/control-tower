package bosh

import (
	"fmt"
	"github.com/apparentlymart/go-cidr/cidr"
	"net"
	"strings"
)

func vars(vars map[string]interface{}) []string {
	var x []string
	for k, v := range vars {
		switch v.(type) {
		case string:
			if k == "tags" {
				x = append(x, "--var", fmt.Sprintf("%s=%s", k, v))
				continue
			}
			x = append(x, "--var", fmt.Sprintf("%s=%q", k, v))
		case int:
			x = append(x, "--var", fmt.Sprintf("%s=%d", k, v))
		case bool:
			x = append(x, "--var", fmt.Sprintf("%s=%t", k, v))
		default:
			panic("unsupported type")
		}
	}
	return x
}

func splitTags(ts []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, t := range ts {
		ss := strings.SplitN(t, "=", 2)
		if len(ss) != 2 {
			return nil, fmt.Errorf("could not split tag %q", t)
		}
		m[ss[0]] = ss[1]
	}
	return m, nil
}

func formatIPRange(forCIDR, sep string, positions []int) (string, error) {
	var ips []string
	_, parsedCIDR, err := net.ParseCIDR(forCIDR)
	if err != nil {
		return "", err
	}
	for _, pos := range positions {
		ip, _ := cidr.Host(parsedCIDR, pos)
		ips = append(ips, ip.String())

	}
	s := fmt.Sprintf(`[%s]`, strings.Join(ips, sep))
	return s, nil
}
