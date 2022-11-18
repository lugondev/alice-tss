package peer

import (
	"alice-tss/types"
	"alice-tss/utils"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/reshare"
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

	data := &types.DKGResult{
		Address:   crypto.PubkeyToAddress(*result.PublicKey.ToPubKey()),
		Share:     encryptedShare,
		PublicKey: hex.EncodeToString(pubkey),
		BKs:       map[string]types.BK{},
		Pubkey: types.Pubkey{
			X: hex.EncodeToString(result.PublicKey.GetX().Bytes()),
			Y: hex.EncodeToString(result.PublicKey.GetY().Bytes()),
		},
	}
	for s, parameter := range result.Bks {
		data.BKs[s] = types.BK{
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

// UpdateDKGResultData update dkg reshare data
func (fsm BadgerFSM) UpdateDKGResultData(hash string, result *reshare.Result) error {
	oldDkg, err := fsm.GetDKGResultData(hash)
	if err != nil {
		log.Error("GetDKGResultData", "err", err)
		return err
	}
	log.Info("UpdateDKGResultData", "hash", hash)

	encryptedShare, err := utils.Encrypt(
		common.Bytes2Hex(result.Share.Bytes()),
		crypto.FromECDSA(fsm.privateKey),
		oldDkg.PublicKey)

	if err != nil {
		log.Error("UpdateDKGResultData", "err", err)
		return err
	}

	oldDkg.Share = encryptedShare

	err = fsm.Set(hash, oldDkg)
	if err != nil {
		return err
	}
	return nil
}

// SaveSignerResultData save cmd result data
func (fsm BadgerFSM) SaveSignerResultData(hash string, result types.RVSignature) error {
	//log.Info("SaveSignerResultData", "hash", hash, "result", result)

	err := fsm.Set(hash, result)
	if err != nil {
		return err
	}
	return nil
}

// GetDKGResultData get dkg result data
func (fsm BadgerFSM) GetDKGResultData(hash string) (*types.DKGResult, error) {
	data, err := fsm.Get(hash)
	log.Info("GetDKGResultData", "hash", hash, "data", data)
	if err != nil {
		return nil, err
	}
	var result types.DKGResult
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

// GetSignerConfig get cmd config
func (fsm BadgerFSM) GetSignerConfig(hash, pubkey string) (*types.SignerConfig, error) {
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

	signerCfg := &types.SignerConfig{
		Share: big.NewInt(0).SetBytes(common.FromHex(share)).String(),
		Pubkey: types.Pubkey{
			X: big.NewInt(0).SetBytes(common.FromHex(resultDKG.Pubkey.X)).String(),
			Y: big.NewInt(0).SetBytes(common.FromHex(resultDKG.Pubkey.Y)).String(),
		},
		BKs: resultDKG.BKs,
	}

	return signerCfg, nil
}
