package service

import (
	"alice-tss/pb/tss"
	"alice-tss/peer"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"sync"
)

type TssRequest interface {
	tss.SignRequest | tss.ReshareRequest | tss.DKGRequest
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
			fmt.Println("send:", peerAddrTarget)
			peerReply, err := SendToPeer(pm.Host, PeerArgs{
				peerAddrTarget,
				svcName,
				svcMethod,
				PingArgs{
					Data: data,
				},
			}, &wg)

			if err != nil {
				fmt.Println("send err", err)
				return err
			}
			fmt.Println("reply:", peerReply)
		}

		wg.Wait()
		return nil
	}
}
