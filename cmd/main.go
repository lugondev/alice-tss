package cmd

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"

	"github.com/dgraph-io/badger"
	"github.com/getamis/sirius/log"
	"github.com/gorilla/mux"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"alice-tss/config"
	"alice-tss/peer"
	"alice-tss/service"
	"alice-tss/utils"
)

var configFile string
var keystoreFile string
var password string
var port int

var Cmd = &cobra.Command{
	Use:   "start",
	Short: "TSS run process with RPC, P2P",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initService(cmd)
		if err != nil {
			log.Crit("Failed to init", "err", err)
		}

		appConfig, err := readAppConfigFile(configFile)
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

		badgerOpt := badger.DefaultOptions(appConfig.BadgerDir)
		badgerDB, err := badger.Open(badgerOpt)
		if err != nil {
			return err
		}

		defer func() {
			if err := badgerDB.Close(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
			}
		}()
		badgerFsm := peer.NewBadger(badgerDB, privateKey)

		// setup local mDNS discovery
		if err := peer.SetupDiscovery(host, pm); err != nil {
			panic(err)
		}

		rpcHost := gorpc.NewServer(host, peer.ProtocolId)
		svc := service.TssService{
			Pm:        pm,
			BadgerFsm: badgerFsm,
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
		if err := service.InitRouter(rpcPort, mux.NewRouter(), pm, badgerFsm); err != nil {
			log.Crit("init router", "err", err)
		}

		return nil
	},
}

func init() {
	Cmd.Flags().String("config", "", "cmd config file path")
	Cmd.Flags().String("keystore", "", "keystore file path")
	Cmd.Flags().String("password", "111111", "password")
	Cmd.Flags().Int("port", 0, "port server")
}

func initService(cmd *cobra.Command) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	configFile = viper.GetString("config")
	keystoreFile = viper.GetString("keystore")
	password = viper.GetString("password")
	port = viper.GetInt("port")

	return nil
}

func readAppConfigFile(filaPath string) (*config.AppConfig, error) {
	c := &config.AppConfig{}
	yamlFile, err := ioutil.ReadFile(filaPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
