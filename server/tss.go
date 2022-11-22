package server

import (
	"alice-tss/pb"
	"alice-tss/peer"
	tssService "alice-tss/service"
	"alice-tss/types"

	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/signer"
	"github.com/getamis/sirius/log"
)

type TssCaller struct {
	StoreDB types.StoreDB
}

func (t *TssCaller) SignMessage(pm *peer.P2PManager, signRequest *pb.SignRequest, call2peer func() error) (*signer.Result, error) {
	log.Info("SignMessage", "hash", signRequest.Hash, "pubkey", signRequest.Pubkey)

	signerCfg, err := t.StoreDB.GetSignerConfig(signRequest.Hash, signRequest.Pubkey)
	if err != nil {
		log.Error("GetSignerConfig", "err", err)
		return nil, err
	}

	service, err := tssService.NewSignerService(signerCfg, pm, t.StoreDB, signRequest.Message)
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

func (t *TssCaller) GetSignerConfig(signRequest *pb.SignRequest) (*types.SignerConfig, error) {
	signerCfg, err := t.StoreDB.GetSignerConfig(signRequest.Hash, signRequest.Pubkey)
	if err != nil {
		log.Error("GetSignerConfig", "err", err)
		return nil, err
	}

	return signerCfg, nil
}

func (t *TssCaller) Reshare(pm *peer.P2PManager, reshareRequest *pb.ReshareRequest, call2peer func() error) error {
	signerCfg, err := t.StoreDB.GetSignerConfig(reshareRequest.Hash, reshareRequest.Pubkey)
	if err != nil {
		log.Error("GetSignerConfig", "err", err)
		return err
	}

	service, err := tssService.NewReshareService(&types.ReshareConfig{
		Threshold: 2,
		Share:     signerCfg.Share,
		Pubkey:    signerCfg.Pubkey,
		BKs:       signerCfg.BKs,
	}, pm, reshareRequest.Hash, t.StoreDB)
	if err != nil {
		log.Error("NewReshareService", "err", err)
		return err
	}

	if call2peer != nil {
		if err := call2peer(); err != nil {
			log.Error("NewReshareService", "err", err)
			return err
		}
	}

	go service.Process()

	return nil
}

func (t *TssCaller) RegisterDKG(pm *peer.P2PManager, hash string, call2peer func() error) (*dkg.Result, error) {
	cfg := &types.DKGConfig{
		Rank:      0,
		Threshold: pm.NumPeers(),
	}
	log.Info("RegisterDKG", "numPeers", pm.NumPeers(), "rank", cfg.Rank)

	service, err := tssService.NewDkgService(cfg, pm, hash, t.StoreDB)
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
