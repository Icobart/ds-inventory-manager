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
