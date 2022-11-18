package server

import (
	"alice-tss/config"
	"alice-tss/pb/tss"
	"alice-tss/peer"
	service2 "alice-tss/service"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/signer"
	"github.com/getamis/sirius/log"
)

type TssCaller struct {
	BadgerFsm *peer.BadgerFSM
}

func (t *TssCaller) SignMessage(pm *peer.P2PManager, signRequest *tss.SignRequest, call2peer func() error) (*signer.Result, error) {
	signerCfg, err := t.BadgerFsm.GetSignerConfig(signRequest.Hash, signRequest.Pubkey)
	if err != nil {
		log.Error("GetSignerConfig", "err", err)
		return nil, err
	}

	service, err := service2.NewSignerService(signerCfg, pm, t.BadgerFsm, &pm.Host, signRequest.Message)
	if err != nil {
		log.Error("NewSignerService", "err", err)
		return nil, err
	}
	if call2peer != nil {
		if err := call2peer(); err != nil {
			return nil, err
		}
		service.Process()
		return service.GetResult()
	} else {
		go service.Process()
	}

	return nil, nil
}

func (t *TssCaller) Reshare(pm *peer.P2PManager, reshareRequest *tss.ReshareRequest, call2peer func() error) error {
	signerCfg, err := t.BadgerFsm.GetSignerConfig(reshareRequest.Hash, reshareRequest.Pubkey)
	if err != nil {
		log.Error("GetSignerConfig", "err", err)
		return err
	}

	service, err := service2.NewReshareService(&config.ReshareConfig{
		Threshold: 2,
		Share:     signerCfg.Share,
		Pubkey:    signerCfg.Pubkey,
		BKs:       signerCfg.BKs,
	}, pm, &pm.Host, reshareRequest.Hash, t.BadgerFsm)
	if err != nil {
		log.Error("NewReshareService", "err", err)
		return err
	}

	if call2peer != nil {
		if err := call2peer(); err != nil {
			return err
		}
	}

	go service.Process()

	return nil
}

func (t *TssCaller) RegisterDKG(pm *peer.P2PManager, hash string, call2peer func() error) (*dkg.Result, error) {
	cfg := &config.DKGConfig{
		Rank:      0,
		Threshold: pm.NumPeers(),
	}

	service, err := service2.NewDkgService(cfg, pm, &pm.Host, hash, t.BadgerFsm)
	if err != nil {
		log.Error("NewDkgService", "err", err)
		return nil, err
	}

	if call2peer != nil {
		if err := call2peer(); err != nil {
			return nil, err
		}
		service.Process()
		return service.GetResult()
	} else {
		go service.Process()
	}

	return nil, nil
}
