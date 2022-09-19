package main_test

import (
	"alice-tss/utils"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"testing"
)

func TestECDSA(t *testing.T) {
	r := "49305996660667747080665206993786445468000605506928750062204102477508275731194"
	s := "104573947396974654709884502830302414362190494746353716510105790268445988646023"

	rBig := new(big.Int)
	rBig.SetString(r, 10)
	fmt.Println("r:", common.Bytes2Hex(rBig.Bytes()))
	sBig := new(big.Int)
	sBig.SetString(s, 10)
	fmt.Println("s:", common.Bytes2Hex(sBig.Bytes()))
	x := new(big.Int)
	x.SetString("17336117875018124740261888937530006260396156587822066601411612153588608601941", 10)
	y := new(big.Int)
	y.SetString("21962150140477952594158381548285773330312292966097687763903683104393773465248", 10)

	publicKey := ecdsa.PublicKey{
		Curve: utils.GetCurve(),
		X:     x,
		Y:     y,
	}

	fmt.Printf("%x \n", publicKey)
	msg := "hello 123123zzzzz"
	hashMessage := utils.EthSignMessage([]byte(msg))
	fmt.Println("signed eth msg:", common.Bytes2Hex(hashMessage))
	status := ecdsa.Verify(&publicKey, hashMessage, rBig, sBig)
	fmt.Println("signature valid:", status)
	addressFromPubKey := crypto.PubkeyToAddress(publicKey)
	fmt.Println("address:", addressFromPubKey)
}

func TestPubkey(t *testing.T) {
	pubkey, err := crypto.UnmarshalPubkey(common.Hex2Bytes("040baaa4c80f6604c977efa80bb890dfe6a88bfa857ba9114eeec4a5a8768ca61e66c87cf35213bc1bde4fcd746c4f5b532e0dcfc29edbcdf2043f9f77d6606e32"))
	if err != nil {
		return
	}
	fmt.Println(pubkey)
	addressFromPubKey := crypto.PubkeyToAddress(*pubkey)
	fmt.Println("address:", addressFromPubKey)
}
