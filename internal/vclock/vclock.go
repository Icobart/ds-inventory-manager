package vclock

import (
	"maps"

	"github.com/Icobart/ds-inventory-manager/pb"
)

// Relation defines the causal relationship between two vector clocks.
type Relation string

const (
	Equal      Relation = "Equal"
	Before     Relation = "Before"
	After      Relation = "After"
	Concurrent Relation = "Concurrent"
)

// Tick increments the logical clock for a specific node.
// It initializes the map if it doesn't exist.
func Tick(vc *pb.VectorClock, nodeID string) {
	if vc.Clocks == nil {
		vc.Clocks = make(map[string]int32)
	}
	vc.Clocks[nodeID]++
}

// Compare determines the causal relationship between two vector clocks.
// It evaluates the counters across all known nodes to determine if an event happened
// Before, After or Concurrently with another event.
func Compare(v1, v2 *pb.VectorClock) Relation {
	v1IsLessOrEqual := true
	v2IsLessOrEqual := true
	m1 := v1.GetClocks()
	if m1 == nil {
		m1 = make(map[string]int32)
	}
	m2 := v2.GetClocks()
	if m2 == nil {
		m2 = make(map[string]int32)
	}
	allNodes := make(map[string]bool)
	for node := range m1 {
		allNodes[node] = true
	}
	for node := range m2 {
		allNodes[node] = true
	}
	for node := range allNodes {
		val1 := m1[node]
		val2 := m2[node]
		if val1 > val2 {
			v1IsLessOrEqual = false
		}
		if val1 < val2 {
			v2IsLessOrEqual = false
		}
	}
	if v1IsLessOrEqual && v2IsLessOrEqual {
		return Equal
	} else if v1IsLessOrEqual && !v2IsLessOrEqual {
		return Before
	} else if !v1IsLessOrEqual && v2IsLessOrEqual {
		return After
	}
	return Concurrent
}

// Merge combines two vector clocks by taking the pairwise maximum of all node counters.
// It returns a newly allocated VectorClock to avoid changing the originals.
func Merge(v1, v2 *pb.VectorClock) *pb.VectorClock {
	merged := &pb.VectorClock{
		Clocks: make(map[string]int32),
	}

	getMap := func(vc *pb.VectorClock) map[string]int32 {
		if vc != nil && vc.Clocks != nil {
			return vc.Clocks
		}
		return make(map[string]int32)
	}

	m1 := getMap(v1)
	m2 := getMap(v2)
	maps.Copy(merged.Clocks, m1)
	for node, val2 := range m2 {
		if val1, exists := merged.Clocks[node]; !exists || val2 > val1 {
			merged.Clocks[node] = val2
		}
	}
	return merged
}
