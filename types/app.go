package types

type StoreType string

const (
	StoreTypeMock   StoreType = "mock"
	StoreTypeBadger StoreType = "badger"
)

type StoreConfig struct {
	Type StoreType
	Path string
}

type AppConfig struct {
	Port  int64
	RPC   int
	Store StoreConfig
}
