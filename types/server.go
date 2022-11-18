package types

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
