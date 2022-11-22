package types

import (
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/reshare"
)

type StoreDB interface {
	SaveDKGResultData(hash string, result *dkg.Result) error
	GetSignerConfig(hash, pubkey string) (*SignerConfig, error)
	UpdateDKGResultData(hash string, result *reshare.Result) error
	SaveSignerResultData(hash string, result RVSignature) error
	GetDKGResultData(hash string) (*DKGResult, error)
	Defer()
}
