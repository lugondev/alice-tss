package service

import (
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"io"

	"alice-tss/config"
	"alice-tss/peer"
	"alice-tss/utils"
)

type DkgService struct {
	config *config.DKGConfig
	pm     types.PeerManager
	fsm    *peer.BadgerFSM

	dkg  *dkg.DKG
	done chan struct{}

	hash       string
	hostClient *host.Host
}

func NewDkgService(config *config.DKGConfig, pm types.PeerManager, hostClient *host.Host, hash string, badgerFsm *peer.BadgerFSM) (*DkgService, error) {
	s := &DkgService{
		config: config,
		pm:     pm,
		fsm:    badgerFsm,
		done:   make(chan struct{}),
	}
	log.Warn("new DKG", "config", config, "hash", hash)

	// Create dkg
	d, err := dkg.NewDKG(utils.GetCurve(), pm, config.Threshold, config.Rank, s)
	if err != nil {
		log.Warn("Cannot create a new DKG", "config", config, "err", err)
		return nil, err
	}
	s.dkg = d

	s.hostClient = hostClient
	s.hash = hash

	(*hostClient).SetStreamHandler(peer.GetProtocol(hash), func(stream network.Stream) {
		s.Handle(stream)
	})

	return s, nil
}

func (p *DkgService) GetResult() (*dkg.Result, error) {
	return p.dkg.GetResult()
}

func (p *DkgService) Handle(s network.Stream) {
	data := &dkg.Message{}
	buf, err := io.ReadAll(s)
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
	err = p.dkg.AddMessage(data)
	if err != nil {
		log.Warn("Cannot add message to DKG", "err", err)
		return
	}
}

func (p *DkgService) Process() {
	// 1. Start a DKG process.
	p.dkg.Start()
	defer p.dkg.Stop()

	// 2. Wait the dkg is done or failed
	<-p.done
}

func (p *DkgService) closeDone() {
	close(p.done)
	(*p.hostClient).RemoveStreamHandler(peer.GetProtocol(p.hash))
}

func (p *DkgService) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("Dkg failed", "old", oldState.String(), "new", newState.String())
		close(p.done)
		return
	} else if newState == types.StateDone {
		log.Info("Dkg done", "old", oldState.String(), "new", newState.String())
		result, err := p.dkg.GetResult()
		close(p.done)

		if err == nil {
			//log.Info("Register dkg", "result", result)
			if err := p.fsm.SaveDKGResultData(p.hash, result); err != nil {
				log.Error("Cannot save dkg result", "err", err)
				return
			}
		} else {
			log.Warn("Failed to get result from DKG", "err", err)
		}
		return
	}
	log.Info("State changed", "old", oldState.String(), "new", newState.String())
}
