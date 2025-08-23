package domainmodels

type Grid struct {
	DimX  int64  `json:"dimX"`
	DimY  int64  `json:"dimY"`
	Cells []Cell `json:"cells"`
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

type RoadSegment struct {
	ID     int64 `json:"id"`
	StartX int64 `json:"start_x"`
	StartY int64 `json:"start_y"`
	EndX   int64 `json:"end_x"`
	EndY   int64 `json:"end_y"`

	// Optional attributes
	SpeedLimit        *int64   `json:"speed_limit,omitempty"`
	Capacity          *int64   `json:"capacity,omitempty"`
	WeatherConditions []string `json:"weather_conditions,omitempty"`
	IsOpen            bool     `json:"is_open"`
}
type RoadGraph struct {
	Adjacency map[int64][]int64
}
