package peer

import (
	"alice-tss/config"
	"alice-tss/utils"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/sirius/log"
	"math/big"
)

// SaveDKGResultData save dkg result data
func (fsm BadgerFSM) SaveDKGResultData(hash string, result *dkg.Result) error {
	pubkey := crypto.CompressPubkey(result.PublicKey.ToPubKey())
	log.Info("SaveDKGResultData", "hash", hash, "pubkey", hex.EncodeToString(pubkey))

	encryptedShare, err := utils.Encrypt(
		common.Bytes2Hex(result.Share.Bytes()),
		crypto.FromECDSA(fsm.privateKey),
		hex.EncodeToString(pubkey))

	if err != nil {
		log.Error("SaveDKGResultData", "err", err)
		return err
	}

	data := &config.DKGResult{
		Address:   crypto.PubkeyToAddress(*result.PublicKey.ToPubKey()),
		Share:     encryptedShare,
		PublicKey: hex.EncodeToString(pubkey),
		BKs:       map[string]config.BK{},
		Pubkey: config.Pubkey{
			X: hex.EncodeToString(result.PublicKey.GetX().Bytes()),
			Y: hex.EncodeToString(result.PublicKey.GetY().Bytes()),
		},
	}
	for s, parameter := range result.Bks {
		data.BKs[s] = config.BK{
			X:    parameter.GetX().String(),
			Rank: parameter.GetRank(),
		}
	}

	err = fsm.Set(hash, data)
	if err != nil {
		return err
	}
	return nil
}

// SaveSignerResultData save signer result data
func (fsm BadgerFSM) SaveSignerResultData(hash string, result config.RVSignature) error {
	log.Info("SaveSignerResultData", "hash", hash, "result", result)

	err := fsm.Set(hash, result)
	if err != nil {
		return err
	}
	return nil
}

// GetDKGResultData get dkg result data
func (fsm BadgerFSM) GetDKGResultData(pubkey string) (*config.DKGResult, error) {
	data, err := fsm.Get(pubkey)
	log.Info("GetDKGResultData", "pubkey", pubkey, "data", data)
	if err != nil {
		return nil, err
	}
	var result config.DKGResult
	byteData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSignerConfig get signer config
func (fsm BadgerFSM) GetSignerConfig(hash, pubkey string) (*config.SignerConfig, error) {
	log.Info("GetSignerConfig", "hash", hash, "pubkey", pubkey)

	resultDKG, err := fsm.GetDKGResultData(hash)
	if err != nil {
		return nil, err
	}
	publicKey := &ecdsa.PublicKey{
		X: big.NewInt(0).SetBytes(common.FromHex(resultDKG.Pubkey.X)),
		Y: big.NewInt(0).SetBytes(common.FromHex(resultDKG.Pubkey.Y)),
	}
	if hex.EncodeToString(crypto.CompressPubkey(publicKey)) != pubkey {
		return nil, fmt.Errorf("pubkey not match")
	}

	share, err := utils.Decrypt(resultDKG.Share, crypto.FromECDSA(fsm.privateKey), pubkey)
	if err != nil {
		return nil, err
	}

	signerCfg := &config.SignerConfig{
		Share: big.NewInt(0).SetBytes(common.FromHex(share)).String(),
		Pubkey: config.Pubkey{
			X: big.NewInt(0).SetBytes(common.FromHex(resultDKG.Pubkey.X)).String(),
			Y: big.NewInt(0).SetBytes(common.FromHex(resultDKG.Pubkey.Y)).String(),
		},
		BKs: resultDKG.BKs,
	}

	return signerCfg, nil
}
