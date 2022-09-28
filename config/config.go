package config

import "github.com/ethereum/go-ethereum/common"

type Pubkey struct {
	X string `yaml:"x"`
	Y string `yaml:"y"`
}

type RVSignature struct {
	R    string `json:"r"`
	S    string `json:"s"`
	Hash string `json:"hash"`
}

type BK struct {
	X    string `yaml:"x"`
	Rank uint32 `yaml:"rank"`
}

type SignerConfig struct {
	Port   int64         `yaml:"port"`
	Share  string        `yaml:"share" json:"share"`
	Pubkey Pubkey        `yaml:"pubkey" json:"pubkey"`
	BKs    map[string]BK `yaml:"bks" json:"bks"`

	BadgerDir string `yaml:"badger-dir"`
}

type DKGConfig struct {
	Rank      uint32 `yaml:"rank" json:"rank"`
	Threshold uint32 `yaml:"threshold" json:"threshold"`
}

type DKGResult struct {
	Share     string         `yaml:"share" json:"share"`
	Pubkey    Pubkey         `yaml:"pubkey" json:"pubkey"`
	PublicKey string         `yaml:"publicKey" json:"publicKey"`
	Address   common.Address `json:"address"`
	BKs       map[string]BK  `yaml:"bks" json:"bks"`
}

type ReshareConfig struct {
	Threshold uint32        `yaml:"threshold" json:"threshold"`
	Share     string        `yaml:"share" json:"share"`
	Pubkey    Pubkey        `yaml:"pubkey" json:"pubkey"`
	BKs       map[string]BK `yaml:"bks" json:"bks"`
}

type ReshareResult struct {
	Share string `yaml:"share" json:"share"`
}
