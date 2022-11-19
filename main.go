package main

import (
	"alice-tss/server"
	"alice-tss/store"
	"alice-tss/types"
	"flag"
	"github.com/getamis/sirius/log"
	"github.com/gorilla/mux"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/spf13/viper"

	"alice-tss/peer"
	"alice-tss/utils"
)

var configFile string
var keystoreFile string
var password string
var port int

func main() {
	flag.Parse()

	log.Info("load config file", "configFile", configFile)
	appConfig, err := readAppConfigFile()
	if err != nil {
		log.Crit("Failed to read config file", "configFile", configFile, "err", err)
	}

	privateKey, err := utils.GetPrivateKeyFromKeystore(keystoreFile, password)
	if err != nil {
		log.Crit("GetPrivateKeyFromKeystore", "err", err)
	}

	// Make a host that listens on the given multiaddress.
	host, pid, err := peer.MakeBasicHost(appConfig.Port, privateKey)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}

	log.Info("peer host", "pid", pid)

	// Create a new peer manager.
	pm := peer.NewPeerManager(pid.String(), host, peer.ProtocolId)
	if err != nil {
		log.Crit("Failed to add peers", "err", err)
	}

	//log.Info("badger dir", "dir", appConfig.Badger)
	//badgerOpt := badger.DefaultOptions(appConfig.Badger)
	//badgerDB, err := badger.Open(badgerOpt)
	//if err != nil {
	//	panic(err)
	//}
	//
	//defer func() {
	//	if err := badgerDB.Close(); err != nil {
	//		_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
	//	}
	//}()
	//badgerFsm := badger2.NewBadger(badgerDB, privateKey)
	//storeDb := badger2.NewBadgerDB(badgerFsm)
	storeDb := store.NewMockDB()

	// setup local mDNS discovery
	if err := peer.SetupDiscovery(host, pm); err != nil {
		panic(err)
	}

	rpcHost := gorpc.NewServer(host, peer.ProtocolId)
	svc := server.TssPeerService{
		Pm:        pm,
		TssCaller: &server.TssCaller{StoreDB: storeDb},
	}

	if err := rpcHost.Register(&svc); err != nil {
		panic(err)
	}
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}

	rpcPort := appConfig.RPC
	if port != 0 {
		rpcPort = port
	}

	go server.StartGRPC(rpcPort+1000, pm, storeDb)
	if err := server.InitRouter(rpcPort, mux.NewRouter(), pm, storeDb); err != nil {
		log.Crit("init router", "err", err)
	}
}

func init() {
	flag.StringVar(&configFile, "config", "", "config file name")
	flag.StringVar(&keystoreFile, "keystore", "", "keystore file path")
	flag.StringVar(&password, "password", "111111", "password")
	flag.IntVar(&port, "port", 0, "port server")
}

func readAppConfigFile() (*types.AppConfig, error) {
	viper.SetConfigFile(configFile)
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()
	if err != nil {
		log.Error("Cannot read configuration file", err)
		panic(err)
	}
	var c types.AppConfig
	if err := viper.Unmarshal(&c); err != nil {
		panic(err)
	}
	log.Info("config", "config", c)

	return &c, nil
}
