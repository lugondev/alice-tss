package service

import (
	"alice-tss/config"
	"alice-tss/peer"
	"github.com/libp2p/go-libp2p/core/host"
)

type RpcService struct {
	pm         *peer.P2PManager
	hostClient host.Host
	config     *config.SignerConfig
	badgerFsm  *peer.BadgerFSM
}

type RpcDataArgs struct {
	Data interface{}
}

type DataRequestSign struct {
	Hash    string `json:"hash"`
	Pubkey  string `json:"pubkey"`
	Message string `json:"message"`
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
