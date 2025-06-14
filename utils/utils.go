package utils

import (
	"alice-tss/types"
	"crypto/ecdsa"
	cryptoElliptic "crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getamis/alice/crypto/birkhoffinterpolation"
	"github.com/getamis/alice/crypto/ecpointgrouplaw"
	"github.com/getamis/alice/crypto/elliptic"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/sirius/log"
)

var (
	// ErrConversion for big int conversion error
	ErrConversion = errors.New("conversion error")
)

// GetCurve returns the curve we used in this example.
func GetCurve() elliptic.Curve {
	// For simplicity, we use S256 curve.
	return elliptic.Secp256k1()
}

// ConvertDKGResult converts DKG result from config.
func ConvertDKGResult(cfgPubkey types.Pubkey, cfgShare string, cfgBKs map[string]types.BK) (*dkg.Result, error) {
	// Build public key.
	x, ok := new(big.Int).SetString(cfgPubkey.X, 10)
	if !ok {
		log.Error("Cannot convert string to big int", "x", cfgPubkey.X)
		return nil, ErrConversion
	}
	y, ok := new(big.Int).SetString(cfgPubkey.Y, 10)
	if !ok {
		log.Error("Cannot convert string to big int", "y", cfgPubkey.Y)
		return nil, ErrConversion
	}
	pubkey, err := ecpointgrouplaw.NewECPoint(GetCurve(), x, y)
	if err != nil {
		log.Error("Cannot get public key", "err", err)
		return nil, err
	}

	// Build share.
	share, ok := new(big.Int).SetString(cfgShare, 10)
	if !ok {
		log.Error("Cannot convert string to big int", "share", share)
		return nil, ErrConversion
	}

	dkgResult := &dkg.Result{
		PublicKey: pubkey,
		Share:     share,
		Bks:       make(map[string]*birkhoffinterpolation.BkParameter),
	}

	// Build bks.
	for peerID, bk := range cfgBKs {
		x, ok := new(big.Int).SetString(bk.X, 10)
		if !ok {
			log.Error("Cannot convert string to big int", "x", bk.X)
			return nil, ErrConversion
		}
		dkgResult.Bks[peerID] = birkhoffinterpolation.NewBkParameter(x, bk.Rank)
	}

	return dkgResult, nil
}

func EthSignMessage(data []byte) []byte {
	msg := fmt.Sprintf("\u0019Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}

func GetPrivateKeyFromKeystore(keyFile, pass string) (*ecdsa.PrivateKey, error) {
	keyJson, err := os.ReadFile(keyFile)
	if err != nil {
		log.Error("Failed to read keystore file", "file", keyFile, "error", err)
		return nil, err
	}

	keyWrapper, err := keystore.DecryptKey(keyJson, pass)
	if err != nil {
		log.Error("Failed to decrypt keystore", "error", err)
		return nil, err
	}

	log.Info("Loaded keystore", "address", keyWrapper.Address.String())

	return keyWrapper.PrivateKey, nil
}

func ToEcdsaP256(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	secp256k1N, _ := new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = cryptoElliptic.P256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

func ToHex(b []byte) string {
	enc := make([]byte, len(b)*2+2)
	copy(enc, "0x")
	hex.Encode(enc[2:], b)
	return string(enc)
}

func ToHexHash(b []byte) string {
	return ToHex(crypto.Keccak256(b))
}

// RandomHash generates a cryptographically secure random hash using current time and crypto/rand.
func RandomHash() string {
	timeNow := time.Now()
	// Use crypto/rand for better security instead of math/rand
	randBytes := make([]byte, 8)
	if _, err := rand.Read(randBytes); err != nil {
		// Fallback to time-based randomness if crypto/rand fails
		rnd := mathrand.Intn(1000000)
		return ToHexHash([]byte(fmt.Sprintf("%s%d", timeNow.String(), rnd)))
	}
	return ToHexHash(append([]byte(timeNow.String()), randBytes...))
}
