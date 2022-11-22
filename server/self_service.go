package server

import (
	"alice-tss/pb"
	"alice-tss/peer"
	"alice-tss/utils"
	"fmt"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/signer"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/host"
	peer2 "github.com/libp2p/go-libp2p/core/peer"
	"time"
)

type SelfService struct {
	hosts   [3]*host.Host
	peerIDs [3]peer2.ID
}

const (
	PortSelf1 = 11111
	PortSelf2 = 11112
	PortSelf3 = 11113
)

func (s *SelfService) RegisterDKG(tssCaller *TssCaller, hash string) (*dkg.Result, error) {
	pms := s.CreatePm(hash)

	go func() {
		if _, err := tssCaller.RegisterDKG(pms[1], fmt.Sprintf("%s-%d", hash, 1), nil); err != nil {
			log.Error("RPC server 2", "RegisterSelfDKG 2", err)
		}
	}()
	go func() {
		if _, err := tssCaller.RegisterDKG(pms[2], fmt.Sprintf("%s-%d", hash, 2), nil); err != nil {
			log.Error("RPC server 3", "RegisterSelfDKG 3", err)
		}
	}()

	return tssCaller.RegisterDKG(pms[0], fmt.Sprintf("%s-%d", hash, 0), func() error {
		return nil
	})
}

func (s *SelfService) SignMessage(tssCaller *TssCaller, dataRequestSign *pb.SignRequest) (*signer.Result, error) {
	hash := utils.ToHexHash([]byte(dataRequestSign.Message))
	pms := s.CreatePm(hash)

	go func() {
		dataRequestSignX := dataRequestSign
		dataRequestSignX.Hash = fmt.Sprintf("%s-%d", hash, 1)
		if _, err := tssCaller.SignMessage(pms[1], dataRequestSignX, nil); err != nil {
			log.Error("RPC server 2", "SelfSignMessage 2", err)
		}
	}()
	go func() {
		dataRequestSignX := dataRequestSign
		dataRequestSignX.Hash = fmt.Sprintf("%s-%d", hash, 2)
		if _, err := tssCaller.SignMessage(pms[2], dataRequestSignX, nil); err != nil {
			log.Error("RPC server 3", "SelfSignMessage 3", err)
		}
	}()

	dataRequestSignX := dataRequestSign
	dataRequestSignX.Hash = fmt.Sprintf("%s-%d", hash, 0)
	return tssCaller.SignMessage(pms[0], dataRequestSignX, func() error {
		return nil
	})
}

func (s *SelfService) CreatePm(protocolId string) [3]*peer.P2PManager {
	var pms [3]*peer.P2PManager
	for i, peerID := range s.peerIDs {
		pms[i] = peer.NewPeerManager(peerID.String(), *s.hosts[i], peer.GetProtocol(protocolId))
		if err := peer.SetupDiscovery(pms[i]); err != nil {
			log.Error("SetupDiscovery", "index", i, "peerID", peerID, "err", err)
		}
	}

	for pms[0].NumPeers() < 2 || pms[1].NumPeers() < 2 || pms[2].NumPeers() < 2 {
		log.Info("Waiting for peers...")
		time.Sleep(1 * time.Second)
	}

	return pms
}

func NewSelfService() *SelfService {
	host1, pid1, _ := peer.MakeBasicHostByID(PortSelf1)
	host2, pid2, _ := peer.MakeBasicHostByID(PortSelf2)
	host3, pid3, _ := peer.MakeBasicHostByID(PortSelf3)

	return &SelfService{
		hosts:   [3]*host.Host{&host1, &host2, &host3},
		peerIDs: [3]peer2.ID{pid1, pid2, pid3},
	}
}
