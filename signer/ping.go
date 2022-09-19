package signer

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/libp2p/go-libp2p/core"
	"github.com/multiformats/go-multiaddr"
	"log"

	"alice-tss/config"
	"alice-tss/peer"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/libp2p/go-libp2p/core/host"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
)

type PeerArgs struct {
	PeerAddrTarget string
	SvcName        string
	SvcMethod      string
	Args           PingArgs
	ProtocolID     core.ProtocolID
}

type PingArgs struct {
	ID   string
	Data []byte
}
type PingReply struct {
	Key  string
	Data []byte
}
type PingService struct {
	service *Service
	pm      *peer.PeerManager
	config  *config.SignerConfig
}

func (t *PingService) Ping(_ context.Context, argType PingArgs, replyType *PingReply) error {
	log.Println("Received a Ping call:", hex.EncodeToString(argType.Data))
	replyData := []byte("reply for ID " + argType.ID + " with data:" + hex.EncodeToString(argType.Data))
	replyType.Data = replyData
	replyType.Key = argType.ID

	t.pm.EnsureAllConnected()
	go t.service.Process()
	return nil
}

func (t *PingService) PrepareMsg(_ context.Context, argType PingArgs, replyType *PingReply) error {
	log.Println("PrepareMsg called:", hex.EncodeToString(argType.Data))
	replyData := []byte("reply for key " + argType.ID + " with data:" + hex.EncodeToString(argType.Data))
	replyType.Data = replyData
	replyType.Key = argType.ID

	if err := t.service.CreateSigner(t.pm, t.config, string(argType.Data)); err != nil {
		fmt.Println("CreateSigner err", err)
		return err
	}
	return nil
}

func SentToPeer(client host.Host, data PeerArgs) (*PingReply, error) {
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

	rpcClient := gorpc.NewClient(client, data.ProtocolID)

	var reply PingReply
	err = rpcClient.Call(peerInfo.ID, data.SvcName, data.SvcMethod, data.Args, &reply)
	if err != nil {
		fmt.Println("Cannot call to peer:", err)
		return nil, err
	}
	return &reply, nil
}
