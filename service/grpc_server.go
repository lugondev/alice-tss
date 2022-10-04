package service

import (
	"alice-tss/pb/tss"
	"alice-tss/peer"
	"alice-tss/utils"
	"context"
	"fmt"
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

func (s *grpcServer) SignMessage(_ context.Context, signRequest *tss.SignRequest) (*tss.ServiceReply, error) {
	//signerCfg, err := s.badgerFsm.GetSignerConfig(signRequest.Hash, signRequest.Pubkey)
	//if err != nil {
	//	log.Error("GetSignerConfig", "err", err)
	//	return nil, err
	//}

	hash := utils.ToHexHash([]byte(signRequest.Message))
	pm := s.pm.ClonePeerManager(peer.GetProtocol(hash))

	bs, err := proto.Marshal(signRequest)
	if err != nil {
		log.Warn("Cannot proto marshal message", "err", err)
		return nil, err
	}
	if err := s.tssCaller.SignMessage(pm, signRequest, RpcToPeer(pm, "TssService", "SignMessage", bs)); err != nil {
		return nil, err
	}

	return &tss.ServiceReply{
		Data: nil,
	}, nil
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
