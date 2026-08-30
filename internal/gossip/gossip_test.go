package gossip

import (
	"database/sql"
	"testing"

	"github.com/Icobart/ds-inventory-manager/internal/storage"
	"github.com/Icobart/ds-inventory-manager/pb"
	_ "modernc.org/sqlite"
)

func TestBuildSyncPayload(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	storage.InitDB(db)

	nodeID := "node-A"
	storage.UpdateItem(db, "item-1", 100)
	storage.UpdateItem(db, "item-2", 50)
	clock := &pb.VectorClock{Clocks: map[string]int32{"node-A": 5, "node-B": 2}}
	storage.SaveVectorClock(db, clock)

	// Create the payload
	req, err := BuildSyncPayload(nodeID, db)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.SourceNodeId != nodeID {
		t.Errorf("Expected SourceNodeId to be %s, got %s", nodeID, req.SourceNodeId)
	}
	if req.Inventory["item-1"] != 100 || req.Inventory["item-2"] != 50 {
		t.Errorf("Inventory payload incorrect: %v", req.Inventory)
	}
	if req.Clock.Clocks["node-A"] != 5 {
		t.Errorf("Clock payload incorrect: %v", req.Clock.Clocks)
	}
}
