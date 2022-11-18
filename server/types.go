package server

import (
	"alice-tss/config"
	"alice-tss/peer"
)

type RpcService struct {
	pm        *peer.P2PManager
	config    *config.AppConfig
	badgerFsm *peer.BadgerFSM
	tssCaller *TssCaller
}

type RpcDataArgs struct {
	Data interface{}
}

type RpcKeyArgs struct {
	Key string `json:"key"`
}

type RpcNoneArgs struct {
}

type RpcDataReply struct {
	Data interface{}
}

type RpcMessageReply struct {
	Message string
}
