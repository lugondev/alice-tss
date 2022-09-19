package signer

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/getamis/sirius/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"alice-tss/peer"
	"alice-tss/utils"
)

var configFile string
var keystoreFile string
var password string

var Cmd = &cobra.Command{
	Use:   "signer",
	Short: "Signer process",
	Long:  `Signing for using the secret shares to generate a signature.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := initService(cmd)
		if err != nil {
			log.Crit("Failed to init", "err", err)
		}

		c, err := readSignerConfigFile(configFile)
		if err != nil {
			log.Crit("Failed to read config file", "configFile", configFile, "err", err)
		}

		privateKey, err := utils.GetPrivateKeyFromKeystore(keystoreFile, password)
		if err != nil {
			log.Crit("GetPrivateKeyFromKeystore", "err", err)
		}

		// Make a host that listens on the given multiaddress.
		host, err := peer.MakeBasicHost(c.Port, privateKey)
		if err != nil {
			log.Crit("Failed to create a basic host", "err", err)
		}

		// Create a new peer manager.
		pm := peer.NewPeerManager(utils.GetPeerIDFromPort(c.Port), host, peer.SignerProtocol)
		err = pm.AddPeers(c.Peers)
		if err != nil {
			log.Crit("Failed to add peers", "err", err)
		}

		badgerOpt := badger.DefaultOptions(c.BadgerDir)
		badgerDB, err := badger.Open(badgerOpt)
		if err != nil {
			panic(err)
		}

		defer func() {
			if err := badgerDB.Close(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
			}
		}()
		badgerFsm := peer.NewBadger(badgerDB)

		// setup local mDNS discovery
		if err := peer.SetupDiscovery(host); err != nil {
			panic(err)
		}

		e := echo.New()
		e.HideBanner = true
		e.HidePort = true
		e.Pre(middleware.RemoveTrailingSlash())

		service, err := NewService(c, pm, badgerFsm)

		rpcHost := gorpc.NewServer(host, peer.ProtocolId)
		svc := PingService{
			service: service,
			pm:      pm,
			config:  c,
		}

		if err := rpcHost.Register(&svc); err != nil {
			panic(err)
		}
		if err != nil {
			log.Crit("Failed to new service", "err", err)
		}

		// Set a stream handler on the host.
		host.SetStreamHandler(peer.SignerProtocol, func(s network.Stream) {
			log.Info("Stream", "protocol", s.Protocol(), "peer", s.Conn().LocalPeer())
			service.Handle(s)
		})

		e.POST("/ping", func(eCtx echo.Context) error {
			for _, peerId := range pm.PeerIDs() {
				peerAddrTarget := pm.Peers()[peerId]
				go func() {
					fmt.Println("send:", peerAddrTarget)
					reply, err := SentToPeer(host, PeerArgs{
						peerAddrTarget,
						"PingService",
						"Ping",
						PingArgs{
							ID:   host.ID().String(),
							Data: []byte("msg"),
						},
						peer.ProtocolId,
					})

					if err != nil {
						fmt.Println("send err", err)
						return
					}
					fmt.Println("reply:", reply)
					service.Process()
				}()
			}

			return eCtx.JSON(http.StatusOK, map[string]interface{}{
				"message": "Ping",
				"data":    "",
			})
		})

		e.POST("/prepare", func(eCtx echo.Context) error {
			msg := eCtx.QueryParam("msg")
			if msg == "" {
				msg = "default_message"
			}
			if err := service.CreateSigner(pm, c, msg); err != nil {
				fmt.Println("CreateSigner err", err)
				return eCtx.JSON(http.StatusBadRequest, map[string]interface{}{
					"message": "CreateSigner err",
					"data":    err.Error(),
				})
			}
			for _, peerId := range pm.PeerIDs() {
				peerAddrTarget := pm.Peers()[peerId]
				go func() {
					fmt.Println("send:", peerAddrTarget)
					reply, err := SentToPeer(host, PeerArgs{
						peerAddrTarget,
						"PingService",
						"PrepareMsg",
						PingArgs{
							ID:   host.ID().String(),
							Data: []byte(msg),
						},
						peer.ProtocolId,
					})

					if err != nil {
						fmt.Println("send err", err)
						return
					}
					fmt.Println("reply:", reply)
				}()
			}

			return eCtx.JSON(http.StatusOK, map[string]interface{}{
				"message": "Ping",
				"data":    "",
			})
		})

		if err := e.StartServer(&http.Server{
			Addr:         ":" + strconv.FormatInt(c.Port-1000, 10),
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		}); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	Cmd.Flags().String("config", "", "signer config file path")
	Cmd.Flags().String("keystore", "", "keystore file path")
	Cmd.Flags().String("password", "111111", "password")
}

func initService(cmd *cobra.Command) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	configFile = viper.GetString("config")
	keystoreFile = viper.GetString("keystore")
	password = viper.GetString("password")

	return nil
}
