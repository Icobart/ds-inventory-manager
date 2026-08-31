// Package gossip handles peer-to-peer state synchronization in the background.
package gossip

import (
	"database/sql"

	"github.com/Icobart/ds-inventory-manager/internal/storage"
	"github.com/Icobart/ds-inventory-manager/pb"

	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// BuildSyncPayload extracts the node's entire state to send to a peer.
func BuildSyncPayload(nodeID string, db *sql.DB) (*pb.SyncLedgerRequest, error) {
	clock, err := storage.GetVectorClock(db)
	if err != nil {
		return nil, err
	}
	inventory, err := storage.GetAllItems(db)
	if err != nil {
		return nil, err
	}
	return &pb.SyncLedgerRequest{
		SourceNodeId: nodeID,
		Clock:        clock,
		Inventory:    inventory,
	}, nil
}

// SendGossip opens a connection to a target peer and sends the sync payload.
func SendGossip(target string, req *pb.SyncLedgerRequest) error {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewInventoryManagerClient(conn)
	// Timeout so a dead peer doesn't freeze this node
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = client.SyncLedger(ctx, req)
	return err
}

// StartGossipLoop creates a background Goroutine that periodically syncs state with peers.
func StartGossipLoop(nodeID string, db *sql.DB, peers []string, interval time.Duration) {
	// Create a background thread with the loop
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			req, err := BuildSyncPayload(nodeID, db)
			if err != nil {
				log.Printf("[Gossip] Failed to build payload: %v", err)
				continue
			}
			for _, peer := range peers {
				if err := SendGossip(peer, req); err != nil {
					log.Printf("[Gossip] Peer %s unreachable: %v", peer, err)
				} else {
					log.Printf("[Gossip] Successfully synced state with %s", peer)
				}
			}
		}
	}()
}
