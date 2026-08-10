package storage

import (
	"database/sql"
	"testing"

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
