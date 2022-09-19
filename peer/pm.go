package peer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"

	"alice-tss/utils"
)

type PeerManager struct {
	id       string
	host     host.Host
	protocol protocol.ID
	peers    map[string]string
}

func NewPeerManager(id string, host host.Host, protocol protocol.ID) *PeerManager {
	return &PeerManager{
		id:       id,
		host:     host,
		protocol: protocol,
		peers:    make(map[string]string),
	}
}

func (p *PeerManager) NumPeers() uint32 {
	return uint32(len(p.peers))
}

func (p *PeerManager) SelfID() string {
	return p.id
}

func (p *PeerManager) PeerIDs() []string {
	ids := make([]string, len(p.peers))
	i := 0
	for id := range p.peers {
		ids[i] = id
		i++
	}
	return ids
}

func (p *PeerManager) Peers() map[string]string {
	return p.peers
}

func (p *PeerManager) MustSend(peerID string, message interface{}) {
	err := send(context.Background(), p.host, p.peers[peerID], message, p.protocol)
	if err != nil {
		fmt.Println("MustSend:", err)
		return
	}
}

// EnsureAllConnected connects the host to specified peer and sends the message to it.
func (p *PeerManager) EnsureAllConnected() {
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
func (p *PeerManager) SendAllConnected(msg string, id protocol.ID) {
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
func (p *PeerManager) AddPeers(peerPorts []int64) error {
	for _, peerPort := range peerPorts {
		peerID := utils.GetPeerIDFromPort(peerPort)
		peerAddr, err := getPeerAddr(peerPort)
		if err != nil {
			log.Warn("Cannot get peer address", "peerPort", peerPort, "peerID", peerID, "err", err)
			return err
		}
		p.peers[peerID] = peerAddr
	}
	return nil
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
