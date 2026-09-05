package main

import (
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Icobart/ds-inventory-manager/pb"
)

func main() {
	// Parse command-line flags
	target := flag.String("target", "localhost:50051", "The address of the local warehouse node")
	itemID := flag.String("item", "", "The ID of the item to update")
	quantity := flag.Int("add", 0, "The quantity to add (use negative numbers to deduct)")
	get := flag.Bool("get", false, "Fetch the current stock for a specific item")
	all := flag.Bool("all", false, "Fetch the entire local inventory state")
	flag.Parse()
	if *itemID == "" {
		log.Fatal("Error: You must provide an -item ID")
	}
	log.Printf("Connecting to node at %s...", *target)

	// Connect to the gRPC server (insecure credentials for now, since it's local TCP)
	conn, err := grpc.NewClient(*target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to node: %v", err)
	}
	defer conn.Close()
	client := pb.NewInventoryManagerClient(conn)

	// Setup the request and a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.UpdateStockRequest{
		ItemId:         *itemID,
		QuantityChange: int32(*quantity),
	}

	// Get Full Inventory
	if *all {
		res, err := client.GetFullInventory(ctx, &pb.GetInventoryRequest{})
		if err != nil {
			log.Fatalf("RPC call failed: %v", err)
		}
		log.Printf("INVENTORY: %v", res.Items)
		log.Printf("VECTOR CLOCK: %v", res.CurrentClock.Clocks)
		return
	}

	if *itemID == "" {
		log.Fatal("Error: You must provide an -item ID for -add or -get commands")
	}

	// Get Single Item
	if *get {
		res, err := client.GetLocalStock(ctx, &pb.GetStockRequest{ItemId: *itemID})
		if err != nil {
			log.Fatalf("RPC call failed: %v", err)
		}
		log.Printf("STOCK: %s = %d", *itemID, res.Quantity)
		log.Printf("VECTOR CLOCK: %v", res.CurrentClock.Clocks)
		return
	}

	// Call the RPC method on the server
	res, err := client.UpdateLocalStock(ctx, req)
	if err != nil {
		log.Fatalf("RPC call failed: %v", err)
	}

	// Output the result to the user
	if res.Success {
		log.Printf("SUCCESS: Stock updated! New quantity for %s is %d", *itemID, res.NewQuantity)
		log.Printf("Node Vector Clock updated to: %v", res.CurrentClock.Clocks)
	} else {
		log.Println("FAILED to update stock.")
	}
}
