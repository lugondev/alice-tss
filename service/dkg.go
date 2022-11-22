package service

import (
	types2 "alice-tss/types"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"io"
	"strings"

	"alice-tss/peer"
	"alice-tss/utils"
)

type Dkg struct {
	config  *types2.DKGConfig
	pm      *peer.P2PManager
	storeDB types2.StoreDB
	done    chan struct{}

	dkg  *dkg.DKG
	hash string
}

func NewDkgService(config *types2.DKGConfig, pm *peer.P2PManager, hash string, storeDB types2.StoreDB) (*Dkg, error) {
	s := &Dkg{
		config:  config,
		pm:      pm,
		storeDB: storeDB,
		done:    make(chan struct{}),
	}
	log.Warn("new DKG", "config", config, "hash", hash, "port", s.getPort())

	// Create dkg
	d, err := dkg.NewDKG(utils.GetCurve(), pm, config.Threshold, config.Rank, s)
	if err != nil {
		log.Warn("Cannot create a new DKG", "config", config, "err", err)
		return nil, err
	}
	s.dkg = d
	s.hash = hash
	streamHandlerProtocol := peer.GetProtocol(hash)
	if strings.Contains(hash, "-") {
		streamHandlerProtocol = peer.GetProtocol(strings.Split(hash, "-")[0])
	}

	pm.Host.SetStreamHandler(streamHandlerProtocol, func(stream network.Stream) {
		s.Handle(stream)
	})

	return s, nil
}

func (p *Dkg) getPort() string {
	addr := p.pm.Host.Addrs()[0].String()

	return addr[len(addr)-5:]
}

func (p *Dkg) GetResult() (*dkg.Result, error) {
	return p.dkg.GetResult()
}

func (p *Dkg) Handle(s network.Stream) {
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
	err = p.dkg.AddMessage(data.GetId(), data)
	if err != nil {
		log.Warn("Cannot add message to DKG", "err", err)
		return
	}
}

func (p *Dkg) Process() {
	// 1. Start a DKG process.
	p.dkg.Start()
	defer p.dkg.Stop()

	// 2. Wait the dkg is done or failed
	<-p.done
}

func (p *Dkg) closeDone() {
	close(p.done)
	p.pm.Host.RemoveStreamHandler(peer.GetProtocol(p.hash))
}

func (p *Dkg) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("Dkg failed", "old", oldState.String(), "new", newState.String())
		close(p.done)
		return
	} else if newState == types.StateDone {
		log.Info("Dkg done", "old", oldState.String(), "new", newState.String())
		result, err := p.dkg.GetResult()
		close(p.done)

		if err == nil {
			log.Debug("Register dkg", "result", result)
			if err := p.storeDB.SaveDKGResultData(p.hash, result); err != nil {
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
