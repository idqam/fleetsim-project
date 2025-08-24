package domainmodels

import (
	"encoding/json"
	"fmt"

	"owenvi.com/fleetsim/internal/utils"
)

type GridLoader struct {
	Width  int64 `json:"grid_width"`
	Height int64 `json:"grid_height"`
	Seed   int64 `json:"seed"`

	//0.0 t0 1.0
	RefuelCellsAllotment  float64
	DepotCellsAllotment   float64
	BlockedCellsAllotment float64
	//0.0 to 1.0
	RoadDensity  float64
	MainRoadBias float64
	DeadEndBias  float64

	SegmentIDCounter  int64
	GenerationStatsSu *GenerationStats
}

type GenerationStats struct {
	TotalCells          int
	RoadCells           int
	SpecialCells        int
	TotalSegments       int
	MainArteries        int
	SecondaryRoads      int
	DeadEnds            int
	ConnectedComponents int
	GenerationTimeMs    int64
}

var GridLoaderDemo = GridLoader{
	Width:                 20,
	Height:                20,
	Seed:                  42,
	RefuelCellsAllotment:  0.05,
	DepotCellsAllotment:   0.02,
	BlockedCellsAllotment: 0.05,
	RoadDensity:           0.7,
	MainRoadBias:          0.3,
	DeadEndBias:           0.1,
	SegmentIDCounter:      1,
}

func (gl *GridLoader) ConfigureForTesting(
	width int64,
	height int64,
	seed int64,
	refuelCellsAllotment float64,
	depotCellsAllotment float64,
	blockedCellsAllotment float64,
	roadDensity float64,
	mainRoadBias float64,
	deadEndBias float64,
) {
	gl.Width = width
	gl.Height = height
	gl.Seed = seed

	if refuelCellsAllotment > 0 {
		gl.RefuelCellsAllotment = refuelCellsAllotment
	} else {
		gl.RefuelCellsAllotment = 0.05
	}

	if depotCellsAllotment > 0 {
		gl.DepotCellsAllotment = depotCellsAllotment
	} else {
		gl.DepotCellsAllotment = 0.02
	}

	if blockedCellsAllotment > 0 {
		gl.BlockedCellsAllotment = blockedCellsAllotment
	} else {
		gl.BlockedCellsAllotment = 0.05
	}

	if roadDensity > 0 {
		gl.RoadDensity = roadDensity
	} else {
		gl.RoadDensity = 0.5
	}

	if mainRoadBias > 0 {
		gl.MainRoadBias = mainRoadBias
	} else {
		gl.MainRoadBias = 0.2
	}

	if deadEndBias > 0 {
		gl.DeadEndBias = deadEndBias
	} else {
		gl.DeadEndBias = 0.1
	}

	gl.SegmentIDCounter = 1

	gl.GenerationStatsSu = &GenerationStats{}
}

func PrettyPrint(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func NewGridLoader() *GridLoader {
	//initial config is demo based
	gl := GridLoaderDemo
	return &gl
}

func PrintValGrid() {
	Ct := NewGridLoader()
	fmt.Println(Ct)

}

type GridParseError struct {
	Details string
}

func (e *GridParseError) Error() string {
	return fmt.Sprintf("[GridParseError]: %s", e.Details)
}

func (gl *GridLoader) validateImportedGrid(grid *Grid) error {
	if grid.DimX <= 0 || grid.DimY <= 0 {
		return fmt.Errorf("invalid grid dimensions: %dx%d (must be positive)", grid.DimX, grid.DimY)
	}

	expectedCells := int(grid.DimX * grid.DimY)
	if len(grid.Cells) != expectedCells {
		return fmt.Errorf("cell count mismatch: expected %d cells for %dx%d grid, got %d",
			expectedCells, grid.DimX, grid.DimY, len(grid.Cells))
	}

	coordinatesSeen := make(map[[2]int64]bool)
	for i, cell := range grid.Cells {

		if cell.Xpos < 0 || cell.Xpos >= grid.DimX ||
			cell.Ypos < 0 || cell.Ypos >= grid.DimY {
			return fmt.Errorf("cell %d has coordinates (%d,%d) outside grid bounds %dx%d",
				i, cell.Xpos, cell.Ypos, grid.DimX, grid.DimY)
		}

		coords := [2]int64{cell.Xpos, cell.Ypos}
		if coordinatesSeen[coords] {
			return fmt.Errorf("duplicate cell coordinates (%d,%d) found", cell.Xpos, cell.Ypos)
		}
		coordinatesSeen[coords] = true

		switch cell.CellType {
		case CellTypeNormal, CellTypeRefuel,
			CellTypeDepot, CellTypeBlocked:

		default:
			return fmt.Errorf("cell %d has invalid cell type: %s", i, cell.CellType)
		}

		if cell.CellType == CellTypeRefuel && cell.RefuelAmount == nil {
			return fmt.Errorf("refuel station at (%d,%d) missing refuel amount", cell.Xpos, cell.Ypos)
		}

		for j, cellRoad := range cell.RoadSegments {
			segment := cellRoad.RoadSegment
			if segment.ID <= 0 {
				return fmt.Errorf("cell (%d,%d) road segment %d has invalid ID: %d",
					cell.Xpos, cell.Ypos, j, segment.ID)
			}

			if !gl.isValidSegmentForCell(segment, cell) {
				return fmt.Errorf("cell (%d,%d) contains segment %d with invalid coordinates",
					cell.Xpos, cell.Ypos, segment.ID)
			}
		}
	}

	if err := gl.validateRoadConnectivity(grid); err != nil {
		return fmt.Errorf("road connectivity validation failed: %w", err)
	}

	return nil
}

func (gl *GridLoader) isValidSegmentForCell(segment RoadSegment, cell Cell) bool {

	cellX, cellY := cell.Xpos, cell.Ypos
	startX, startY := segment.StartX, segment.StartY
	endX, endY := segment.EndX, segment.EndY

	if (startX == cellX && startY == cellY) || (endX == cellX && endY == cellY) {
		return true
	}

	if startX == endX {
		if cellX == startX {
			minY, maxY := startY, endY
			if minY > maxY {
				minY, maxY = maxY, minY
			}
			return cellY >= minY && cellY <= maxY
		}
	} else if startY == endY {
		if cellY == startY {
			minX, maxX := startX, endX
			if minX > maxX {
				minX, maxX = maxX, minX
			}
			return cellX >= minX && cellX <= maxX
		}
	}

	return false
}
func (gl *GridLoader) validateRoadConnectivity(grid *Grid) error {
	adjacency := make(map[int64][]int64)
	allSegments := make(map[int64]bool)

	for _, cell := range grid.Cells {
		for _, cellRoad := range cell.RoadSegments {
			segment := cellRoad.RoadSegment
			allSegments[segment.ID] = true

			connections := gl.findConnectedSegments(segment, grid)
			adjacency[segment.ID] = connections
		}
	}

	if len(allSegments) == 0 {
		return nil
	}

	//bfs search
	visited := make(map[int64]bool)
	componentCount := 0

	for segmentID := range allSegments {
		if !visited[segmentID] {
			componentCount++
			gl.BFSMarkConnected(segmentID, adjacency, visited)
		}
	}

	if componentCount > 1 {
		return fmt.Errorf("road network has %d isolated components (should be 1 for full connectivity)", componentCount)
	}

	return nil
}
func (gl *GridLoader) findConnectedSegments(target RoadSegment, grid *Grid) []int64 {
	var connections []int64

	targetStartX, targetStartY := target.StartX, target.StartY
	targetEndX, targetEndY := target.EndX, target.EndY

	for _, cell := range grid.Cells {
		for _, cellRoad := range cell.RoadSegments {
			other := cellRoad.RoadSegment
			if other.ID == target.ID {
				continue
			}

			if (other.StartX == targetStartX && other.StartY == targetStartY) ||
				(other.StartX == targetEndX && other.StartY == targetEndY) ||
				(other.EndX == targetStartX && other.EndY == targetStartY) ||
				(other.EndX == targetEndX && other.EndY == targetEndY) {
				connections = append(connections, other.ID)
			}
		}
	}

	return connections
}

func (gl *GridLoader) BFSMarkConnected(segmentID int64, adjacency map[int64][]int64, explored map[int64]bool) {
	size := len(adjacency)
	queue := utils.NewIntQueue(size)

	explored[segmentID] = true
	queue.Enqueue(segmentID)
	for !queue.IsEmpty() {
		currSeg := queue.Dequeue()
		for _, neighbor := range adjacency[currSeg] {
			if !explored[neighbor] {
				explored[neighbor] = true
				queue.Enqueue(neighbor)
			}
		}
	}
}

// func (gl *GridLoader) dfsMarkConnected(segmentID int64, adjacency map[int64][]int64, visited map[int64]bool) {
// 	visited[segmentID] = true

//		for _, connectedID := range adjacency[segmentID] {
//			if !visited[connectedID] {
//				gl.dfsMarkConnected(connectedID, adjacency, visited)
//			}
//		}
//	}
func (gl *GridLoader) countRoadSegments(grid *Grid) int {
	segmentsSeen := make(map[int64]bool)

	for _, cell := range grid.Cells {
		for _, cellRoad := range cell.RoadSegments {
			segmentsSeen[cellRoad.RoadSegment.ID] = true
		}
	}

	return len(segmentsSeen)
}

func (gl *GridLoader) analyzeGenerationResults(grid *Grid) {
	roadCells := 0
	specialCells := 0

	for _, cell := range grid.Cells {
		if len(cell.RoadSegments) > 0 {
			roadCells++
		}

		if cell.CellType != CellTypeNormal {
			specialCells++
		}
	}

	gl.GenerationStatsSu.RoadCells = roadCells
	gl.GenerationStatsSu.SpecialCells = specialCells
	gl.GenerationStatsSu.TotalSegments = gl.countRoadSegments(grid)

	// Additional analysis like dead end counting and connectivity analysis
	// will be added as we implement the road generation algorithms
}

// TO USE LATER
func (gl *GridLoader) logGenerationStats() {
	stats := gl.GenerationStatsSu
	if stats == nil {
		return
	}

	fmt.Printf("\n=== Grid Generation Statistics ===\n")
	fmt.Printf("Total cells: %d\n", stats.TotalCells)
	fmt.Printf("Road cells: %d (%.1f%%)\n", stats.RoadCells,
		float64(stats.RoadCells)/float64(stats.TotalCells)*100)
	fmt.Printf("Special cells: %d (%.1f%%)\n", stats.SpecialCells,
		float64(stats.SpecialCells)/float64(stats.TotalCells)*100)
	fmt.Printf("Total road segments: %d\n", stats.TotalSegments)
	fmt.Printf("Generation time: %d ms\n", stats.GenerationTimeMs)
	fmt.Printf("=====================================\n\n")
}
