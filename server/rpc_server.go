package server

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"alice-tss/pb"
	"alice-tss/peer"
	"alice-tss/store"
	"alice-tss/types"
	"alice-tss/utils"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getamis/sirius/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	rpcjson "github.com/gorilla/rpc/v2/json2"
)

// unmarshalRequestData is a helper function to unmarshal request data with proper error handling
func unmarshalRequestData(data interface{}, target interface{}) error {
	argData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}
	if err := json.Unmarshal(argData, target); err != nil {
		return fmt.Errorf("failed to unmarshal request data: %w", err)
	}
	return nil
}

type RpcService struct {
	pm          *peer.P2PManager
	config      *types.AppConfig
	storeDB     store.HandlerData
	selfService *SelfService
	tssCaller   *TssCaller
}

// GetSignerConfig retrieves the signer configuration for a given request.
func (h *RpcService) GetSignerConfig(_ *http.Request, args *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server GetSignerConfig called", "args", args)

	var dataRequestSign pb.SignRequest
	if err := unmarshalRequestData(args.Data, &dataRequestSign); err != nil {
		log.Error("Failed to unmarshal sign request", "error", err)
		return err
	}

	result, err := h.tssCaller.GetSignerConfig(&dataRequestSign)
	if err != nil {
		log.Error("Failed to get signer config", "error", err)
		return err
	}

	reply.Data = result
	return nil
}

// SignMessage performs threshold signature generation for a given message.
func (h *RpcService) SignMessage(_ *http.Request, args *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server SignMessage called", "args", args)

	var dataRequestSign pb.SignRequest
	if err := unmarshalRequestData(args.Data, &dataRequestSign); err != nil {
		log.Error("Failed to unmarshal sign request", "error", err)
		return err
	}

	hash := utils.ToHexHash([]byte(dataRequestSign.Message))
	pm := h.pm.ClonePeerManager(peer.GetProtocol(hash))

	argData, err := json.Marshal(args.Data)
	if err != nil {
		log.Error("Failed to marshal args data", "error", err)
		return err
	}

	result, err := h.tssCaller.SignMessage(pm, &dataRequestSign, RpcToPeer(pm, "TssPeerService", "SignMessage", argData))
	if err != nil {
		log.Error("Failed to sign message", "error", err)
		return err
	}

	reply.Data = types.RVSignature{
		R:    hex.EncodeToString(result.R.Bytes()),
		S:    hex.EncodeToString(result.S.Bytes()),
		Hash: hash,
	}

	return nil
}

// SelfSignMessage performs threshold signature generation using the self-service cluster.
func (h *RpcService) SelfSignMessage(_ *http.Request, args *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server SelfSignMessage called", "args", args)
	if h.selfService == nil {
		return errors.New("self service is not available")
	}

	var dataRequestSign pb.SignRequest
	if err := unmarshalRequestData(args.Data, &dataRequestSign); err != nil {
		log.Error("Failed to unmarshal sign request", "error", err)
		return err
	}

	ctx := context.Background()
	result, err := h.selfService.SignMessage(ctx, h.tssCaller, &dataRequestSign)
	if err != nil {
		log.Error("Failed to sign message with self service", "error", err)
		return err
	}

	reply.Data = types.RVSignature{
		R:    hex.EncodeToString(result.R.Bytes()),
		S:    hex.EncodeToString(result.S.Bytes()),
		Hash: utils.ToHexHash([]byte(dataRequestSign.Message)),
	}

	return nil
}

// RegisterDKG initiates a Distributed Key Generation process across connected peers.
func (h *RpcService) RegisterDKG(_ *http.Request, _ *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server RegisterDKG called")

	hash := utils.RandomHash()
	pm := h.pm.ClonePeerManager(peer.GetProtocol(hash))

	result, err := h.tssCaller.RegisterDKG(pm, hash, RpcToPeer(pm, "TssPeerService", "RegisterDKG", []byte(hash)))
	if err != nil {
		log.Error("Failed to register DKG", "error", err)
		return err
	}

	pubkey := crypto.CompressPubkey(result.PublicKey.ToPubKey())
	reply.Data = pb.DkgReply{
		X:       hex.EncodeToString(result.PublicKey.GetX().Bytes()),
		Y:       hex.EncodeToString(result.PublicKey.GetY().Bytes()),
		Address: crypto.PubkeyToAddress(*result.PublicKey.ToPubKey()).String(),
		Pubkey:  hex.EncodeToString(pubkey),
		Hash:    hash,
	}
	return nil
}

func (h *RpcService) RegisterSelfDKG(_ *http.Request, _ *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server", "RegisterSelfDKG", "called", "port", h.config.Port)
	if h.selfService == nil {
		return errors.New("self service is not available")
	}

	hash := utils.RandomHash()

	ctx := context.Background()
	dkgResult, err := h.selfService.RegisterDKG(ctx, h.tssCaller, hash)
	if err != nil {
		return err
	}

	pubkey := crypto.CompressPubkey(dkgResult.PublicKey.ToPubKey())
	reply.Data = pb.DkgReply{
		X:       hex.EncodeToString(dkgResult.PublicKey.GetX().Bytes()),
		Y:       hex.EncodeToString(dkgResult.PublicKey.GetY().Bytes()),
		Address: crypto.PubkeyToAddress(*dkgResult.PublicKey.ToPubKey()).String(),
		Pubkey:  hex.EncodeToString(pubkey),
		Hash:    hash,
	}
	return nil
}

// Reshare initiates a key resharing process to refresh the threshold shares.
func (h *RpcService) Reshare(_ *http.Request, args *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server Reshare called", "args", args)

	var dataShare pb.ReshareRequest
	if err := unmarshalRequestData(args.Data, &dataShare); err != nil {
		log.Error("Failed to unmarshal reshare request", "error", err)
		return err
	}

	reply.Data = dataShare.Hash
	pm := h.pm.ClonePeerManager(peer.GetProtocol(dataShare.Hash))

	argData, err := json.Marshal(args.Data)
	if err != nil {
		log.Error("Failed to marshal args data", "error", err)
		return err
	}

	if err := h.tssCaller.Reshare(pm, &dataShare, RpcToPeer(pm, "TssPeerService", "Reshare", argData)); err != nil {
		log.Error("Failed to reshare", "error", err)
		return err
	}

	return nil
}

// GetDKG retrieves DKG result data by hash key.
func (h *RpcService) GetDKG(_ *http.Request, args *types.RpcKeyArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server GetDKG called", "key", args.Key)

	data, err := h.tssCaller.StoreDB.GetDKGResultData(args.Key)
	if err != nil {
		log.Error("Failed to get DKG result data", "key", args.Key, "error", err)
		return err
	}
	reply.Data = data
	return nil
}

// CheckSignature verifies an ECDSA signature against a message and public key.
func (h *RpcService) CheckSignature(_ *http.Request, args *types.RpcDataArgs, reply *types.RpcDataReply) error {
	log.Info("RPC server CheckSignature called", "args", args)

	var dataSignature pb.CheckSignatureByPubkeyRequest
	if err := unmarshalRequestData(args.Data, &dataSignature); err != nil {
		log.Error("Failed to unmarshal signature check request", "error", err)
		return err
	}

	hash := utils.ToHexHash([]byte(dataSignature.Message))
	log.Info("CheckSignature", "hash", hash)

	data, err := h.tssCaller.StoreDB.GetDKGResultData(hash)
	if err != nil {
		log.Error("Failed to get DKG result data", "hash", hash, "error", err)
		return err
	}

	var rvSignature types.RVSignature
	rvsData, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal signature data", "error", err)
		return err
	}
	if err := json.Unmarshal(rvsData, &rvSignature); err != nil {
		log.Error("Failed to unmarshal signature data", "error", err)
		return err
	}

	checkedSignature, err := utils.CheckSignatureECDSA(dataSignature.Message, rvSignature, dataSignature.Pubkey)
	if err != nil {
		log.Error("Failed to check signature", "error", err)
		return err
	}

	reply.Data = checkedSignature
	return nil
}

// InitRouter initializes and starts the HTTP RPC server with timeout middleware.
// It registers the RPC service and starts listening on the configured port.
func InitRouter(config *types.AppConfig, pm *peer.P2PManager, storeDB store.HandlerData, selfService *SelfService) error {
	log.Info("init router rpc", "port", config.RPC)
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(rpcjson.NewCodec(), "application/json")

	err := rpcServer.RegisterService(&RpcService{
		pm:          pm,
		config:      config,
		selfService: selfService,
		tssCaller:   &TssCaller{StoreDB: storeDB},
	}, "signer")
	if err != nil {
		log.Crit("start service service failed", "err", err)
	}

	r := mux.NewRouter()
	r.Handle("/tss", rpcServer)

	muxWithMiddlewares := http.TimeoutHandler(r, time.Second*5, "Timeout!")
	return http.ListenAndServe(fmt.Sprintf(":%d", config.RPC), muxWithMiddlewares)
}
