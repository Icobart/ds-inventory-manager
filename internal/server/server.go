package server

import (
	"context"
	"database/sql"
	"sync"

	"github.com/Icobart/ds-inventory-manager/internal/storage"
	"github.com/Icobart/ds-inventory-manager/internal/vclock"
	"github.com/Icobart/ds-inventory-manager/pb"
)

// NodeServer implements the InventoryManagerServer interface.
type NodeServer struct {
	pb.UnimplementedInventoryManagerServer
	nodeID string
	db     *sql.DB
	mu     sync.RWMutex
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
	s.mu.Lock()
	defer s.mu.Unlock()
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
func (s *NodeServer) SyncLedger(ctx context.Context, req *pb.SyncLedgerRequest) (*pb.SyncLedgerResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	localClock, err := storage.GetVectorClock(s.db)
	if err != nil {
		return nil, err
	}
	relation := vclock.Compare(localClock, req.Clock)
	mergedClock := vclock.Merge(localClock, req.Clock)
	if err := storage.SaveVectorClock(s.db, mergedClock); err != nil {
		return nil, err
	}
	mergedInventory := make(map[string]int32)
	for itemID, remoteQty := range req.Inventory {
		localQty, err := storage.GetItem(s.db, itemID)
		if err != nil {
			return nil, err
		}
		var finalQty int32
		switch relation {
		case vclock.Before:
			finalQty = remoteQty
		case vclock.Concurrent:
			finalQty = localQty + remoteQty
		default:
			finalQty = localQty
		}
		if err := storage.UpdateItem(s.db, itemID, finalQty); err != nil {
			return nil, err
		}
		mergedInventory[itemID] = finalQty
	}
	return &pb.SyncLedgerResponse{
		Success:         true,
		MergedClock:     mergedClock,
		MergedInventory: mergedInventory,
	}, nil
}

func (s *NodeServer) GetLocalStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	qty, err := storage.GetItem(s.db, req.ItemId)
	if err != nil {
		return nil, err
	}
	vc, err := storage.GetVectorClock(s.db)
	if err != nil {
		return nil, err
	}
	return &pb.GetStockResponse{
		Quantity:     qty,
		CurrentClock: vc,
	}, nil
}

// GetFullInventory retrieves the entire local database state.
func (s *NodeServer) GetFullInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inventory, err := storage.GetAllItems(s.db)
	if err != nil {
		return nil, err
	}
	vc, err := storage.GetVectorClock(s.db)
	if err != nil {
		return nil, err
	}
	return &pb.GetInventoryResponse{
		Items:        inventory,
		CurrentClock: vc,
	}, nil
}
