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

func TestMerge(t *testing.T) {
	// Create two divergent clocks
	clock1 := &pb.VectorClock{
		Clocks: map[string]int32{"A": 2, "B": 1, "C": 5},
	}
	clock2 := &pb.VectorClock{
		Clocks: map[string]int32{"A": 1, "B": 3, "D": 4},
	}
	// Merge the clocks
	merged := Merge(clock1, clock2)
	// The result should contain the max value from both clocks
	expected := map[string]int32{"A": 2, "B": 3, "C": 5, "D": 4}

	if len(merged.Clocks) != len(expected) {
		t.Fatalf("Expected %d nodes in merged clock, got %d", len(expected), len(merged.Clocks))
	}

	for node, expectedVal := range expected {
		if merged.Clocks[node] != expectedVal {
			t.Errorf("Expected node %s to be %d, got %d", node, expectedVal, merged.Clocks[node])
		}
	}
}
