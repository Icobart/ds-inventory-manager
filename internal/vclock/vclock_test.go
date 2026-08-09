package vclock

import (
	"testing"

	"github.com/Icobart/ds-inventory-manager/pb"
)

func TestTick(t *testing.T) {
	// Create a new protobuf VectorClock
	clock := &pb.VectorClock{
		Clocks: make(map[string]int32),
	}

	nodeID := "node-A"

	// Tick the clock for the node twice
	Tick(clock, nodeID)
	Tick(clock, nodeID)

	// Node's clock should be exactly 2
	if clock.Clocks[nodeID] != 2 {
		t.Errorf("Expected clock for %s to be 2, got %d", nodeID, clock.Clocks[nodeID])
	}
}
