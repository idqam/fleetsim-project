package gridengine

import (
	"encoding/binary"
	"math"
	"math/rand"

	"owenvi.com/simsim/internal/coremodels"
)

type BaseParams struct {
	BoxWidth   float64
	BoxHeight  float64
	CenterX    float64
	CenterY    float64
	JitterMax  float64
}

type NodeSegmentCounters struct {
	NextNode int64
	NextSeg  int64
}

type SpatialGrid struct {
	cells    [][][]int
	cellSize float64
	rows     int
	cols     int
	minX     float64
	minY     float64
}

type Point struct {
	X, Y  float64
	Index int
	Alive bool
}

func NewSpatialGrid(points []Point, cellSize float64, bounds BaseParams) *SpatialGrid {
	cols := int(bounds.BoxWidth/cellSize) + 1
	rows := int(bounds.BoxHeight/cellSize) + 1
	
	grid := &SpatialGrid{
		cells:    make([][][]int, rows),
		cellSize: cellSize,
		rows:     rows,
		cols:     cols,
		minX:     bounds.CenterX - bounds.BoxWidth/2,
		minY:     bounds.CenterY - bounds.BoxHeight/2,
	}
	
	for i := range grid.cells {
		grid.cells[i] = make([][]int, cols)
	}
	
	grid.UpdatePoints(points)
	return grid
}

func (sg *SpatialGrid) UpdatePoints(points []Point) {
	for i := range sg.cells {
		for j := range sg.cells[i] {
			sg.cells[i][j] = sg.cells[i][j][:0]
		}
	}
	
	for _, p := range points {
		if !p.Alive {
			continue
		}
		gx := int((p.X - sg.minX) / sg.cellSize)
		gy := int((p.Y - sg.minY) / sg.cellSize)
		if gx >= 0 && gx < sg.cols && gy >= 0 && gy < sg.rows {
			sg.cells[gy][gx] = append(sg.cells[gy][gx], p.Index)
		}
	}
}

func (sg *SpatialGrid) QueryRadius(x, y, radius float64) []int {
	minGx := max(0, int((x-radius-sg.minX)/sg.cellSize))
	maxGx := min(sg.cols-1, int((x+radius-sg.minX)/sg.cellSize))
	minGy := max(0, int((y-radius-sg.minY)/sg.cellSize))
	maxGy := min(sg.rows-1, int((y+radius-sg.minY)/sg.cellSize))
	
	var result []int
	for gy := minGy; gy <= maxGy; gy++ {
		for gx := minGx; gx <= maxGx; gx++ {
			result = append(result, sg.cells[gy][gx]...)
		}
	}
	return result
}

func ReverseClosestLookup(sources []Point, targets []*coremodels.Node, maxDist float64) map[int]int64 {
	result := make(map[int]int64)
	maxDist2 := maxDist * maxDist
	
	for _, src := range sources {
		if !src.Alive {
			continue
		}
		
		closest := int64(-1)
		best2 := maxDist2
		
		for nodeID, target := range targets {
			dx := src.X - target.Pos_X
			dy := src.Y - target.Pos_Y
			d2 := dx*dx + dy*dy
			
			if d2 < best2 {
				best2 = d2
				closest = int64(nodeID)
			}
		}
		
		if closest != -1 {
			result[src.Index] = closest
		}
	}
	return result
}

func AddNodeWithCounter(g *coremodels.Grid, x, y float64, counter *NodeSegmentCounters) int64 {
	nodeID := counter.NextNode
	g.Nodes[nodeID] = &coremodels.Node{
		ID:    nodeID,
		Pos_X: x,
		Pos_Y: y,
	}
	counter.NextNode++
	return nodeID
}

func AddSegmentWithCounter(g *coremodels.Grid, from, to int64, congestion float64, counter *NodeSegmentCounters) {
	n1 := g.Nodes[from]
	n2 := g.Nodes[to]
	lengthKm := distanceKm(n1.Pos_X, n1.Pos_Y, n2.Pos_X, n2.Pos_Y)

	segID := counter.NextSeg
	g.Segments[segID] = &coremodels.RoadSegment{
		ID:               segID,
		StartNode:        from,
		EndNode:          to,
		LengthKM:         lengthKm,
		CongestionFactor: congestion,
	}

	g.Adjacency[from] = append(g.Adjacency[from], segID)
	g.Adjacency[to] = append(g.Adjacency[to], segID)
	counter.NextSeg++
}

func AddNodesGrid(g *coremodels.Grid, rows, cols int64, cellSize, jitters float64, r *rand.Rand, counter *NodeSegmentCounters) [][]int64 {
	nodeGrid := make([][]int64, rows)
	for i := range nodeGrid {
		nodeGrid[i] = make([]int64, cols)
	}
	
	for y := int64(0); y < rows; y++ {
		for x := int64(0); x < cols; x++ {
			px := jitter(r, float64(x)*cellSize, jitters)
			py := jitter(r, float64(y)*cellSize, jitters)
			nodeGrid[y][x] = AddNodeWithCounter(g, px, py, counter)
		}
	}
	return nodeGrid
}

func AddNodesRadial(g *coremodels.Grid, centerX, centerY float64, numRays, numRings int, ringSpacing, rayJitter float64, r *rand.Rand, counter *NodeSegmentCounters) (int64, [][]int64) {
	centerIdx := AddNodeWithCounter(g, centerX, centerY, counter)
	
	ringNodes := make([][]int64, numRings)
	
	for i := 0; i < numRays; i++ {
		theta := (2 * math.Pi * float64(i)) / float64(numRays)
		for ring := 1; ring <= numRings; ring++ {
			radius := float64(ring) * ringSpacing
			x := centerX + math.Cos(theta)*radius
			y := centerY + math.Sin(theta)*radius
			x = jitter(r, x, rayJitter)
			y = jitter(r, y, rayJitter)
			idx := AddNodeWithCounter(g, x, y, counter)
			ringNodes[ring-1] = append(ringNodes[ring-1], idx)
		}
	}
	return centerIdx, ringNodes
}

func NewRandFromSeed(seed coremodels.GridConfig) *rand.Rand {
	var s uint64
	idBytes := seed.Seed.Bytes()
	if len(idBytes) >= 8 {
		s = binary.BigEndian.Uint64(idBytes[len(idBytes)-8:])
	}
	return rand.New(rand.NewSource(int64(s)))
}

func addNode(g *coremodels.Grid, x, y float64, nextNode *int64) int64 {
	nodeID := *nextNode
	g.Nodes[nodeID] = &coremodels.Node{
		ID:    nodeID,
		Pos_X: x,
		Pos_Y: y,
	}
	*nextNode++
	return nodeID
}

func addSegment(g *coremodels.Grid, id int64, from, to int64, congestion float64) {
	n1 := g.Nodes[from]
	n2 := g.Nodes[to]
	lengthKm := distanceKm(n1.Pos_X, n1.Pos_Y, n2.Pos_X, n2.Pos_Y)

	g.Segments[id] = &coremodels.RoadSegment{
		ID:               id,
		StartNode:        from,
		EndNode:          to,
		LengthKM:         lengthKm,
		CongestionFactor: congestion,
	}

	g.Adjacency[from] = append(g.Adjacency[from], id)
	g.Adjacency[to] = append(g.Adjacency[to], id)
}

func distanceKm(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	meters := math.Hypot(dx, dy)
	return meters / 1000.0
}

func jitter(r *rand.Rand, val, maxJitter float64) float64 {
	if maxJitter <= 0 {
		return val
	}
	return val + (r.Float64()*2-1)*maxJitter
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b { return a }
	return b
}

func min(a, b int) int {
	if a < b { return a }
	return b
}

func isBoundary(x, y, maxX, maxY int64) bool {
	return x == 0 || y == 0 || x == maxX || y == maxY
}

func NeighborNodes(g *coremodels.Grid, nodeID int64) []int64 {
	segIDs := g.Adjacency[nodeID]
	if len(segIDs) == 0 {
		return nil
	}
	neighbors := make([]int64, 0, len(segIDs))
	for _, segID := range segIDs {
		seg := g.Segments[segID]
		if seg.StartNode == nodeID {
			neighbors = append(neighbors, seg.EndNode)
		} else {
			neighbors = append(neighbors, seg.StartNode)
		}
	}
	return neighbors
}

func NeighborSegments(g *coremodels.Grid, nodeID int64) []*coremodels.RoadSegment {
	segIDs := g.Adjacency[nodeID]
	if len(segIDs) == 0 {
		return nil
	}
	segments := make([]*coremodels.RoadSegment, 0, len(segIDs))
	for _, segID := range segIDs {
		segments = append(segments, g.Segments[segID])
	}
	return segments
}

func OtherNode(seg *coremodels.RoadSegment, nodeID int64) int64 {
	if seg.StartNode == nodeID {
		return seg.EndNode
	}
	return seg.StartNode
}

func NodeDegree(g *coremodels.Grid, nodeID int64) int {
	return len(g.Adjacency[nodeID])
}

func HasDirectConnection(g *coremodels.Grid, n1, n2 int64) bool {
	for _, segID := range g.Adjacency[n1] {
		seg := g.Segments[segID]
		if (seg.StartNode == n1 && seg.EndNode == n2) ||
			(seg.StartNode == n2 && seg.EndNode == n1) {
			return true
		}
	}
	return false
}

func RandomNeighbor(g *coremodels.Grid, nodeID int64, randFn func(int) int) (int64, bool) {
	segIDs := g.Adjacency[nodeID]
	if len(segIDs) == 0 {
		return 0, false
	}
	idx := randFn(len(segIDs))
	seg := g.Segments[segIDs[idx]]
	return OtherNode(seg, nodeID), true
}