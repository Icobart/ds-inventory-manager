package storage

import (
	"database/sql"
	"testing"

	"github.com/Icobart/ds-inventory-manager/pb"
	_ "modernc.org/sqlite"
)

func TestInventoryOperations(t *testing.T) {
	// Create an in-memory database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close() // to ensure the database is closed after the test
	err = InitDB(db)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Update an item's quantity
	itemID := "item-A"
	expectedQty := int32(150)
	err = UpdateItem(db, itemID, expectedQty)
	if err != nil {
		t.Fatalf("Failed to update item: %v", err)
	}

	// Retrieve the item and check the quantity
	qty, err := GetItem(db, itemID)
	if err != nil {
		t.Fatalf("Failed to get item: %v", err)
	}
	if qty != expectedQty {
		t.Errorf("Expected quantity %d, got %d", expectedQty, qty)
	}
}

func TestVectorClockOperations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	_ = InitDB(db)

	// Create a mock clock
	clock := &pb.VectorClock{
		Clocks: map[string]int32{"node-A": 5, "node-B": 3},
	}

	// Save the clock to SQLite
	err = SaveVectorClock(db, clock)
	if err != nil {
		t.Fatalf("Failed to save clock: %v", err)
	}

	// Retrieve the clock and verify its contents
	retrieved, err := GetVectorClock(db)
	if err != nil {
		t.Fatalf("Failed to get clock: %v", err)
	}
	if retrieved.Clocks["node-A"] != 5 || retrieved.Clocks["node-B"] != 3 {
		t.Errorf("Retrieved clock does not match saved clock. Got: %v", retrieved.Clocks)
	}
}
