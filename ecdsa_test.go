package main_test

import (
	"alice-tss/utils"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func TestRsHex(t *testing.T) {
	//"r":"7e535dd1f357021d5a7f03251a6e8536c0f86f1bc9a8cefa8522c02d0f8eb73d","s":"cfa84733cabaca52f84e41475e88163e7211ff01dd18aaa85af41a576b14343e"
	r := "7e535dd1f357021d5a7f03251a6e8536c0f86f1bc9a8cefa8522c02d0f8eb73d"
	s := "cfa84733cabaca52f84e41475e88163e7211ff01dd18aaa85af41a576b14343e"

	rBig := new(big.Int)
	rBig.SetString(r, 16)
	fmt.Println("r:", common.Bytes2Hex(rBig.Bytes()))
	sBig := new(big.Int)
	sBig.SetString(s, 16)
	fmt.Println("s:", common.Bytes2Hex(sBig.Bytes()))
	x := new(big.Int)
	//"x": "0c56c52d12cc437e8cb693efa16a768581fa3f96881f349af39ea4a5be19483f",
	//			"y": "ed36822b8f1710035830afb801bdbb72fa98bd9b4c95383fe6cce3ab783ce519",
	x.SetString("0c56c52d12cc437e8cb693efa16a768581fa3f96881f349af39ea4a5be19483f", 16)
	y := new(big.Int)
	y.SetString("ed36822b8f1710035830afb801bdbb72fa98bd9b4c95383fe6cce3ab783ce519", 16)

	publicKey := ecdsa.PublicKey{
		Curve: utils.GetCurve(),
		X:     x,
		Y:     y,
	}

	rs := append(rBig.Bytes(), sBig.Bytes()...)
	rs = append(rs, big.NewInt(28).Bytes()...)
	fmt.Println("rs:", common.Bytes2Hex(rs))
	fmt.Printf("%x \n", publicKey)
	msg := "013eadf0078f80119df16210a27dffcede2ecf692d22ba2174bcdc2a64d69ccd"
	status := ecdsa.Verify(&publicKey, common.Hex2Bytes(msg), rBig, sBig)
	fmt.Println("signature valid:", status)
	addressFromPubKey := crypto.PubkeyToAddress(publicKey)

	fmt.Println("address:", addressFromPubKey)
	xx := recoverSig("0x"+common.Bytes2Hex(rs), common.Hex2Bytes(msg))
	fmt.Println("address:", xx.String())
}

func recoverSig(sigHex string, msg []byte) common.Address {
	fmt.Println("msg len:", len(msg))
	sig := hexutil.MustDecode(sigHex)
	if sig[64] != 27 && sig[64] != 28 {
		return common.HexToAddress("0x")
	}
	sig[64] -= 27

	pubKey, err := crypto.SigToPub(msg, sig)
	if err != nil {
		fmt.Println("err:", err)
		return common.HexToAddress("0x")
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr
}

func TestPubkey(t *testing.T) {
	x := new(big.Int)
	x.SetString("f7d1a2b1dbed87f356ce9a65552fa88d017bde55f93e7cbbecf29f31a672b76f", 16)
	y := new(big.Int)
	y.SetString("7ae2cdf0f8b102593e92f6977fd3b7e2df43c7d66e48a3dc8a1f31e826357d10", 16)

	publicKey := ecdsa.PublicKey{
		Curve: utils.GetCurve(),
		X:     x,
		Y:     y,
	}
	addressFromPubKey := crypto.PubkeyToAddress(publicKey)
	fmt.Println("address:", addressFromPubKey)
}
