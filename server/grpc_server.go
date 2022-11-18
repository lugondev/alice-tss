package server

import (
	"alice-tss/pb/tss"
	"alice-tss/peer"
	"alice-tss/utils"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"net"
)

// grpcServer is used to implement proto.GreeterServer.
type grpcServer struct {
	tss.TssServiceServer

	pm        *peer.P2PManager
	badgerFsm *peer.BadgerFSM
	tssCaller *TssCaller
}

func (s *grpcServer) SignMessage(_ context.Context, signRequest *tss.SignRequest) (*tss.RVSignatureReply, error) {
	hash := utils.ToHexHash([]byte(signRequest.Message))
	pm := s.pm.ClonePeerManager(peer.GetProtocol(hash))

	bs, err := proto.Marshal(signRequest)
	if err != nil {
		log.Warn("Cannot proto marshal message", "err", err)
		return nil, err
	}

	result, err := s.tssCaller.SignMessage(pm, signRequest, RpcToPeer(pm, "TssPeerService", "SignMessage", bs))

	if err == nil {
		signature := &tss.RVSignatureReply{
			R:    hex.EncodeToString(result.R.Bytes()),
			S:    hex.EncodeToString(result.S.Bytes()),
			Hash: hash,
		}
		return signature, nil
	}

	return nil, err
}

func (s *grpcServer) RegisterDKG(_ context.Context, _ *tss.DKGRequest) (*tss.DkgReply, error) {
	hash := utils.RandomHash()
	pm := s.pm.ClonePeerManager(peer.GetProtocol(hash))

	result, err := s.tssCaller.RegisterDKG(pm, hash, RpcToPeer(pm, "TssPeerService", "RegisterDKG", []byte(hash)))
	log.Info("RegisterDKG", "result", result, "err", err)
	if err == nil {
		pubkey := crypto.CompressPubkey(result.PublicKey.ToPubKey())
		dkgReply := &tss.DkgReply{
			X:       hex.EncodeToString(result.PublicKey.GetX().Bytes()),
			Y:       hex.EncodeToString(result.PublicKey.GetY().Bytes()),
			Address: crypto.PubkeyToAddress(*result.PublicKey.ToPubKey()).String(),
			Pubkey:  hex.EncodeToString(pubkey),
		}
		return dkgReply, nil
	}

	return nil, err
}

func (s *grpcServer) Reshare(_ context.Context, reshareRequest *tss.ReshareRequest) (*tss.ServiceReply, error) {
	bs, err := proto.Marshal(reshareRequest)
	if err != nil {
		log.Warn("Cannot proto marshal message", "err", err)
		return nil, err
	}
	pm := s.pm.ClonePeerManager(peer.GetProtocol(reshareRequest.Hash))

	if err := s.tssCaller.Reshare(pm, reshareRequest, RpcToPeer(pm, "TssPeerService", "Reshare", bs)); err != nil {
		return nil, err
	}

	return &tss.ServiceReply{}, nil
}

func StartGRPC(port int, pm *peer.P2PManager, badgerFsm *peer.BadgerFSM) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Crit("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	tss.RegisterTssServiceServer(s, &grpcServer{
		pm:        pm,
		badgerFsm: badgerFsm,
		tssCaller: &TssCaller{BadgerFsm: badgerFsm},
	})

	log.Info("server listening", "addr", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Crit("failed to serve: %v", err)
	}
}
