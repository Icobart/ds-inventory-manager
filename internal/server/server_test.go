package server

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Icobart/ds-inventory-manager/internal/storage"
	"github.com/Icobart/ds-inventory-manager/pb"
	_ "modernc.org/sqlite"
)

func TestUpdateLocalStock(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	storage.InitDB(db)
	nodeID := "node-A"
	srv := NewNodeServer(nodeID, db)
	// Adding 50 units  of item-1 to inventory
	req := &pb.UpdateStockRequest{
		ItemId:         "item-1",
		QuantityChange: 50,
	}

	res, err := srv.UpdateLocalStock(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !res.Success {
		t.Errorf("Expected success to be true")
	}
	if res.NewQuantity != 50 {
		t.Errorf("Expected new quantity to be 50, got %d", res.NewQuantity)
	}

	// Verify the vector clock ticked and was returned
	if res.CurrentClock == nil || res.CurrentClock.Clocks[nodeID] != 1 {
		t.Errorf("Expected clock for %s to tick to 1", nodeID)
	}
	// Verify the storage layer was actually updated
	savedQty, _ := storage.GetItem(db, "item-1")
	if savedQty != 50 {
		t.Errorf("Expected database quantity to be 50, got %d", savedQty)
	}
}
