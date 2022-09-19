package dkg

import (
	"alice-tss/peer"
	"alice-tss/utils"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const dkgProtocol = "/dkg/1.0.0"

var configFile string

var Cmd = &cobra.Command{
	Use:   "dkg",
	Short: "DKG process",
	Long:  `Distributed key generation for creating secret shares without any dealer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initService(cmd)
		if err != nil {
			log.Crit("Failed to init", "err", err)
		}

		config, err := readDKGConfigFile(configFile)
		if err != nil {
			log.Crit("Failed to read config file", "configFile", configFile, "err", err)
		}

		// Make a host that listens on the given multiaddress.
		host, err := peer.MakeBasicHost(config.Port)
		if err != nil {
			log.Crit("Failed to create a basic host", "err", err)
		}

		// Create a new peer manager.
		pm := peer.NewPeerManager(utils.GetPeerIDFromPort(config.Port), host, dkgProtocol)
		err = pm.AddPeers(config.Peers)
		if err != nil {
			log.Crit("Failed to add peers", "err", err)
		}

		// Create a new service.
		service, err := NewService(config, pm)
		if err != nil {
			log.Crit("Failed to new service", "err", err)
		}
		// Set a stream handler on the host.
		host.SetStreamHandler(dkgProtocol, func(s network.Stream) {
			service.Handle(s)
		})

		// Ensure all peers are connected before starting DKG process.
		pm.EnsureAllConnected()

		// Start DKG process.
		service.Process()

		return nil
	},
}

func init() {
	Cmd.Flags().String("config", "", "dkg config file path")
}

func initService(cmd *cobra.Command) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	configFile = viper.GetString("config")

	return nil
}
