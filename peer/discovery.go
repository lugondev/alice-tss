package peer

import (
	"context"
	"fmt"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core"
	"github.com/libp2p/go-libp2p/core/host"
	p2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const (
	SignerProtocol = "/signer/1.0.0"
	DKGProtocol    = "/dkg/1.0.0"
	ProtocolId     = "/p2p/1.0.0"
	// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
	DiscoveryServiceTag = "discovery-wallet-as-a-service"
)

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h  host.Host
	pm *P2PManager
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi p2pPeer.AddrInfo) {
	log.Info("discovered new peer", "id", pi.ID.String(), "addr", pi.Addrs)
	if pi.ID.String() == n.h.ID().String() {
		log.Error("cannot dial to self", "id", pi.ID.String())
	} else {
		err := n.h.Connect(context.Background(), pi)
		if err != nil {
			log.Error("error connecting to peer", "id", pi.ID.String())
		} else {
			n.pm.AddPeerID(pi.ID, pi.Addrs[0].String())
		}
	}
}

// SetupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func SetupDiscovery(h host.Host, pm *P2PManager) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{
		h:  h,
		pm: pm,
	})
	return s.Start()
}

func GetProtocol(id string) core.ProtocolID {
	return core.ProtocolID(fmt.Sprintf("/%s/1.0.0", id))
}
