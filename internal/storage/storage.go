package storage

import (
	"database/sql"
)

// InitDB creates the necessary tables for the inventory and vector clocks if they don't exist.
func InitDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS inventory (
		item_id TEXT PRIMARY KEY,
		quantity INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS vector_clocks (
		node_id TEXT PRIMARY KEY,
		counter INTEGER NOT NULL
	);
	`
	_, err := db.Exec(query)
	return err
}

// UpdateItem sets the quantity of a specific item in the local ledger.
// If the item does not exist, it is created. If it exists, it is overwritten.
func UpdateItem(db *sql.DB, itemID string, quantity int32) error {
	query := `
	INSERT INTO inventory (item_id, quantity) 
	VALUES (?, ?)
	ON CONFLICT(item_id) DO UPDATE SET quantity = excluded.quantity;
	`
	_, err := db.Exec(query, itemID, quantity)
	return err
}

// GetItem retrieves the quantity of a specific item.
// It returns 0 if the item is not found.
func GetItem(db *sql.DB, itemID string) (int32, error) {
	query := `SELECT quantity FROM inventory WHERE item_id = ?`
	var qty int32
	err := db.QueryRow(query, itemID).Scan(&qty)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return qty, err
}
