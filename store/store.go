package store

import (
	"alice-tss/types"
	"crypto/ecdsa"
	"errors"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/reshare"
	"github.com/getamis/sirius/log"
)

type HandlerData interface {
	SaveDKGResultData(hash string, result *dkg.Result) error
	GetSignerConfig(hash, pubkey string) (*types.SignerConfig, error)
	UpdateDKGResultData(hash string, result *reshare.Result) error
	SaveSignerResultData(hash string, result types.RVSignature) error
	GetDKGResultData(hash string) (*types.DKGResult, error)
	Defer()
}

func NewStoreHandler(config types.StoreConfig, privateKey *ecdsa.PrivateKey) (HandlerData, error) {
	switch config.Type {
	case types.StoreTypeMock:
		log.Info("Store type is mock")
		return NewMockDB(), nil
	case types.StoreTypeBadger:
		if config.Path == "" {
			return nil, errors.New("badger path is empty")
		}
		if privateKey == nil {
			return nil, errors.New("badger private key is nil")
		}
		log.Info("Store type is badger", "path", config.Path)
		return NewBadgerDB(config.Path, privateKey), nil
	default:
		log.Info("Store type is mock")
		return NewMockDB(), nil
	}

}
