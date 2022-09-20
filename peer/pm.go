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
	host     host.Host
	protocol protocol.ID
	peers    map[string]string
}

func NewPeerManager(id string, host host.Host, protocol protocol.ID) *P2PManager {
	log.Info("Peer manager", "id", id, "protocol", protocol)
	return &P2PManager{
		id:       id,
		host:     host,
		protocol: protocol,
		peers:    make(map[string]string),
	}
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

func (p *P2PManager) MustSend(peerID string, message interface{}) {
	log.Info("Pm must send", "peerID", peerID, "message", message)
	err := send(context.Background(), p.host, p.peers[peerID], message, p.protocol)
	if err != nil {
		fmt.Println("MustSend:", err)
		return
	}
}

// EnsureAllConnected connects the host to specified peer and sends the message to it.
func (p *P2PManager) EnsureAllConnected() {
	var wg sync.WaitGroup
	fmt.Println("start EnsureAllConnected")

	for _, peerAddr := range p.peers {
		wg.Add(1)
		go connectToPeer(p.host, peerAddr, &wg)
	}
	wg.Wait()

	fmt.Println("end EnsureAllConnected")
}

// SendAllConnected connects the host to specified peer and sends the message to it.
func (p *P2PManager) SendAllConnected(msg string, id protocol.ID) {
	fmt.Println("start SendAllConnected")

	for _, peerAddr := range p.peers {
		peerAddrTarget := peerAddr
		go func() {
			fmt.Println("send:", peerAddrTarget)
			err := sendMsg(context.Background(), p.host, peerAddrTarget, []byte(msg), id)
			if err != nil {
				fmt.Println("send err", err)
				return
			}
			return
		}()
	}

	fmt.Println("end SendAllConnected")
}

// AddPeers adds peers to peer list.
func (p *P2PManager) AddPeers(peerPorts []int64) error {
	//for _, peerPort := range peerPorts {
	//	peerID := utils.GetPeerIDFromPort(peerPort)
	//	peerAddr, err := getPeerAddr(peerPort)
	//	if err != nil {
	//		log.Warn("Cannot get peer address", "peerPort", peerPort, "peerID", peerID, "err", err)
	//		return err
	//	}
	//	log.Info("peer", "addr", peerAddr)
	//	p.peers[peerID] = peerAddr
	//}
	return nil
}

// AddPeerID adds peerID to peer list.
func (p *P2PManager) AddPeerID(peerID peer.ID, addr string) {
	peerAddr := fmt.Sprintf("%s/p2p/%s", addr, peerID)
	log.Info("Peer added", "addr", peerAddr)
	p.peers[peerID.String()] = peerAddr
	return
}

func connectToPeer(host host.Host, peerAddr string, wg *sync.WaitGroup) {
	defer wg.Done()

	logger := log.New("to", peerAddr)
	for {
		// Connect the host to the peer.
		err := connect(context.Background(), host, peerAddr)
		if err != nil {
			logger.Warn("Failed to connect to peer", "err", err)
			time.Sleep(3 * time.Second)
			continue
		}
		logger.Debug("Successfully connect to peer")
		return
	}
}
