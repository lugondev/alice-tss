package service

import (
	"io/ioutil"

	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/reshare"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"alice-tss/config"
	"alice-tss/peer"
	"alice-tss/utils"
)

type ReshareService struct {
	config *config.ReshareConfig
	pm     types.PeerManager
	fsm    *peer.BadgerFSM

	reshare *reshare.Reshare
	done    chan struct{}

	hash       string
	hostClient *host.Host
}

func NewReshareService(config *config.ReshareConfig, pm types.PeerManager, hostClient *host.Host, hash string, badgerFsm *peer.BadgerFSM) (*ReshareService, error) {
	s := &ReshareService{
		config: config,
		pm:     pm,
		fsm:    badgerFsm,
		done:   make(chan struct{}),
	}

	// Reshare needs results from DKG.
	dkgResult, err := utils.ConvertDKGResult(config.Pubkey, config.Share, config.BKs)
	if err != nil {
		log.Warn("Cannot get DKG result", "err", err)
		return nil, err
	}

	// Create reshare
	s.reshare, err = reshare.NewReshare(pm, config.Threshold, dkgResult.PublicKey, dkgResult.Share, dkgResult.Bks, s)
	if err != nil {
		log.Warn("Cannot create a new reshare", "err", err)
		return nil, err
	}

	s.hostClient = hostClient
	s.hash = hash

	(*hostClient).SetStreamHandler(peer.GetProtocol(hash), func(stream network.Stream) {
		s.Handle(stream)
	})

	return s, nil
}

func (p *ReshareService) Handle(s network.Stream) {
	data := &reshare.Message{}
	buf, err := ioutil.ReadAll(s)
	if err != nil {
		log.Warn("Cannot read data from stream", "err", err)
		return
	}
	_ = s.Close()

	// unmarshal it
	err = proto.Unmarshal(buf, data)
	if err != nil {
		log.Error("Cannot unmarshal data", "err", err)
		return
	}

	log.Info("Received request", "from", s.Conn().RemotePeer())
	err = p.reshare.AddMessage(data)
	if err != nil {
		log.Warn("Cannot add message to reshare", "err", err)
		return
	}
}

func (p *ReshareService) Process() {
	// 1. Start a reshare process.
	p.reshare.Start()
	defer p.reshare.Stop()

	// 2. Wait reshare is done or failed
	<-p.done
}
func (p *ReshareService) closeDone() {
	close(p.done)
	(*p.hostClient).RemoveStreamHandler(peer.GetProtocol(p.hash))
}

func (p *ReshareService) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("Reshare failed", "old", oldState.String(), "new", newState.String())
		p.closeDone()
		return
	} else if newState == types.StateDone {
		log.Info("Reshare done", "old", oldState.String(), "new", newState.String())
		result, err := p.reshare.GetResult()
		if err == nil {
			//log.Info("reshare", "result", result)
			if err := p.fsm.UpdateDKGResultData(p.hash, result); err != nil {
				log.Error("Cannot update DKG result data", "err", err)
				return
			}
		} else {
			log.Warn("Failed to get result from reshare", "err", err)
		}
		p.closeDone()
		return
	}
	log.Info("State changed", "old", oldState.String(), "new", newState.String())
}