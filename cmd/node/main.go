package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/Icobart/ds-inventory-manager/internal/storage"
	_ "modernc.org/sqlite"
)

func main() {
	nodeID := flag.String("id", "node-1", "The unique identifier for this node")
	port := flag.Int("port", 50051, "The port the gRPC server will listen on")
	flag.Parse()

	log.Printf("Starting Warehouse Node: %s on port %d", *nodeID, *port)

	dbPath := fmt.Sprintf("%s.db", *nodeID)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := storage.InitDB(db); err != nil {
		log.Fatalf("Failed to initialize database tables: %v", err)
	}
	log.Println("Database initialized successfully.")
}
