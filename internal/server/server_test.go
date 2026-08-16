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

func TestSyncLedger(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	storage.InitDB(db)
	srv := NewNodeServer("node-A", db)

	// Loading some local state (Node A has 100 units, clock is A:1)
	storage.UpdateItem(db, "item-1", 100)
	localClock := &pb.VectorClock{Clocks: map[string]int32{"node-A": 1}}
	storage.SaveVectorClock(db, localClock)

	// Making a remote request (Node B has 90 units, clock is B:1, A:1)
	remoteClock := &pb.VectorClock{Clocks: map[string]int32{"node-A": 1, "node-B": 1}}
	req := &pb.SyncLedgerRequest{
		SourceNodeId: "node-B",
		Clock:        remoteClock,
		Inventory:    map[string]int32{"item-1": 90},
	}

	res, err := srv.SyncLedger(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !res.Success {
		t.Errorf("Expected success to be true")
	}

	// The merged clock should now have A:1 and B:1
	if res.MergedClock.Clocks["node-A"] != 1 || res.MergedClock.Clocks["node-B"] != 1 {
		t.Errorf("Merged clock incorrect: %v", res.MergedClock.Clocks)
	}
	// In a simple state-merge (remote is newer), the inventory should update to 90
	if res.MergedInventory["item-1"] != 90 {
		t.Errorf("Expected merged inventory for item-1 to be 90, got %d", res.MergedInventory["item-1"])
	}
}
