package server

import (
	"alice-tss/types"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getamis/sirius/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	rpcjson "github.com/gorilla/rpc/v2/json2"
	"net/http"

	"alice-tss/pb"
	"alice-tss/peer"
	"alice-tss/utils"
)

type RpcService struct {
	pm        *peer.P2PManager
	config    *types.AppConfig
	storeDB   types.StoreDB
	tssCaller *TssCaller
}

func (h *RpcService) SignMessage(_ *http.Request, args *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server", "SignMessage", "called", "args", args)

	var dataRequestSign pb.SignRequest
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

	result, err := h.tssCaller.SignMessage(pm, &dataRequestSign, RpcToPeer(pm, "TssPeerService", "SignMessage", argData))
	if err == nil {
		reply.Data = types.RVSignature{
			R:    hex.EncodeToString(result.R.Bytes()),
			S:    hex.EncodeToString(result.S.Bytes()),
			Hash: hash,
		}
	}

	return err
}

func (h *RpcService) RegisterDKG(_ *http.Request, _ *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server", "RegisterDKG", "called")

	hash := utils.RandomHash()
	pm := h.pm.ClonePeerManager(peer.GetProtocol(hash))

	result, err := h.tssCaller.RegisterDKG(pm, hash, RpcToPeer(pm, "TssPeerService", "RegisterDKG", []byte(hash)))
	if err == nil {
		pubkey := crypto.CompressPubkey(result.PublicKey.ToPubKey())
		reply.Data = pb.DkgReply{
			X:       hex.EncodeToString(result.PublicKey.GetX().Bytes()),
			Y:       hex.EncodeToString(result.PublicKey.GetY().Bytes()),
			Address: crypto.PubkeyToAddress(*result.PublicKey.ToPubKey()).String(),
			Pubkey:  hex.EncodeToString(pubkey),
			Hash:    hash,
		}
	}

	return err
}

func (h *RpcService) Reshare(_ *http.Request, args *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server", "Reshare", "called", "args", args)

	var dataShare pb.ReshareRequest
	argData, err := json.Marshal(args.Data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(argData, &dataShare)
	if err != nil {
		return err
	}

	reply.Data = dataShare.Hash
	pm := h.pm.ClonePeerManager(peer.GetProtocol(dataShare.Hash))

	return h.tssCaller.Reshare(pm, &dataShare, RpcToPeer(pm, "TssPeerService", "Reshare", argData))
}

func (h *RpcService) GetDKG(_ *http.Request, args *types.RpcKeyArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server", "GetKey", args)

	data, err := h.tssCaller.StoreDB.GetDKGResultData(args.Key)
	if err != nil {
		return err
	}
	reply.Data = data
	return nil
}

func (h *RpcService) CheckSignature(_ *http.Request, args *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server", "CheckSignature", "called", "args", args)

	var dataSignature pb.CheckSignatureByPubkeyRequest
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

	data, err := h.tssCaller.StoreDB.GetDKGResultData(hash)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	var rvSignature types.RVSignature
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

func InitRouter(port int, r *mux.Router, pm *peer.P2PManager, storeDB types.StoreDB) error {
	log.Info("init router rpc", "port", port)
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(rpcjson.NewCodec(), "application/json")

	err := rpcServer.RegisterService(&RpcService{
		pm:        pm,
		tssCaller: &TssCaller{StoreDB: storeDB},
	}, "signer")
	if err != nil {
		log.Crit("start service service failed", "err", err)
	}

	r.Handle("/tss", rpcServer)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
