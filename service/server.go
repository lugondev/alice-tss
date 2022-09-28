package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/getamis/sirius/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	rpcjson "github.com/gorilla/rpc/v2/json2"

	"alice-tss/config"
	"alice-tss/peer"
	"alice-tss/utils"
)

func (h *RpcService) SignMessage(_ *http.Request, args *RpcDataArgs, reply *RpcDataReply) error {
	log.Info("RPC server", "SignMessage", "called", "args", args)
	var dataRequestSign DataRequestSign
	argData, err := json.Marshal(args.Data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(argData, &dataRequestSign)
	if err != nil {
		return err
	}

	signerCfg, err := h.badgerFsm.GetSignerConfig(dataRequestSign.Hash, dataRequestSign.Pubkey)
	if err != nil {
		log.Error("GetSignerConfig", "err", err)
		return err
	}

	hash := utils.ToHexHash([]byte(dataRequestSign.Message))
	reply.Data = hash
	pm := h.pm.ClonePeerManager(peer.GetProtocol(hash))

	service, err := NewSignerService(signerCfg, pm, h.badgerFsm, &pm.Host, dataRequestSign.Message)
	if err != nil {
		log.Error("NewSignerService", "err", err)
		return err
	}

	var wg sync.WaitGroup

	for _, peerId := range pm.PeerIDs() {
		wg.Add(1)
		peerAddrTarget := pm.Peers()[peerId]
		fmt.Println("send:", peerAddrTarget)
		peerReply, err := SendToPeer(pm.Host, PeerArgs{
			peerAddrTarget,
			"TssService",
			"SignMessage",
			PingArgs{
				Data: argData,
			},
		}, &wg)

		if err != nil {
			fmt.Println("send err", err)
			return err
		}
		fmt.Println("reply:", peerReply)
	}

	wg.Wait()

	go service.Process()
	return nil
}

func (h *RpcService) RegisterDKG(_ *http.Request, _ *RpcDataArgs, reply *RpcDataReply) error {
	log.Info("RPC server", "RegisterDKG", "called")

	cfg := &config.DKGConfig{
		Rank:      0,
		Threshold: h.pm.NumPeers(),
	}

	timeNow := time.Now()
	hash := utils.ToHexHash([]byte(timeNow.String()))
	reply.Data = struct {
		Hash   string            `json:"hash"`
		Config *config.DKGConfig `json:"config"`
	}{
		Hash:   hash,
		Config: cfg,
	}

	pm := h.pm.ClonePeerManager(peer.GetProtocol(hash))
	service, err := NewDkgService(cfg, pm, &pm.Host, hash, h.badgerFsm)
	if err != nil {
		log.Error("NewDkgService", "err", err)
		return err
	}
	var wg sync.WaitGroup

	for _, peerId := range h.pm.PeerIDs() {
		wg.Add(1)
		peerAddrTarget := h.pm.Peers()[peerId]
		fmt.Println("send:", peerAddrTarget)
		peerReply, err := SendToPeer(h.pm.Host, PeerArgs{
			peerAddrTarget,
			"TssService",
			"RegisterDKG",
			PingArgs{
				Data: []byte(hash),
			},
		}, &wg)

		if err != nil {
			fmt.Println("send err", err)
			return err
		}
		fmt.Println("reply:", peerReply)
	}

	wg.Wait()

	go service.Process()
	return nil
}

func (h *RpcService) Reshare(_ *http.Request, args *RpcDataArgs, reply *RpcDataReply) error {
	log.Info("RPC server", "Reshare", "called", "args", args)

	var dataShare DataReshare
	argData, err := json.Marshal(args.Data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(argData, &dataShare)
	if err != nil {
		return err
	}
	reply.Data = dataShare.Hash
	signerCfg, err := h.badgerFsm.GetSignerConfig(dataShare.Hash, dataShare.Pubkey)
	if err != nil {
		log.Error("GetSignerConfig", "err", err)
		return err
	}

	pm := h.pm.ClonePeerManager(peer.GetProtocol(dataShare.Hash))
	service, err := NewReshareService(&config.ReshareConfig{
		Threshold: 2,
		Share:     signerCfg.Share,
		Pubkey:    signerCfg.Pubkey,
		BKs:       signerCfg.BKs,
	}, pm, &pm.Host, dataShare.Hash, h.badgerFsm)
	if err != nil {
		log.Error("NewDkgService", "err", err)
		return err
	}
	var wg sync.WaitGroup

	for _, peerId := range h.pm.PeerIDs() {
		wg.Add(1)
		peerAddrTarget := h.pm.Peers()[peerId]
		fmt.Println("send:", peerAddrTarget)
		peerReply, err := SendToPeer(h.pm.Host, PeerArgs{
			peerAddrTarget,
			"TssService",
			"Reshare",
			PingArgs{
				Data: argData,
			},
		}, &wg)

		if err != nil {
			fmt.Println("send err", err)
			return err
		}
		fmt.Println("reply:", peerReply)
	}

	wg.Wait()

	go service.Process()
	return nil
}

func (h *RpcService) GetKey(_ *http.Request, args *RpcKeyArgs, reply *RpcDataReply) error {
	log.Info("RPC server", "GetKey", args)

	data, err := h.badgerFsm.Get(args.Key)
	if err != nil {
		return err
	}
	reply.Data = data
	return nil
}

func (h *RpcService) GetDKG(_ *http.Request, args *RpcKeyArgs, reply *RpcDataReply) error {
	log.Info("RPC server", "GetKey", args)

	data, err := h.badgerFsm.GetDKGResultData(args.Key)
	if err != nil {
		return err
	}
	reply.Data = data
	return nil
}

func InitRouter(port int, r *mux.Router, pm *peer.P2PManager, badgerFsm *peer.BadgerFSM) error {
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(rpcjson.NewCodec(), "application/json")

	err := rpcServer.RegisterService(&RpcService{
		pm:        pm,
		badgerFsm: badgerFsm,
	}, "signer")
	if err != nil {
		log.Crit("start service service failed", "err", err)
	}

	r.Handle("/tss", rpcServer)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}