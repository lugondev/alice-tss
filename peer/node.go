package peer

import (
	"alice-tss/utils"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"math/rand"
	"time"
)

const (
	timeRetryCreateStream  = 10
	delayRetryCreateStream = 500 * time.Millisecond
)

// MakeBasicHost creates a LibP2P host.
func MakeBasicHost(port int64, privateKey *ecdsa.PrivateKey) (host.Host, peer.ID, error) {
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	if err != nil {
		return nil, "", err
	}

	privP256, _ := utils.ToEcdsaP256(privateKey.D.Bytes(), false)
	cryptoPriv, _, _ := crypto.ECDSAKeyPairFromKey(privP256)
	pid, err := peer.IDFromPrivateKey(cryptoPriv)
	if err != nil {
		return nil, "", err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(cryptoPriv),
	}

	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, "", err
	}

	return basicHost, pid, nil
}

// MakeBasicHostByID creates a LibP2P host.
func MakeBasicHostByID(port int64) (host.Host, peer.ID, error) {
	log.Info("MakeBasicHostByID", "port", port)
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	if err != nil {
		return nil, "", err
	}

	cryptoPriv, _ := generateIdentity(port)
	pid, err := peer.IDFromPrivateKey(cryptoPriv)
	if err != nil {
		return nil, "", err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(cryptoPriv),
	}

	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, "", err
	}

	return basicHost, pid, nil
}

// generateIdentity generates a fixed key pair by using port as random source.
func generateIdentity(port int64) (crypto.PrivKey, error) {
	// Use the port as the randomness source in this example.
	// #nosec: G404: Use of weak random number generator (math/rand instead of crypto/rand)
	r := rand.New(rand.NewSource(port))

	// Generate a key pair for this host.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.ECDSA, 2048, r)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

// send the proto message to specified peer.
func send(ctx context.Context, host host.Host, target string, data interface{}, protocol protocol.ID) error {
	isProtoMsg := true
	msg, ok := data.(proto.Message)
	if !ok {
		log.Warn("invalid proto message")
		//return errors.New("invalid proto message")
		isProtoMsg = false
	}
	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(target)
	if err != nil {
		log.Warn("Cannot parse the target address", "target", target, "err", err)
		return err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Warn("Cannot parse addr", "addr", maddr, "err", err)
		return err
	}

	var s network.Stream
	for i := 0; i < timeRetryCreateStream; i++ {
		log.Info("NewStream", "id", info.ID, "protocol", protocol, "addr", info.Addrs, "info", info.String())

		s, err = host.NewStream(ctx, info.ID, protocol)
		if err != nil {
			log.Warn("Try create a new stream", "after", fmt.Sprintf("%d miliseconds", delayRetryCreateStream), "to", target, "err", err)
			time.Sleep(delayRetryCreateStream)
		}
	}
	if s == nil {
		log.Error("Cannot create a new stream", "from", host.ID(), "to", target, "err", err)
		return err
	}

	var bs []byte
	if isProtoMsg {
		bs, err = proto.Marshal(msg)
		if err != nil {
			log.Warn("Cannot proto marshal message", "err", err)
			return err
		}
	} else {
		bs, err = json.Marshal(data)
		if err != nil {
			log.Warn("Cannot json marshal message", "err", err)
			return err
		}
	}

	_, err = s.Write(bs)
	if err != nil {
		log.Warn("Cannot write message to IO", "err", err)
		return err
	}
	err = s.Close()
	if err != nil {
		log.Warn("Cannot close the stream", "err", err)
		return err
	}

	log.Info("Sent message", "peer", target)
	return nil
}

// connect the host to the specified peer.
func connect(ctx context.Context, host host.Host, target string) error {
	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(target)
	if err != nil {
		log.Warn("Cannot parse the target address", "target", target, "err", err)
		return err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Error("Cannot parse addr", "addr", maddr, "err", err)
		return err
	}

	// Connect the host to the peer.
	err = host.Connect(ctx, *info)
	if err != nil {
		log.Warn("Failed to connect to peer", "err", err)
		return err
	}
	return nil
}
