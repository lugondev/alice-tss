package dkg

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"alice-tss/config"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/sirius/log"
)

type DKGConfig struct {
	Port      int64   `yaml:"port"`
	Rank      uint32  `yaml:"rank"`
	Threshold uint32  `yaml:"threshold"`
	Peers     []int64 `yaml:"peers"`
}

type DKGResult struct {
	Share  string               `yaml:"share"`
	Pubkey config.Pubkey        `yaml:"pubkey"`
	BKs    map[string]config.BK `yaml:"bks"`
}

func readDKGConfigFile(filaPath string) (*DKGConfig, error) {
	c := &DKGConfig{}
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

func writeDKGResult(id string, result *dkg.Result) error {
	dkgResult := &DKGResult{
		Share: result.Share.String(),
		Pubkey: config.Pubkey{
			X: result.PublicKey.GetX().String(),
			Y: result.PublicKey.GetY().String(),
		},
		BKs: make(map[string]config.BK),
	}
	for peerID, bk := range result.Bks {
		dkgResult.BKs[peerID] = config.BK{
			X:    bk.GetX().String(),
			Rank: bk.GetRank(),
		}
	}
	err := config.WriteYamlFile(dkgResult, getFilePath(id))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func getFilePath(id string) string {
	return fmt.Sprintf("dkg/%s-output.yaml", id)
}
