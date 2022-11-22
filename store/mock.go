package store

import (
	"alice-tss/types"
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/reshare"
	"github.com/getamis/sirius/log"
	"math/big"
)

type MockDB struct {
}

var (
	dkgResults    map[string]*types.DKGResult
	signerConfigs map[string]*types.SignerConfig
)

func (d *MockDB) SaveDKGResultData(hash string, result *dkg.Result) error {
	log.Info("SaveDKGResultData", "hash", hash, "result", result)

	pubkey := crypto.CompressPubkey(result.PublicKey.ToPubKey())
	dkgResults[hash] = &types.DKGResult{
		Address:   crypto.PubkeyToAddress(*result.PublicKey.ToPubKey()),
		Share:     common.Bytes2Hex(result.Share.Bytes()),
		PublicKey: hex.EncodeToString(pubkey),
		BKs:       map[string]types.BK{},
		Pubkey: types.Pubkey{
			X: hex.EncodeToString(result.PublicKey.GetX().Bytes()),
			Y: hex.EncodeToString(result.PublicKey.GetY().Bytes()),
		},
	}
	for s, parameter := range result.Bks {
		dkgResults[hash].BKs[s] = types.BK{
			X:    parameter.GetX().String(),
			Rank: parameter.GetRank(),
		}
	}

	signerCfg := &types.SignerConfig{
		Share: big.NewInt(0).SetBytes(result.Share.Bytes()).String(),
		Pubkey: types.Pubkey{
			X: big.NewInt(0).SetBytes(result.PublicKey.GetX().Bytes()).String(),
			Y: big.NewInt(0).SetBytes(result.PublicKey.GetY().Bytes()).String(),
		},
		BKs: map[string]types.BK{},
	}

	for s, parameter := range result.Bks {
		signerCfg.BKs[s] = types.BK{
			X:    parameter.GetX().String(),
			Rank: parameter.GetRank(),
		}
	}

	signerConfigs[hash] = signerCfg

	return nil
}

func (d *MockDB) GetSignerConfig(hash, pubkey string) (*types.SignerConfig, error) {
	log.Info("GetSignerConfig", "hash", hash, "pubkey", pubkey)

	signerConfig := signerConfigs[hash]
	if signerConfig == nil {
		return nil, errors.New("signer config not found")
	}

	return signerConfig, nil
}

func (d *MockDB) UpdateDKGResultData(hash string, result *reshare.Result) error {
	log.Info("UpdateDKGResultData", "hash", hash, "result", result)
	return nil
}

func (d *MockDB) SaveSignerResultData(hash string, result types.RVSignature) error {
	log.Info("SaveSignerResultData", "hash", hash, "result", result)
	return nil
}

func (d *MockDB) GetDKGResultData(hash string) (*types.DKGResult, error) {
	log.Info("GetDKGResultData", "hash", hash)
	return dkgResults[hash], nil
}

func (d *MockDB) Defer() {
}

// NewMockDB implementation using mock
func NewMockDB() types.StoreDB {
	dkgResults = map[string]*types.DKGResult{}
	signerConfigs = make(map[string]*types.SignerConfig)

	return &MockDB{}
}
