package types

import "github.com/ethereum/go-ethereum/common"

type Pubkey struct {
	X string
	Y string
}

type BK struct {
	X    string
	Rank uint32
}

type SignerConfig struct {
	Share  string        `json:"share"`
	Pubkey Pubkey        `json:"pubkey"`
	BKs    map[string]BK `json:"bks"`
}

type DKGConfig struct {
	Rank      uint32 `json:"rank"`
	Threshold uint32 `json:"threshold"`
}

type DKGResult struct {
	Share     string         `json:"share"`
	Pubkey    Pubkey         `json:"pubkey"`
	PublicKey string         `json:"publicKey"`
	Address   common.Address `json:"address"`
	BKs       map[string]BK  `json:"bks"`
}

type ReshareConfig struct {
	Threshold uint32        `json:"threshold"`
	Share     string        `json:"share"`
	Pubkey    Pubkey        `json:"pubkey"`
	BKs       map[string]BK `json:"bks"`
}

type ReshareResult struct {
	Share string `json:"share"`
}

type RVSignature struct {
	R    string `json:"r"`
	S    string `json:"s"`
	Hash string `json:"hash"`
}
