package vclock

import (
	"github.com/Icobart/ds-inventory-manager/pb"
)

func Tick(vc *pb.VectorClock, nodeID string) {
	if vc.Clocks == nil {
		vc.Clocks = make(map[string]int32)
	}
	vc.Clocks[nodeID]++
}
