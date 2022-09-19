package dkg

import (
	"io/ioutil"

	"alice-tss/utils"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
)

type service struct {
	config *DKGConfig
	pm     types.PeerManager

	dkg  *dkg.DKG
	done chan struct{}
}

func NewService(config *DKGConfig, pm types.PeerManager) (*service, error) {
	s := &service{
		config: config,
		pm:     pm,
		done:   make(chan struct{}),
	}

	// Create dkg
	d, err := dkg.NewDKG(utils.GetCurve(), pm, config.Threshold, config.Rank, s)
	if err != nil {
		log.Warn("Cannot create a new DKG", "config", config, "err", err)
		return nil, err
	}
	s.dkg = d
	return s, nil
}

func (p *service) Handle(s network.Stream) {
	data := &dkg.Message{}
	buf, err := ioutil.ReadAll(s)
	if err != nil {
		log.Warn("Cannot read data from stream", "err", err)
		return
	}
	s.Close()

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

func (p *service) Process() {
	// 1. Start a DKG process.
	p.dkg.Start()
	defer p.dkg.Stop()

	// 2. Wait the dkg is done or failed
	<-p.done
}

func (p *service) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("Dkg failed", "old", oldState.String(), "new", newState.String())
		close(p.done)
		return
	} else if newState == types.StateDone {
		log.Info("Dkg done", "old", oldState.String(), "new", newState.String())
		result, err := p.dkg.GetResult()
		if err == nil {
			writeDKGResult(p.pm.SelfID(), result)
		} else {
			log.Warn("Failed to get result from DKG", "err", err)
		}
		close(p.done)
		return
	}
	log.Info("State changed", "old", oldState.String(), "new", newState.String())
}
