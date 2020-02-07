package util

import "gopkg.in/yaml.v2"

func GetDirectorCACertFromCreds(directorCreds []byte) (string, error) {
	var directorCredsStruct struct {
		DirectorCA struct {
			Cert string `yaml:"certificate"`
		} `yaml:"default_ca"`
	}

	err := yaml.Unmarshal(directorCreds, &directorCredsStruct)
	if err != nil {
		return "", err
	}

	return directorCredsStruct.DirectorCA.Cert, nil
}
