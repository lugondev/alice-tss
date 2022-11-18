package types

type AppConfig struct {
	Port   int64
	RPC    int
	Badger string
}

type RVSignature struct {
	R    string `json:"r"`
	S    string `json:"s"`
	Hash string `json:"hash"`
}
