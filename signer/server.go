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

func (h *RpcService) Process(_ *http.Request, _ *RpcProcessArgs, reply *RpcReply) error {
	log.Info("RPC server", "process", "called")
	reply.Data = "Processed"
	for _, peerId := range h.pm.PeerIDs() {
		peerAddrTarget := h.pm.Peers()[peerId]
		go func() {
			fmt.Println("send:", peerAddrTarget)
			peerReply, err := SentToPeer(h.hostClient, PeerArgs{
				peerAddrTarget,
				"PingService",
				"Process",
				PingArgs{
					ID:   h.hostClient.ID().String(),
					Data: []byte("msg"),
				},
				peer.ProtocolId,
			})

			if err != nil {
				fmt.Println("send err", err)
				return
			}
			fmt.Println("reply:", peerReply)
			h.service.Process()
		}()
	}
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
	if err := h.service.CreateSigner(h.pm, h.config, args.Data); err != nil {
		log.Error("CreateSigner", "err", err)
		return err
	}
	if err := h.badgerFsm.Set(hash, "1"); err != nil {
		return err
	}
	for _, peerId := range h.pm.PeerIDs() {
		peerAddrTarget := h.pm.Peers()[peerId]
		go func() {
			fmt.Println("send:", peerAddrTarget)
			peerReply, err := SentToPeer(h.hostClient, PeerArgs{
				peerAddrTarget,
				"PingService",
				"PrepareDataToSign",
				PingArgs{
					ID:   h.hostClient.ID().String(),
					Data: []byte(args.Data),
				},
				peer.ProtocolId,
			})

			if err != nil {
				fmt.Println("send err", err)
				return
			}
			fmt.Println("reply:", peerReply)
		}()
	}
	return nil
}

func InitRouter(port int, r *mux.Router, pm *peer.P2PManager, service *Service, hostClient host.Host, c *config.SignerConfig, badgerFsm *peer.BadgerFSM) error {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(rpcjson.NewCodec(), "application/json")

	err := rpcServer.RegisterService(&RpcService{
		pm:         pm,
		service:    service,
		hostClient: hostClient,
		config:     c,
		badgerFsm:  badgerFsm,
	}, "signer")
	if err != nil {
		log.Crit("start rpc service failed", "err", err)
	}

	r.Handle("/rpc", rpcServer)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
