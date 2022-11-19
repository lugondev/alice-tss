package service

import (
	"alice-tss/peer"
	types2 "alice-tss/types"
	"alice-tss/utils"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/getamis/alice/crypto/homo/paillier"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/signer"
	"github.com/getamis/alice/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"io"
)

type Signer struct {
	config *types2.SignerConfig
	pm     types.PeerManager
	fsm    *peer.BadgerFSM

	signer *signer.Signer
	done   chan struct{}

	hash       string
	hostClient host.Host
}

func NewSignerService(
	config *types2.SignerConfig,
	pm types.PeerManager,
	badgerFsm *peer.BadgerFSM,
	hostClient host.Host,
	msg string,
) (*Signer, error) {
	s := &Signer{
		config: config,
		pm:     pm,
		fsm:    badgerFsm,
		done:   make(chan struct{}),
	}

	log.Info("Service call")
	if err := s.createSigner(msg); err != nil {
		return nil, err
	}
	hash := utils.ToHexHash([]byte(msg))
	s.hash = hash
	s.hostClient = hostClient

	hostClient.SetStreamHandler(peer.GetProtocol(hash), func(stream network.Stream) {
		s.Handle(stream)
	})

	return s, nil
}

func (p *Signer) createSigner(msg string) error {
	// For simplicity, we use Paillier algorithm in cmd.
	newPaillier, err := paillier.NewPaillier(2048)
	if err != nil {
		log.Warn("Cannot create a paillier function", "err", err)
		return err
	}

	dkgResult, err := utils.ConvertDKGResult(p.config.Pubkey, p.config.Share, p.config.BKs)
	if err != nil {
		log.Warn("Cannot get DKG result", "err", err)
		return err
	}

	log.Info("Signer created", "msg", msg)
	//byteMessage := []byte(msg)
	byteMessage := common.Hex2Bytes(msg)
	newSigner, err := signer.NewSigner(p.pm, dkgResult.PublicKey, newPaillier, dkgResult.Share, dkgResult.Bks, byteMessage, p)
	if err != nil {
		log.Warn("Cannot create a new cmd", "err", err)
		return err
	}
	p.signer = newSigner

	return nil
}

func (p *Signer) GetResult() (*signer.Result, error) {
	return p.signer.GetResult()
}

func (p *Signer) Handle(s network.Stream) {
	if p.signer == nil {
		log.Warn("Signer is not created")
		return
	}
	data := &signer.Message{}
	buf, err := io.ReadAll(s)
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

	log.Info("Received request", "from", s.Conn().RemotePeer())
	err = p.signer.AddMessage(data.GetId(), data)
	if err != nil {
		log.Warn("Cannot add message to cmd", "err", err)
		return
	}
}

func (p *Signer) Process() {
	// 1. Start a cmd process.
	p.signer.Start()
	log.Info("Signer process", "action", "start")
	defer func() {
		log.Info("Signer process", "action", "stop")
		p.signer.Stop()
	}()

	// 2. Wait the cmd is done or failed
	<-p.done
}

func (p *Signer) closeDone() {
	close(p.done)
	p.hostClient.RemoveStreamHandler(peer.GetProtocol(p.hash))
}

func (p *Signer) OnStateChanged(oldState types.MainState, newState types.MainState) {
	if newState == types.StateFailed {
		log.Error("Signer failed", "old", oldState.String(), "new", newState.String())
		p.closeDone()
		return
	} else if newState == types.StateDone {
		log.Info("Signer done", "old", oldState.String(), "new", newState.String())
		result, err := p.signer.GetResult()
		if err == nil {
			//log.Info("signed", "result", result)

			if err := p.fsm.SaveSignerResultData(p.hash, types2.RVSignature{
				R:    hex.EncodeToString(result.R.Bytes()),
				S:    hex.EncodeToString(result.S.Bytes()),
				Hash: p.hash,
			}); err != nil {
				log.Error("Cannot save sign result", "err", err)
				return
			}
		} else {
			log.Warn("Failed to get result from cmd", "err", err)
		}
		p.closeDone()
		return
	}
	log.Info("State changed", "old", oldState.String(), "new", newState.String())
}
