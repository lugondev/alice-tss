package server

import (
	"alice-tss/pb"
	"alice-tss/service"
	"alice-tss/types"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/getamis/sirius/log"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/libp2p/go-libp2p/core/host"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

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

type TssPeerService struct {
	Pm        *peer.P2PManager
	BadgerFsm *peer.BadgerFSM
	TssCaller *TssCaller
}

func (t *TssPeerService) SignMessage(_ context.Context, args PingArgs, _ *PingReply) error {
	log.Info("RPC server", "SignMessage", "called", "args", args)
	var signRequest pb.SignRequest
	err := UnmarshalRequest(args.Data, &signRequest)
	if err != nil {
		return errors.New("invalid message, cannot unmarshal")
	}

	hash := utils.ToHexHash([]byte(signRequest.Message))
	pm := t.Pm.ClonePeerManager(peer.GetProtocol(hash))

	_, err = t.TssCaller.SignMessage(pm, &signRequest, nil)
	return err
}

func (t *TssPeerService) Reshare(_ context.Context, args PingArgs, _ *PingReply) error {
	log.Info("RPC server", "Reshare", "called", "args", args)
	var reshareRequest pb.ReshareRequest
	err := UnmarshalRequest(args.Data, &reshareRequest)
	if err != nil {
		return errors.New("invalid message, cannot unmarshal")
	}

	signerCfg, err := t.BadgerFsm.GetSignerConfig(reshareRequest.Hash, reshareRequest.Pubkey)
	if err != nil {
		log.Error("GetSignerConfig", "err", err)
		return err
	}

	pm := t.Pm.ClonePeerManager(peer.GetProtocol(reshareRequest.Hash))
	reshare, err := service.NewReshareService(&types.ReshareConfig{
		Threshold: 2,
		Share:     signerCfg.Share,
		Pubkey:    signerCfg.Pubkey,
		BKs:       signerCfg.BKs,
	}, pm, pm.Host, reshareRequest.Hash, t.BadgerFsm)
	if err != nil {
		log.Error("NewSignerService", "err", err)
		return err
	}

	log.Info("Stream Test", "service process", "called")
	go reshare.Process()

	return nil
}

func (t *TssPeerService) RegisterDKG(_ context.Context, argType PingArgs, _ *PingReply) error {
	log.Info("RegisterDKG")

	pm := t.Pm.ClonePeerManager(peer.GetProtocol(string(argType.Data)))
	_, err := t.TssCaller.RegisterDKG(pm, string(argType.Data), nil)

	return err
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
