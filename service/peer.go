package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/getamis/sirius/log"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/libp2p/go-libp2p/core/host"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"alice-tss/config"
	"alice-tss/peer"
	"alice-tss/utils"
)

type PeerArgs struct {
	PeerAddrTarget string
	SvcName        string
	SvcMethod      string
	Args           PingArgs
}

type PingArgs struct {
	Data []byte
}

type PingReply struct {
}

type TssService struct {
	Pm        *peer.P2PManager
	Config    *config.SignerConfig
	BadgerFsm *peer.BadgerFSM
}

func (t *TssService) SignMessage(_ context.Context, args PingArgs, _ *PingReply) error {
	log.Info("RPC server", "SignMessage", "called", "args", args)
	var data DataRequestSign
	err := json.Unmarshal(args.Data, &data)
	if err != nil {
		return err
	}
	hash := utils.ToHexHash([]byte(data.Message))
	signerCfg, err := t.BadgerFsm.GetSignerConfig(data.Hash, data.Pubkey)
	if err != nil {
		log.Error("GetSignerConfig", "err", err)
		return err
	}

	pm := t.Pm.ClonePeerManager(peer.GetProtocol(hash))
	service, err := NewSignerService(signerCfg, pm, t.BadgerFsm, &pm.Host, data.Message)
	if err != nil {
		log.Error("NewSignerService", "err", err)
		return err
	}

	log.Info("Stream Test", "service process", "called")
	go service.Process()

	return nil
}

func (t *TssService) RegisterDKG(_ context.Context, argType PingArgs, _ *PingReply) error {
	log.Info("RegisterDKG")

	cfg := &config.DKGConfig{
		Rank:      0,
		Threshold: t.Pm.NumPeers(),
	}

	pm := t.Pm.ClonePeerManager(peer.GetProtocol(string(argType.Data)))
	service, err := NewDkgService(cfg, pm, &pm.Host, string(argType.Data), t.BadgerFsm)
	if err != nil {
		log.Error("NewDkgService", "err", err)
		return err
	}

	log.Info("Stream Test", "service process", "called")
	go service.Process()

	return nil
}

func MsgToPeer(client host.Host, data PeerArgs) (*PingReply, error) {
	ma, err := multiaddr.NewMultiaddr(data.PeerAddrTarget)
	if err != nil {
		fmt.Println("New Multi addr:", err)
		return nil, err
	}
	peerInfo, err := libp2pPeer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		fmt.Println("Addr Info From P2p Addr:", err)
		return nil, err
	}
	err = client.Connect(context.Background(), *peerInfo)
	if err != nil {
		fmt.Println("Connect to peer err:", err)
		return nil, err
	}

	rpcClient := gorpc.NewClient(client, peer.ProtocolId)

	var reply PingReply
	err = rpcClient.Call(peerInfo.ID, data.SvcName, data.SvcMethod, data.Args, &reply)
	if err != nil {
		fmt.Println("Cannot call to peer:", err)
		return nil, err
	}
	return &reply, nil
}

func SendToPeer(client host.Host, data PeerArgs, wg *sync.WaitGroup) (*PingReply, error) {
	defer wg.Done()

	for {
		// Connect the host to the peer.
		reply, err := MsgToPeer(client, data)
		if err != nil {
			log.Warn("Failed to sent to peer", "to", client.ID().String(), "err", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Debug("Successfully connect to peer")
		return reply, nil
	}
}
