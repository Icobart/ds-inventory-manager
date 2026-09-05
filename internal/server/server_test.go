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

func TestSyncLedger_ConcurrentMerge(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	storage.InitDB(db)
	srv := NewNodeServer("node-A", db)

	// Simulate Local State: Node A processed 50 units locally while isolated
	storage.UpdateItem(db, "tablet", 50)
	localClock := &pb.VectorClock{Clocks: map[string]int32{"node-A": 1}}
	storage.SaveVectorClock(db, localClock)

	// Simulate Remote State: Node B processed 100 units concurrently
	// Because neither clock knows about the other's tick
	remoteClock := &pb.VectorClock{Clocks: map[string]int32{"node-B": 1}}
	req := &pb.SyncLedgerRequest{
		SourceNodeId: "node-B",
		Clock:        remoteClock,
		Inventory:    map[string]int32{"tablet": 100},
	}

	res, err := srv.SyncLedger(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// The inventory must safely add the concurrent changes (50 + 100 = 150)
	if res.MergedInventory["tablet"] != 150 {
		t.Errorf("Expected merged inventory for tablet to be 150, got %d", res.MergedInventory["tablet"])
	}
}

func TestGetLocalStock(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	storage.InitDB(db)
	srv := NewNodeServer("node-A", db)

	// Seed the database with known state
	storage.UpdateItem(db, "monitor", 45)
	localClock := &pb.VectorClock{Clocks: map[string]int32{"node-A": 2}}
	storage.SaveVectorClock(db, localClock)

	req := &pb.GetStockRequest{ItemId: "monitor"}
	res, err := srv.GetLocalStock(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Verify the correct quantity is returned
	if res.Quantity != 45 {
		t.Errorf("Expected quantity 45, got %d", res.Quantity)
	}
	// Verify the clock did NOT tick
	if res.CurrentClock.Clocks["node-A"] != 2 {
		t.Errorf("Expected clock to remain 2, got %d", res.CurrentClock.Clocks["node-A"])
	}
}

func TestGetFullInventory(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	storage.InitDB(db)
	srv := NewNodeServer("node-A", db)

	// Seed multiple items
	storage.UpdateItem(db, "monitor", 45)
	storage.UpdateItem(db, "keyboard", 100)
	localClock := &pb.VectorClock{Clocks: map[string]int32{"node-A": 2}}
	storage.SaveVectorClock(db, localClock)

	req := &pb.GetInventoryRequest{}
	res, err := srv.GetFullInventory(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Verify all items are present and accurate
	if len(res.Items) != 2 {
		t.Errorf("Expected 2 items in inventory, got %d", len(res.Items))
	}
	if res.Items["keyboard"] != 100 {
		t.Errorf("Expected keyboard quantity 100, got %d", res.Items["keyboard"])
	}
	// Verify the clock did not tick
	if res.CurrentClock.Clocks["node-A"] != 2 {
		t.Errorf("Expected clock to remain 2, got %d", res.CurrentClock.Clocks["node-A"])
	}
}
