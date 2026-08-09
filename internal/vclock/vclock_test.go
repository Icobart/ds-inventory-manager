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

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		v1       map[string]int32
		v2       map[string]int32
		expected Relation
	}{
		{
			name:     "Equal clocks",
			v1:       map[string]int32{"A": 1, "B": 1},
			v2:       map[string]int32{"A": 1, "B": 1},
			expected: Equal,
		},
		{
			name:     "v1 Happened Before v2",
			v1:       map[string]int32{"A": 1, "B": 1},
			v2:       map[string]int32{"A": 2, "B": 1}, // A ticked in v2
			expected: Before,
		},
		{
			name:     "v1 Happened After v2",
			v1:       map[string]int32{"A": 2, "B": 1},
			v2:       map[string]int32{"A": 1, "B": 1},
			expected: After,
		},
		{
			name:     "Concurrent",
			v1:       map[string]int32{"A": 2, "B": 1}, // Node A updated
			v2:       map[string]int32{"A": 1, "B": 2}, // Node B updated concurrently
			expected: Concurrent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock1 := &pb.VectorClock{Clocks: tt.v1}
			clock2 := &pb.VectorClock{Clocks: tt.v2}

			result := Compare(clock1, clock2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
