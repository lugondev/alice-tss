package signer

import (
	"alice-tss/config"
	"alice-tss/peer"
	"alice-tss/utils"
	"errors"
	"fmt"
	"github.com/getamis/sirius/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	rpcjson "github.com/gorilla/rpc/v2/json2"
	"github.com/libp2p/go-libp2p/core/host"
	"net/http"
	"sync"
)

type RpcService struct {
	pm         *peer.P2PManager
	service    *Service
	hostClient host.Host
	config     *config.SignerConfig
	badgerFsm  *peer.BadgerFSM
}

type RpcPreDataArgs struct {
	Data string
}

type RpcProcessArgs struct {
}

type RpcReply struct {
	Data interface{}
}

func (h *RpcService) GetSay(_ *http.Request, args *RpcPreDataArgs, reply *RpcReply) error {
	reply.Data = "Rpc, " + args.Data + "!"
	return nil
}

func (h *RpcService) GetHash(_ *http.Request, args *RpcPreDataArgs, reply *RpcReply) error {
	log.Info("get hash request", "arg", args.Data)
	data, err := h.badgerFsm.Get(args.Data)
	if err != nil {
		return err
	}
	log.Info("get hash", "data", data)
	reply.Data = data
	return nil
}

func (h *RpcService) SignerStop(_ *http.Request, _ *RpcProcessArgs, _ *RpcReply) error {
	h.service.signer.Stop()
	return nil
}

func (h *RpcService) Process(_ *http.Request, _ *RpcProcessArgs, reply *RpcReply) error {
	log.Info("RPC server", "process", "called")
	reply.Data = "Processed"

	var wg sync.WaitGroup
	for _, peerId := range h.pm.PeerIDs() {
		wg.Add(1)
		peerAddrTarget := h.pm.Peers()[peerId]
		go func() {
			fmt.Println("send:", peerAddrTarget)
			peerReply, err := SendToPeer(h.pm.Host, PeerArgs{
				peerAddrTarget,
				"PingService",
				"Process",
				PingArgs{
					ID:   h.pm.Host.ID().String(),
					Data: []byte("msg"),
				},
				peer.ProtocolId,
			}, &wg)

			if err != nil {
				fmt.Println("send err", err)
				return
			}
			fmt.Println("reply:", peerReply)
		}()
	}
	wg.Wait()

	go h.service.Process()
	return nil
}

func (h *RpcService) Test(_ *http.Request, _ *RpcProcessArgs, reply *RpcReply) error {
	log.Info("RPC server", "test", "called")
	reply.Data = "Test"
	//service, err := NewService(h.config, h.pm, h.badgerFsm)
	//if err != nil {
	//	log.Error("NewService", "err", err)
	//	return err
	//}

	//h.pm.Host.SetStreamHandler(peer.SignerProtocol, func(s network.Stream) {
	//	log.Info("Stream handler", "protocol", s.Protocol(), "peer", s.Conn().LocalPeer())
	//	if service.signer != nil {
	//		service.Handle(s)
	//	}
	//})
	var wg sync.WaitGroup
	msgTest := fmt.Sprintf("%s-%d", "msg-test", 123123)
	if err := h.service.CreateSigner(msgTest); err != nil {
		log.Error("CreateSigner", "err", err)
		return err
	}

	for _, peerId := range h.pm.PeerIDs() {
		wg.Add(1)
		peerAddrTarget := h.pm.Peers()[peerId]
		fmt.Println("send:", peerAddrTarget)
		peerReply, err := SendToPeer(h.pm.Host, PeerArgs{
			peerAddrTarget,
			"PingService",
			"Test",
			PingArgs{
				ID:   h.pm.Host.ID().String(),
				Data: []byte(msgTest),
			},
			peer.ProtocolId,
		}, &wg)

		if err != nil {
			fmt.Println("send err", err)
			return err
		}
		fmt.Println("reply:", peerReply)
	}

	//for _, peerId := range h.pm.PeerIDs() {
	//	wg.Add(1)
	//	peerAddrTarget := h.pm.Peers()[peerId]
	//	fmt.Println("send:", peerAddrTarget)
	//	fmt.Println("send:", peerAddrTarget)
	//	peerReply, err := SendToPeer(h.pm.Host, PeerArgs{
	//		peerAddrTarget,
	//		"PingService",
	//		"Process",
	//		PingArgs{
	//			ID:   h.pm.Host.ID().String(),
	//			Data: []byte("msg"),
	//		},
	//		peer.ProtocolId,
	//	}, &wg)
	//
	//	if err != nil {
	//		fmt.Println("send err", err)
	//		return err
	//	}
	//	fmt.Println("reply:", peerReply)
	//}

	wg.Wait()

	go h.service.Process()
	return nil
}

func (h *RpcService) Merge(_ *http.Request, _ *RpcProcessArgs, reply *RpcReply) error {
	log.Info("RPC server", "test", "called")
	reply.Data = "Test"
	//service, err := NewService(h.config, h.pm, h.badgerFsm)
	//if err != nil {
	//	log.Error("NewService", "err", err)
	//	return err
	//}

	//h.pm.Host.SetStreamHandler(peer.SignerProtocol, func(s network.Stream) {
	//	log.Info("Stream handler", "protocol", s.Protocol(), "peer", s.Conn().LocalPeer())
	//	if service.signer != nil {
	//		service.Handle(s)
	//	}
	//})
	var wg sync.WaitGroup
	msgTest := fmt.Sprintf("%s-%d", "msg-test", 123123)
	if err := h.service.CreateSigner(msgTest); err != nil {
		log.Error("CreateSigner", "err", err)
		return err
	}

	for _, peerId := range h.pm.PeerIDs() {
		wg.Add(1)
		peerAddrTarget := h.pm.Peers()[peerId]
		fmt.Println("send:", peerAddrTarget)
		peerReply, err := SendToPeer(h.pm.Host, PeerArgs{
			peerAddrTarget,
			"PingService",
			"PrepareDataToSign",
			PingArgs{
				ID:   h.pm.Host.ID().String(),
				Data: []byte(msgTest),
			},
			peer.ProtocolId,
		}, &wg)

		if err != nil {
			fmt.Println("send err", err)
			return err
		}
		fmt.Println("reply:", peerReply)
	}

	for _, peerId := range h.pm.PeerIDs() {
		wg.Add(1)
		peerAddrTarget := h.pm.Peers()[peerId]
		fmt.Println("send:", peerAddrTarget)
		peerReply, err := SendToPeer(h.pm.Host, PeerArgs{
			peerAddrTarget,
			"PingService",
			"Process",
			PingArgs{
				ID:   h.pm.Host.ID().String(),
				Data: []byte("msg"),
			},
			peer.ProtocolId,
		}, &wg)

		if err != nil {
			fmt.Println("send err", err)
			return err
		}
		fmt.Println("reply:", peerReply)
	}

	wg.Wait()

	go h.service.Process()
	return nil
}

func (h *RpcService) PrepareDataToSign(_ *http.Request, args *RpcPreDataArgs, reply *RpcReply) error {
	log.Info("RPC server", "PrepareDataToSign", "called")
	if args.Data == "" {
		return errors.New("data required")
	}
	hash := utils.ToHexHash([]byte(args.Data))
	reply.Data = map[string]interface{}{
		"hash": hash,
		"args": args,
	}
	if err := h.service.CreateSigner(args.Data); err != nil {
		log.Error("CreateSigner", "err", err)
		return err
	}
	if err := h.badgerFsm.Set(hash, "1"); err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, peerId := range h.pm.PeerIDs() {
		wg.Add(1)
		peerAddrTarget := h.pm.Peers()[peerId]
		go func() {
			fmt.Println("send:", peerAddrTarget)
			peerReply, err := SendToPeer(h.pm.Host, PeerArgs{
				peerAddrTarget,
				"PingService",
				"PrepareDataToSign",
				PingArgs{
					ID:   h.pm.Host.ID().String(),
					Data: []byte(args.Data),
				},
				peer.ProtocolId,
			}, &wg)

			if err != nil {
				fmt.Println("send err", err)
				return
			}
			fmt.Println("reply:", peerReply)
		}()
	}
	wg.Wait()

	return nil
}

func InitRouter(port int, r *mux.Router, pm *peer.P2PManager, service *Service, c *config.SignerConfig, badgerFsm *peer.BadgerFSM) error {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(rpcjson.NewCodec(), "application/json")

	err := rpcServer.RegisterService(&RpcService{
		pm:        pm,
		service:   service,
		config:    c,
		badgerFsm: badgerFsm,
	}, "signer")
	if err != nil {
		log.Crit("start rpc service failed", "err", err)
	}

	r.Handle("/rpc", rpcServer)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
