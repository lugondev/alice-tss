package signer

import (
	"alice-tss/config"
	"alice-tss/peer"
	"alice-tss/utils"
	"context"
	"fmt"
	"github.com/getamis/sirius/log"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/libp2p/go-libp2p/core"
	"github.com/libp2p/go-libp2p/core/host"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"sync"
	"time"
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
	service   *Service
	pm        *peer.P2PManager
	config    *config.SignerConfig
	badgerFsm *peer.BadgerFSM
}

func (t *PingService) Process(_ context.Context, argType PingArgs, replyType *PingReply) error {
	replyType.Key = argType.ID

	t.pm.EnsureAllConnected()
	go t.service.Process()
	return nil
}

func (t *PingService) PrepareDataToSign(_ context.Context, argType PingArgs, replyType *PingReply) error {
	hash := utils.ToHexHash(argType.Data)
	log.Info("PrepareDataToSign", "id", argType.ID, "data", string(argType.Data), "hash", hash)
	replyType.Key = argType.ID
	replyType.Data = []byte(hash)

	if err := t.service.CreateSigner(string(argType.Data)); err != nil {
		fmt.Println("CreateSigner err", err)
		return err
	}

	if err := t.badgerFsm.Set(hash, "1"); err != nil {
		return err
	}
	return nil
}

func (t *PingService) Test(_ context.Context, argType PingArgs, replyType *PingReply) error {
	hash := utils.ToHexHash(argType.Data)
	log.Info("PrepareDataToSign", "id", argType.ID, "data", string(argType.Data), "hash", hash)
	replyType.Key = argType.ID
	replyType.Data = []byte(hash)

	if err := t.badgerFsm.Set(hash, "1"); err != nil {
		return err
	}

	//service, err := NewService(t.config, t.pm, t.badgerFsm)
	//if err != nil {
	//	log.Error("NewService", "err", err)
	//	return err
	//}

	//t.pm.Host.SetStreamHandler(peer.SignerProtocol, func(s network.Stream) {
	//	log.Info("Stream handler from other peer", "protocol", s.Protocol(), "peer", s.Conn().LocalPeer())
	//	if service.signer != nil {
	//		service.Handle(s)
	//	}
	//})

	if err := t.service.CreateSigner(string(argType.Data)); err != nil {
		fmt.Println("CreateSigner err", err)
		return err
	} else {
		log.Info("Stream Test", "service process", "called")
		go t.service.Process()
	}

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

	rpcClient := gorpc.NewClient(client, data.ProtocolID)

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
