package utils

import (
	"alice-tss/config"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getamis/sirius/log"
)

type ResponseCheckSignature struct {
	IsValid     bool   `json:"isValid"`
	Message     string `json:"message"`
	HashMessage string `json:"hashMessage"`
	Address     string `json:"address"`
}

func CheckSignatureECDSA(msg string, signature config.RVSignature, pubkey string) (*ResponseCheckSignature, error) {
	rBig := big.NewInt(0).SetBytes(common.FromHex(signature.R))
	sBig := big.NewInt(0).SetBytes(common.FromHex(signature.S))

	pubkeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		log.Error("CheckSignatureECDSA", "err", err)
		return nil, err
	}
	publicKey, err := crypto.DecompressPubkey(pubkeyBytes)
	if err != nil {
		log.Error("Failed to decompress pubkey", "err", err, "pubkey", pubkey)
		return nil, err
	}

	hashedMsg := EthSignMessage([]byte(msg))

	status := ecdsa.Verify(publicKey, hashedMsg, rBig, sBig)
	return &ResponseCheckSignature{
		IsValid:     status,
		Message:     msg,
		HashMessage: hex.EncodeToString(hashedMsg),
		Address:     crypto.PubkeyToAddress(*publicKey).String(),
	}, nil
}
