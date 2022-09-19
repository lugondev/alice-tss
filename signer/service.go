package signer

import (
	"io/ioutil"

	"github.com/getamis/alice/crypto/homo/paillier"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/signer"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"

	"alice-tss/config"
	"alice-tss/peer"
	"alice-tss/utils"
)

type Service struct {
	config *config.SignerConfig
	pm     types.PeerManager
	fsm    *peer.BadgerFSM

	signer *signer.Signer
	done   chan struct{}
}

func NewService(config *config.SignerConfig, pm types.PeerManager, badgerFsm *peer.BadgerFSM) (*Service, error) {
	s := &Service{
		config: config,
		pm:     pm,
		fsm:    badgerFsm,
		done:   make(chan struct{}),
	}
	log.Info("Service call")

	//Create newSigner
	//if err := s.CreateSigner(pm, config, "MsgSign"); err != nil {
	//	log.Warn("Cannot create a new newSigner", "err", err)
	//	return nil, err
	//}
	return s, nil
}

func (p *Service) closeDone() {
	_, ok := <-p.done
	if ok {
		close(p.done)
		p.done = make(chan struct{})
		p.signer = nil
	}
}

func (p *Service) CreateSigner(pm types.PeerManager, config *config.SignerConfig, msg string) error {
	// For simplicity, we use Paillier algorithm in signer.
	newPaillier, err := paillier.NewPaillier(2048)
	if err != nil {
		log.Warn("Cannot create a paillier function", "err", err)
		return err
	}

	dkgResult, err := utils.ConvertDKGResult(config.Pubkey, config.Share, config.BKs)
	if err != nil {
		log.Warn("Cannot get DKG result", "err", err)
		return err
	}

	log.Info("Signer created", "msg", msg)
	hashMessage := utils.EthSignMessage([]byte(msg))
	newSigner, err := signer.NewSigner(pm, dkgResult.PublicKey, newPaillier, dkgResult.Share, dkgResult.Bks, hashMessage, p)
	if err != nil {
		log.Warn("Cannot create a new signer", "err", err)
		return err
	}

	p.signer = newSigner
	return nil
}

func (p *Service) Handle(s network.Stream) {
	if p.signer == nil {
		log.Warn("Signer is not created")
		return
	}
	data := &signer.Message{}
	buf, err := ioutil.ReadAll(s)
	if err != nil {
		log.Warn("Cannot read data from stream", "err", err)
		return
	}
	_ = s.Close()
	err = proto.Unmarshal(buf, data)
	if err != nil {
		log.Error("Cannot unmarshal data", "err", err)
		return
	}

	log.Info("Received request", "from", s.Conn().RemotePeer(), "data", data.String())
	err = p.signer.AddMessage(data)
	if err != nil {
		log.Warn("Cannot add message to signer", "err", err)
		return
	}
}

func (p *Service) Process() {
	// 1. Start a signer process.
	p.signer.Start()
	log.Info("Signer process", "action", "start")
	defer func() {
		log.Info("Signer process", "action", "stop")
		p.signer.Stop()
	}()

	// 2. Wait the signer is done or failed
	<-p.done
}

func (p *Service) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("Signer failed", "old", oldState.String(), "new", newState.String())
		p.closeDone()
		return
	} else if newState == types.StateDone {
		log.Info("Signer done", "old", oldState.String(), "new", newState.String())
		result, err := p.signer.GetResult()
		if err == nil {
			writeSignerResult(p.pm.SelfID(), result)
		} else {
			log.Warn("Failed to get result from signer", "err", err)
		}
		p.closeDone()
		return
	}
	log.Info("State changed", "old", oldState.String(), "new", newState.String())
}
