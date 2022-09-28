package cmd

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"alice-tss/config"
)

func readAppConfigFile(filaPath string) (*config.AppConfig, error) {
	c := &config.AppConfig{}
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
