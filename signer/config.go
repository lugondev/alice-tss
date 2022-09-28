package signer

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"alice-tss/config"
)

func readSignerConfigFile(filaPath string) (*config.SignerConfig, error) {
	c := &config.SignerConfig{}
	yamlFile, err := ioutil.ReadFile(filaPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
