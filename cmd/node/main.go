package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/Icobart/ds-inventory-manager/internal/server"
	"github.com/Icobart/ds-inventory-manager/internal/storage"
	"github.com/Icobart/ds-inventory-manager/pb"
	_ "modernc.org/sqlite"
)

func main() {
	nodeID := flag.String("id", "node-1", "The unique identifier for this node")
	port := flag.Int("port", 50051, "The port the gRPC server will listen on")
	flag.Parse()

	log.Printf("Starting Warehouse Node: %s on port %d", *nodeID, *port)

	// Starting the SQLite database
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

	// Starting the TCP listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", *port, err)
	}

	// Initializing and registering the gRPC server
	grpcServer := grpc.NewServer()
	nodeServer := server.NewNodeServer(*nodeID, db)
	pb.RegisterInventoryManagerServer(grpcServer, nodeServer)
	log.Printf("gRPC server actively listening at %v", lis.Addr())

	// Start Serving to block the main thread, avoiding the program exiting
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
