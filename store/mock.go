package store

import (
	"alice-tss/types"
	"encoding/hex"
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
	dkgResult *types.DKGResult
)

func (d *MockDB) SaveDKGResultData(hash string, result *dkg.Result) error {
	log.Info("SaveDKGResultData", "hash", hash, "result", result)

	pubkey := crypto.CompressPubkey(result.PublicKey.ToPubKey())
	dkgResult = &types.DKGResult{
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
		dkgResult.BKs[s] = types.BK{
			X:    parameter.GetX().String(),
			Rank: parameter.GetRank(),
		}
	}

	return nil
}

func (d *MockDB) GetSignerConfig(hash, pubkey string) (*types.SignerConfig, error) {
	log.Info("GetSignerConfig", "hash", hash, "pubkey", pubkey)

	signerConfig := &types.SignerConfig{
		Share: big.NewInt(0).SetBytes(common.FromHex(dkgResult.Share)).String(),
		Pubkey: types.Pubkey{
			X: big.NewInt(0).SetBytes(common.FromHex(dkgResult.Pubkey.X)).String(),
			Y: big.NewInt(0).SetBytes(common.FromHex(dkgResult.Pubkey.Y)).String(),
		},
		BKs: dkgResult.BKs,
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
	return dkgResult, nil
}

// NewMockDB implementation using mock
func NewMockDB() types.StoreDB {
	return &MockDB{}
}
