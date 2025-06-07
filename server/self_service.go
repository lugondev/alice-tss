package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"alice-tss/pb"
	"alice-tss/peer"
	"alice-tss/utils"

	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/crypto/tss/ecdsa/gg18/signer"
	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p/core/host"
	peer2 "github.com/libp2p/go-libp2p/core/peer"
)

// SelfService manages a cluster of 3 TSS nodes for self-contained operations
type SelfService struct {
	hosts   [3]*host.Host
	peerIDs [3]peer2.ID
	mu      sync.RWMutex // Protects concurrent access to hosts and peerIDs
}

const (
	// Port constants for the three self-service nodes
	PortSelf1 = 11111
	PortSelf2 = 11112
	PortSelf3 = 11113

	// Configuration constants
	numNodes          = 3
	peerWaitTimeout   = 30 * time.Second
	peerCheckInterval = 1 * time.Second
)

// RegisterDKG initiates a Distributed Key Generation process across all nodes
func (s *SelfService) RegisterDKG(ctx context.Context, tssCaller *TssCaller, hash string) (*dkg.Result, error) {
	if tssCaller == nil {
		return nil, fmt.Errorf("tssCaller cannot be nil")
	}
	if hash == "" {
		return nil, fmt.Errorf("hash cannot be empty")
	}

	pms, err := s.CreatePm(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer managers: %w", err)
	}

	// Use a WaitGroup to ensure all goroutines complete before function returns
	var wg sync.WaitGroup
	errChan := make(chan error, numNodes-1)

	// Start DKG on nodes 1 and 2 in parallel
	for i := 1; i < numNodes; i++ {
		wg.Add(1)
		go func(nodeIndex int) {
			defer wg.Done()
			nodeID := fmt.Sprintf("%s-%d", hash, nodeIndex)
			if _, err := tssCaller.RegisterDKG(pms[nodeIndex], nodeID, nil); err != nil {
				log.Error("RegisterDKG failed", "node", nodeIndex, "nodeID", nodeID, "error", err)
				select {
				case errChan <- fmt.Errorf("node %d DKG failed: %w", nodeIndex, err):
				default:
				}
			}
		}(i)
	}

	// Start DKG on node 0 (primary node) and wait for result
	result, err := tssCaller.RegisterDKG(pms[0], fmt.Sprintf("%s-%d", hash, 0), func() error {
		// Wait for other nodes to complete with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		}
	})

	return result, err
}

// SignMessage performs threshold signing across all nodes
func (s *SelfService) SignMessage(ctx context.Context, tssCaller *TssCaller, dataRequestSign *pb.SignRequest) (*signer.Result, error) {
	if tssCaller == nil {
		return nil, fmt.Errorf("tssCaller cannot be nil")
	}
	if dataRequestSign == nil {
		return nil, fmt.Errorf("dataRequestSign cannot be nil")
	}
	if dataRequestSign.Message == "" {
		return nil, fmt.Errorf("message cannot be empty")
	}

	hash := utils.ToHexHash([]byte(dataRequestSign.Message))
	pms, err := s.CreatePm(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer managers: %w", err)
	}

	// Use a WaitGroup to ensure all goroutines complete before function returns
	var wg sync.WaitGroup
	errChan := make(chan error, numNodes-1)

	// Start signing on nodes 1 and 2 in parallel
	for i := 1; i < numNodes; i++ {
		wg.Add(1)
		go func(nodeIndex int) {
			defer wg.Done()
			signRequest := &pb.SignRequest{
				Hash:    fmt.Sprintf("%s-%d", dataRequestSign.Hash, nodeIndex),
				Pubkey:  dataRequestSign.Pubkey,
				Message: dataRequestSign.Message,
			}
			if _, err := tssCaller.SignMessage(pms[nodeIndex], signRequest, nil); err != nil {
				log.Error("SignMessage failed", "node", nodeIndex, "hash", signRequest.Hash, "error", err)
				select {
				case errChan <- fmt.Errorf("node %d signing failed: %w", nodeIndex, err):
				default:
				}
			}
		}(i)
	}

	// Start signing on node 0 (primary node) and wait for result
	primarySignRequest := &pb.SignRequest{
		Hash:    fmt.Sprintf("%s-%d", dataRequestSign.Hash, 0),
		Pubkey:  dataRequestSign.Pubkey,
		Message: dataRequestSign.Message,
	}

	result, err := tssCaller.SignMessage(pms[0], primarySignRequest, func() error {
		// Wait for other nodes to complete with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		}
	})

	return result, err
}

// CreatePm creates peer managers for all nodes with proper error handling and timeout
func (s *SelfService) CreatePm(ctx context.Context, protocolID string) ([3]*peer.P2PManager, error) {
	if protocolID == "" {
		return [3]*peer.P2PManager{}, fmt.Errorf("protocolID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var pms [3]*peer.P2PManager
	var setupErrors []error

	// Create peer managers for all nodes
	for i, peerID := range s.peerIDs {
		if s.hosts[i] == nil {
			setupErrors = append(setupErrors, fmt.Errorf("host %d is nil", i))
			continue
		}

		pms[i] = peer.NewPeerManager(peerID.String(), *s.hosts[i], peer.GetProtocol(protocolID))
		if err := peer.SetupDiscovery(pms[i]); err != nil {
			log.Error("SetupDiscovery failed", "index", i, "peerID", peerID, "error", err)
			setupErrors = append(setupErrors, fmt.Errorf("setup discovery for node %d: %w", i, err))
		}
	}

	if len(setupErrors) > 0 {
		return [3]*peer.P2PManager{}, fmt.Errorf("failed to setup peer managers: %v", setupErrors)
	}

	// Wait for peers with timeout and context cancellation
	if err := s.waitForPeers(ctx, pms); err != nil {
		return [3]*peer.P2PManager{}, fmt.Errorf("failed to wait for peers: %w", err)
	}

	return pms, nil
}

// waitForPeers waits for all nodes to discover each other with proper timeout handling
func (s *SelfService) waitForPeers(ctx context.Context, pms [3]*peer.P2PManager) error {
	timeout := time.NewTicker(peerCheckInterval)
	defer timeout.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, peerWaitTimeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for peers to connect: %w", timeoutCtx.Err())
		case <-timeout.C:
			allConnected := true
			for i, pm := range pms {
				if pm == nil || pm.NumPeers() < numNodes-1 {
					connectedPeers := 0
					if pm != nil {
						connectedPeers = int(pm.NumPeers())
					}
					log.Info("Waiting for peers", "node", i, "connected_peers", connectedPeers, "required_peers", numNodes-1)
					allConnected = false
					break
				}
			}
			if allConnected {
				log.Info("All peers connected successfully")
				return nil
			}
		}
	}
}

// NewSelfService creates a new SelfService with proper error handling
func NewSelfService() (*SelfService, error) {
	hosts := [3]*host.Host{}
	peerIDs := [3]peer2.ID{}
	ports := [3]int{PortSelf1, PortSelf2, PortSelf3}

	// Create hosts with proper error handling
	for i, port := range ports {
		host, peerID, err := peer.MakeBasicHostByID(int64(port))
		if err != nil {
			// Clean up any successfully created hosts
			for j := 0; j < i; j++ {
				if hosts[j] != nil {
					if err := (*hosts[j]).Close(); err != nil {
						log.Error("Failed to close host during cleanup", "index", j, "error", err)
					}
				}
			}
			return nil, fmt.Errorf("failed to create host %d on port %d: %w", i, port, err)
		}
		hosts[i] = &host
		peerIDs[i] = peerID
		log.Info("Created TSS node", "index", i, "port", port, "peerID", peerID.String())
	}

	service := &SelfService{
		hosts:   hosts,
		peerIDs: peerIDs,
	}

	log.Info("SelfService initialized successfully", "nodes", numNodes)
	return service, nil
}

// Close gracefully shuts down the SelfService and releases resources
func (s *SelfService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errors []error
	for i, host := range s.hosts {
		if host != nil {
			if err := (*host).Close(); err != nil {
				log.Error("Failed to close host", "index", i, "error", err)
				errors = append(errors, fmt.Errorf("failed to close host %d: %w", i, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	log.Info("SelfService closed successfully")
	return nil
}
