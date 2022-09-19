package peer

import (
	"alice-tss/utils"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/multiformats/go-multiaddr"
	"math/rand"

	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// MakeBasicHost creates a LibP2P host.
func MakeBasicHost(port int64, privateKey *ecdsa.PrivateKey) (host.Host, error) {
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	if err != nil {
		return nil, err
	}

	priv, err := generateIdentity(port)
	if err != nil {
		return nil, err
	}

	privP256, _ := utils.ToEcdsaP256(privateKey.D.Bytes(), false)
	cryptoPriv, _, _ := crypto.ECDSAKeyPairFromKey(privP256)
	pid, err := peer.IDFromPrivateKey(cryptoPriv)
	fmt.Println("pid: ", pid)

	opts := []libp2p.Option{
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	return basicHost, nil
}

// getPeerAddr gets peer full address from port.
func getPeerAddr(port int64) (string, error) {
	priv, err := generateIdentity(port)
	if err != nil {
		return "", err
	}

	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, pid), nil
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

	s, err := host.NewStream(ctx, info.ID, protocol)
	if err != nil {
		log.Warn("Cannot create a new stream", "from", host.ID(), "to", target, "err", err)
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

// send message to specified peer.
func sendMsg(ctx context.Context, host host.Host, target string, msg []byte, protocol protocol.ID) error {
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

	s, err := host.NewStream(ctx, info.ID, protocol)
	if err != nil {
		log.Warn("Cannot create a new stream", "from", host.ID(), "to", target, "err", err)
		return err
	}

	_, err = s.Write(msg)
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
