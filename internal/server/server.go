package server

import (
	"context"
	"database/sql"

	"github.com/Icobart/ds-inventory-manager/internal/storage"
	"github.com/Icobart/ds-inventory-manager/internal/vclock"
	"github.com/Icobart/ds-inventory-manager/pb"
)

// NodeServer implements the InventoryManagerServer interface.
type NodeServer struct {
	pb.UnimplementedInventoryManagerServer
	nodeID string
	db     *sql.DB
}

// NewNodeServer creates a new instance of the gRPC server handler.
func NewNodeServer(nodeID string, db *sql.DB) *NodeServer {
	return &NodeServer{
		nodeID: nodeID,
		db:     db,
	}
}

// UpdateLocalStock processes a local inventory change and updates the vector clock.
func (s *NodeServer) UpdateLocalStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	currentQty, err := storage.GetItem(s.db, req.ItemId)
	if err != nil {
		return nil, err
	}
	newQty := currentQty + req.QuantityChange
	err = storage.UpdateItem(s.db, req.ItemId, newQty)
	if err != nil {
		return nil, err
	}
	vc, err := storage.GetVectorClock(s.db)
	if err != nil {
		return nil, err
	}
	vclock.Tick(vc, s.nodeID)
	err = storage.SaveVectorClock(s.db, vc)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateStockResponse{
		Success:      true,
		NewQuantity:  newQty,
		CurrentClock: vc,
	}, nil
}

// SyncLedger handles incoming state synchronization requests from peer nodes.
// It merges the remote vector clock with the local one to maintain causality,
// applies the remote inventory state to the local ledger, and returns the reconciled state.
func (s *NodeServer) SyncLedger(ctx context.Context, req *pb.SyncLedgerRequest) (*pb.SyncLedgerResponse, error) {
	localClock, err := storage.GetVectorClock(s.db)
	if err != nil {
		return nil, err
	}
	mergedClock := vclock.Merge(localClock, req.Clock)
	if err := storage.SaveVectorClock(s.db, mergedClock); err != nil {
		return nil, err
	}
	mergedInventory := make(map[string]int32)
	for itemID, remoteQty := range req.Inventory {
		if err := storage.UpdateItem(s.db, itemID, remoteQty); err != nil {
			return nil, err
		}
		mergedInventory[itemID] = remoteQty
	}
	return &pb.SyncLedgerResponse{
		Success:         true,
		MergedClock:     mergedClock,
		MergedInventory: mergedInventory,
	}, nil
}
