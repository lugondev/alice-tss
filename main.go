package main

import (
	"alice-tss/server"
	"alice-tss/store"
	"alice-tss/types"
	"flag"

	"github.com/getamis/sirius/log"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/spf13/viper"

	"alice-tss/peer"
	"alice-tss/utils"
)

var configFile string
var keystoreFile string
var password string
var port int
var selfHost bool

// main initializes and starts the TSS (Threshold Signature Scheme) service.
// It sets up peer-to-peer networking, storage, and RPC servers for distributed
// cryptographic operations including DKG, signing, and key resharing.
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

	storeDb, err := store.NewStoreHandler(appConfig.Store, privateKey)
	if err != nil {
		log.Crit("Failed to create a store handler", "err", err)
	}
	defer storeDb.Defer()

	var selfService *server.SelfService = nil
	if selfHost {
		selfService, err = server.NewSelfService()
		if err != nil {
			log.Crit("Failed to create self service", "err", err)
		}
		defer func() {
			if selfService != nil {
				if closeErr := selfService.Close(); closeErr != nil {
					log.Error("Failed to close self service", "err", closeErr)
				}
			}
		}()
	} else {
		// setup local mDNS discovery
		if err := peer.SetupDiscovery(pm); err != nil {
			log.Crit("Failed to setup discovery", "err", err)
		}
	}

	rpcServer := server.NewRpcServer(pm, storeDb)
	rpcHost := gorpc.NewServer(host, peer.ProtocolId)

	if err := rpcHost.Register(rpcServer); err != nil {
		log.Crit("Failed to register rpc server", "err", err)
	}

	if port != 0 {
		appConfig.RPC = port
	}

	//go server.StartGRPC(rpcPort+1000, pm, storeDb)
	if err := server.InitRouter(appConfig, pm, storeDb, selfService); err != nil {
		log.Crit("init router", "err", err)
	}
}

func init() {
	flag.StringVar(&configFile, "config", "", "config file name")
	flag.StringVar(&keystoreFile, "keystore", "", "keystore file path")
	flag.StringVar(&password, "password", "", "password for keystore file")
	flag.IntVar(&port, "port", 0, "port server")
	flag.BoolVar(&selfHost, "self-host", false, "run self host")
}

// readAppConfigFile reads and parses the application configuration file.
// It returns the parsed configuration or an error if the file cannot be read or parsed.
func readAppConfigFile() (*types.AppConfig, error) {
	viper.SetConfigFile(configFile)
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()
	if err != nil {
		log.Error("Cannot read configuration file", "error", err)
		return nil, err
	}
	var c types.AppConfig
	if err := viper.Unmarshal(&c); err != nil {
		log.Error("Cannot unmarshal configuration", "error", err)
		return nil, err
	}
	log.Info("config", "config", c)

	return &c, nil
}
