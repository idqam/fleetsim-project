package domainmodels

type Grid struct {
	DimX  int64  `json:"dimX"`
	DimY  int64  `json:"dimY"`
	Cells []Cell `json:"cells"`

	CoordIndex   map[[2]int64]*Cell `json:"-"` // (x,y) → *Cell
	SegmentIndex map[int64]*Cell    `json:"-"` // segmentID → *Cell
	RoadGraph    *RoadGraph         `json:"-"` // adjacency of road network
}
type CellType string

const (
	CellTypeNormal  CellType = "normal"
	CellTypeRefuel  CellType = "refuel"
	CellTypeDepot   CellType = "depot"
	CellTypeBlocked CellType = "blocked"
)

type Cell struct {
	Xpos         int64      `json:"xpos"`
	Ypos         int64      `json:"ypos"`
	CellType     CellType   `json:"cell_type"`
	RoadSegments []CellRoad `json:"road_segments"`
	RefuelAmount *float64   `json:"refuel_amount,omitempty"`
}
type CellRoad struct {
	RoadSegmentID int64       `json:"road_segment_id"`
	RoadSegment   RoadSegment `json:"road_segment"`
}

type RoadGraph struct {
	Adjacency map[int64][]int64
}
