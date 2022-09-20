package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Pubkey struct {
	X string `yaml:"x"`
	Y string `yaml:"y"`
}

type BK struct {
	X    string `yaml:"x"`
	Rank uint32 `yaml:"rank"`
}

func WriteYamlFile(yamlData interface{}, filePath string) error {
	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0600)
}

type SignerConfig struct {
	Port   int64         `yaml:"port"`
	Share  string        `yaml:"share"`
	Pubkey Pubkey        `yaml:"pubkey"`
	BKs    map[string]BK `yaml:"bks"`

	BadgerDir string `yaml:"badger-dir"`
}

type MsgConfig struct {
	Message string `yaml:"msg"`
}
