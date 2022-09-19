package signer

import (
	"fmt"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/signer"
	"github.com/getamis/sirius/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"alice-tss/config"
	"alice-tss/peer"
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

func writeSignerResult(id string, result *signer.Result) error {
	signerResult := &peer.SignerResult{
		R: result.R.String(),
		S: result.S.String(),
	}
	err := config.WriteYamlFile(signerResult, getFilePath(id))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func getFilePath(id string) string {
	return fmt.Sprintf("signer/%s-output.yaml", id)
}
