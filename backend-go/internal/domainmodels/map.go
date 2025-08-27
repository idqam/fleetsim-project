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

func (g *Grid) GetAdjacencyData() map[int64][]int64 {
	if g.RoadGraph == nil {
		return make(map[int64][]int64)
	}
	result := make(map[int64][]int64)
	for nodeID, neighbors := range g.RoadGraph.Adjacency {
		neighborsCopy := make([]int64, len(neighbors))
		copy(neighborsCopy, neighbors)
		result[nodeID] = neighborsCopy
	}
	return result
}

func (g *Grid) GetGraphNodes() []GraphNode {
	var nodes []GraphNode
	segmentsSeen := make(map[int64]bool)

	for _, cell := range g.Cells {
		for _, cellRoad := range cell.RoadSegments {
			if segmentsSeen[cellRoad.RoadSegment.ID] {
				continue
			}
			segmentsSeen[cellRoad.RoadSegment.ID] = true

			segment := cellRoad.RoadSegment
			nodes = append(nodes, GraphNode{
				ID:       segment.ID,
				StartX:   segment.StartX,
				StartY:   segment.StartY,
				EndX:     segment.EndX,
				EndY:     segment.EndY,
				Capacity: segment.Capacity,
				IsOpen:   segment.IsOpen,
			})
		}
	}
	return nodes
}

func (g *Grid) GetGraphEdges() []GraphEdge {
	var edges []GraphEdge
	if g.RoadGraph == nil {
		return edges
	}

	for fromID, neighbors := range g.RoadGraph.Adjacency {
		for _, toID := range neighbors {
			if fromID < toID {
				edges = append(edges, GraphEdge{
					From: fromID,
					To:   toID,
				})
			}
		}
	}
	return edges
}

type GraphNode struct {
	ID       int64  `json:"id"`
	StartX   int64  `json:"start_x"`
	StartY   int64  `json:"start_y"`
	EndX     int64  `json:"end_x"`
	EndY     int64  `json:"end_y"`
	Capacity *int64 `json:"capacity,omitempty"`
	IsOpen   bool   `json:"is_open"`
}

type GraphEdge struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}
