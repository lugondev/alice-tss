package signer

import (
	"alice-tss/config"
	"alice-tss/peer"
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
}

type RpcPreDataArgs struct {
	Data string
}

type RpcProcessArgs struct {
}

type RpcReply struct {
	Message string
}

func (h *RpcService) GetSay(r *http.Request, args *RpcPreDataArgs, reply *RpcReply) error {
	reply.Message = "Rpc, " + args.Data + "!"
	fmt.Printf("request: %v\nargs: %v\nreply: %v\n", r, args, reply)
	return nil
}

func (h *RpcService) Process(_ *http.Request, _ *RpcProcessArgs, reply *RpcReply) error {
	log.Info("RPC server", "process", "called")
	reply.Message = "Processed"
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
	reply.Message = "data prepared: " + args.Data
	if err := h.service.CreateSigner(h.pm, h.config, args.Data); err != nil {
		fmt.Println("CreateSigner err", err)
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

func InitRouter(port int, r *mux.Router, pm *peer.P2PManager, service *Service, hostClient host.Host, c *config.SignerConfig) error {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(rpcjson.NewCodec(), "application/json")

	err := rpcServer.RegisterService(&RpcService{
		pm:         pm,
		service:    service,
		hostClient: hostClient,
		config:     c,
	}, "signer")
	if err != nil {
		log.Crit("start rpc service failed", "err", err)
	}

	r.Handle("/rpc", rpcServer)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
