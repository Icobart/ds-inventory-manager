package storage

import (
	"database/sql"

	"github.com/Icobart/ds-inventory-manager/pb"
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

// SaveVectorClock overwrites the current vector clock state in the database.
func SaveVectorClock(db *sql.DB, vc *pb.VectorClock) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	// Clear the old clock state to prepare for the new one
	_, err = tx.Exec(`DELETE FROM vector_clocks`)
	if err != nil {
		tx.Rollback()
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO vector_clocks (node_id, counter) VALUES (?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	// Insert every node's counter from the map
	for nodeID, counter := range vc.GetClocks() {
		_, err = stmt.Exec(nodeID, counter)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetVectorClock retrieves the entire vector clock map for this node from SQLite.
func GetVectorClock(db *sql.DB) (*pb.VectorClock, error) {
	query := `SELECT node_id, counter FROM vector_clocks`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	vc := &pb.VectorClock{
		Clocks: make(map[string]int32),
	}
	// Iterate through all rows and reconstruct the map
	for rows.Next() {
		var nodeID string
		var counter int32
		if err := rows.Scan(&nodeID, &counter); err != nil {
			return nil, err
		}
		vc.Clocks[nodeID] = counter
	}
	return vc, nil
}

// GetAllItems returns the entire inventory state as a map.
func GetAllItems(db *sql.DB) (map[string]int32, error) {
	query := `SELECT item_id, quantity FROM inventory`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	inventory := make(map[string]int32)
	for rows.Next() {
		var itemID string
		var qty int32
		if err := rows.Scan(&itemID, &qty); err != nil {
			return nil, err
		}
		inventory[itemID] = qty
	}
	return inventory, nil
}
