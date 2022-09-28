package peer

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/peer"
	"sync"
	"time"

	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type P2PManager struct {
	id       string
	Host     host.Host
	protocol protocol.ID
	peers    map[string]string
}

func NewPeerManager(id string, host host.Host, protocol protocol.ID) *P2PManager {
	log.Info("P2PManager", "id", id, "protocol", protocol)

	return &P2PManager{
		id:       id,
		Host:     host,
		protocol: protocol,
		peers:    make(map[string]string),
	}
}

func (p *P2PManager) ClonePeerManager(protocol protocol.ID) *P2PManager {
	pm := *p
	pm.SetProtocol(protocol)

	return &pm
}

func (p *P2PManager) NumPeers() uint32 {
	return uint32(len(p.peers))
}

func (p *P2PManager) SelfID() string {
	return p.id
}

func (p *P2PManager) PeerIDs() []string {
	ids := make([]string, len(p.peers))
	i := 0
	for id := range p.peers {
		fmt.Println("p.peers:", id)
		ids[i] = id
		i++
	}
	return ids
}

func (p *P2PManager) Peers() map[string]string {
	return p.peers
}

func (p *P2PManager) SetProtocol(id protocol.ID) {
	p.protocol = id
}

func (p *P2PManager) MustSend(peerID string, message interface{}) {
	log.Info("P2PManager MustSend", "peerID", peerID, "protocol", p.protocol)
	err := send(context.Background(), p.Host, p.peers[peerID], message, p.protocol)
	if err != nil {
		log.Error("MustSend", "err", err, "protocol", p.protocol)
		return
	}
}

// EnsureAllConnected connects the host to specified peer and sends the message to it.
func (p *P2PManager) EnsureAllConnected() {
	log.Info("P2PManager", "call", "EnsureAllConnected")
	var wg sync.WaitGroup
	for _, peerAddr := range p.peers {
		wg.Add(1)
		go connectToPeer(p.Host, peerAddr, &wg)
	}
	wg.Wait()
}

// AddPeerID adds peerID to peer list.
func (p *P2PManager) AddPeerID(peerID peer.ID, addr string) {
	peerAddr := fmt.Sprintf("%s/p2p/%s", addr, peerID)
	log.Info("P2PManager", "action", "peer added", "addr", peerAddr)
	p.peers[peerID.String()] = peerAddr
	log.Info("P2PManager", "num peers", p.NumPeers())
	return
}

func connectToPeer(host host.Host, peerAddr string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// Connect the host to the peer.
		err := connect(context.Background(), host, peerAddr)
		if err != nil {
			log.Warn("Failed to connect to peer", "to", peerAddr, "err", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Debug("Successfully connect to peer")
		return
	}
}
