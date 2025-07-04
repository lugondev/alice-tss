package server

import (
	"alice-tss/pb"
	"alice-tss/peer"
	"encoding/json"
	"errors"
	"sync"

	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
)

type TssRequest interface {
	pb.SignRequest | pb.ReshareRequest | pb.DKGRequest
}

func UnmarshalRequest[T TssRequest](data []byte, request *T) error {
	err := json.Unmarshal(data, request)
	if err != nil {
		log.Warn("invalid json message")
		errProto := proto.Unmarshal(data, any(request).(proto.Message))
		if errProto != nil {
			log.Warn("invalid proto message")
			return errors.New("invalid message, cannot unmarshal")
		} else {
			err = nil
		}
	}

	if err != nil {
		return errors.New("invalid message, cannot unmarshal")
	}
	return nil
}

func RpcToPeer(pm *peer.P2PManager, svcName, svcMethod string, data []byte) func() error {
	return func() error {
		var wg sync.WaitGroup

		for _, peerId := range pm.PeerIDs() {
			wg.Add(1)
			peerAddrTarget := pm.Peers()[peerId]
			log.Debug("Sending message to peer", "target", peerAddrTarget)
			peerReply, err := SendToPeer(pm.Host, PeerArgs{
				peerAddrTarget,
				svcName,
				svcMethod,
				PingArgs{
					Data: data,
				},
			}, &wg)

			if err != nil {
				log.Error("Failed to send message to peer", "error", err)
				return err
			}
			log.Debug("Received reply from peer", "reply", peerReply)
		}

		wg.Wait()
		return nil
	}
}
