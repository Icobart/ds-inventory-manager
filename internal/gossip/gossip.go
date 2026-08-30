// Package gossip handles peer-to-peer state synchronization in the background.
package gossip

import (
	"database/sql"

	"github.com/Icobart/ds-inventory-manager/internal/storage"
	"github.com/Icobart/ds-inventory-manager/pb"
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
