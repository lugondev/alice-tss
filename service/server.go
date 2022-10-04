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
	"alice-tss/pb/tss"
	"alice-tss/peer"
	"alice-tss/utils"
)

func (h *RpcService) SignMessage(_ *http.Request, args *RpcDataArgs, reply *RpcDataReply) error {
	log.Info("RPC server", "SignMessage", "called", "args", args)

	var dataRequestSign tss.SignRequest
	argData, err := json.Marshal(args.Data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(argData, &dataRequestSign)
	if err != nil {
		return err
	}

	hash := utils.ToHexHash([]byte(dataRequestSign.Message))
	pm := h.pm.ClonePeerManager(peer.GetProtocol(hash))
	reply.Data = hash

	return h.tssCaller.SignMessage(pm, &dataRequestSign, RpcToPeer(pm, "TssService", "SignMessage", argData))
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

	var dataShare tss.ReshareRequest
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

func (h *RpcService) CheckSignature(_ *http.Request, args *RpcDataArgs, reply *RpcDataReply) error {
	log.Info("RPC server", "CheckSignature", "called", "args", args)

	var dataSignature tss.CheckSignatureByPubkeyRequest
	argData, err := json.Marshal(args.Data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(argData, &dataSignature)
	if err != nil {
		return err
	}

	hash := utils.ToHexHash([]byte(dataSignature.Message))
	log.Info("CheckSignature", "hash", hash)

	data, err := h.badgerFsm.Get(hash)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	var rvSignature config.RVSignature
	rvsData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(rvsData, &rvSignature)
	if err != nil {
		return err
	}
	checkedSignature, err := utils.CheckSignatureECDSA(dataSignature.Message, rvSignature, dataSignature.Pubkey)
	if err != nil {
		return err
	}
	reply.Data = checkedSignature
	return nil
}

func InitRouter(port int, r *mux.Router, pm *peer.P2PManager, badgerFsm *peer.BadgerFSM) error {
	log.Info("init router rpc", "port", port)
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(rpcjson.NewCodec(), "application/json")

	err := rpcServer.RegisterService(&RpcService{
		pm:        pm,
		badgerFsm: badgerFsm,
		tssCaller: &TssCaller{BadgerFsm: badgerFsm},
	}, "signer")
	if err != nil {
		log.Crit("start service service failed", "err", err)
	}

	r.Handle("/tss", rpcServer)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
