package runtime

import "owenvi.com/fleetsim/internal/domainmodels"

type RoadGraph struct {
	Adjacency map[int64][]int64
}

type CellIndex struct {
	ByCoord map[[2]int64]*domainmodels.Cell
}

type RoadSegmentIndex struct {
	ByID map[int64]*domainmodels.RoadSegment
}

type CellSegmentIndex struct {
	ByCell map[[2]int64][]*domainmodels.RoadSegment
}

type MoveRequest struct {
	VehicleID string
	FromCell  *domainmodels.Cell
	ToCell    *domainmodels.Cell
	SegmentID int64
	SimTime   int64
}
