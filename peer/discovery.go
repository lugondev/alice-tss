package peer

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	p2pPeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const (
	SignerProtocol = "/signer/1.0.0"
	ProtocolId     = "/p2p/1.0.0"
	// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
	DiscoveryServiceTag = "pub-sub-discovery-bam-bridge"
)

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi p2pPeer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.Pretty())
	//if !funk.ContainsString(p2p.Peers, pi.ID.String()) {
	//	fmt.Printf("not in whitelist %s\n", pi.ID.Pretty())
	//} else
	if pi.ID.String() == n.h.ID().String() {
		fmt.Println("cannot dial to self")
	} else {
		err := n.h.Connect(context.Background(), pi)
		if err != nil {
			fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
		}
	}
}

// SetupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func SetupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}
