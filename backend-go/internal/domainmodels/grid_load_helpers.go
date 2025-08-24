package domainmodels

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)



func (gl *GridLoader) buildSpatialIndexes(grid *Grid) {
	fmt.Printf("Building spatial indexes for %d cells...\n", len(grid.Cells))

	grid.CoordIndex = make(map[[2]int64]*Cell)
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		grid.CoordIndex[[2]int64{cell.Xpos, cell.Ypos}] = cell
	}

	grid.SegmentIndex = make(map[int64]*Cell)
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		for _, road := range cell.RoadSegments {
			grid.SegmentIndex[road.RoadSegmentID] = cell
		}
	}

	adjacency := make(map[int64][]int64)
	directions := [][2]int64{
		{0, -1},
		{0, 1},
		{-1, 0},
		{1, 0},
	}

	for _, cell := range grid.Cells {
		for _, road := range cell.RoadSegments {

			for _, d := range directions {
				neighborPos := [2]int64{cell.Xpos + d[0], cell.Ypos + d[1]}
				if neighbor, ok := grid.CoordIndex[neighborPos]; ok {

					for _, neighborRoad := range neighbor.RoadSegments {
						adjacency[road.RoadSegmentID] = append(adjacency[road.RoadSegmentID], neighborRoad.RoadSegmentID)
					}
				}
			}
		}
	}

	grid.RoadGraph = &RoadGraph{Adjacency: adjacency}

	fmt.Printf(
		"Spatial indexing completed: %d coord mappings, %d road segments, %d adjacency entries\n",
		len(grid.CoordIndex), len(grid.SegmentIndex), len(adjacency),
	)
}

func (gl *GridLoader) LoadFromJSON(filepath string) (*Grid, error) {
	startTime := time.Now()

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read grid file %s: %w", filepath, err)
	}

	var grid Grid
	if err := json.Unmarshal(data, &grid); err != nil {
		return nil, fmt.Errorf("failed to parse grid JSON from %s: %w", filepath, err)
	}

	if err := gl.validateImportedGrid(&grid); err != nil {
		return nil, fmt.Errorf("imported grid from %s failed validation: %w", filepath, err)
	}

	gl.buildSpatialIndexes(&grid)

	gl.GenerationStatsSu = &GenerationStats{
		TotalCells:       len(grid.Cells),
		TotalSegments:    gl.countRoadSegments(&grid),
		GenerationTimeMs: time.Since(startTime).Milliseconds(),
	}

	fmt.Printf("Successfully loaded %dx%d grid from %s\n", grid.DimX, grid.DimY, filepath)
	fmt.Printf("Grid contains %d cells with %d road segments\n",
		gl.GenerationStatsSu.TotalCells, gl.GenerationStatsSu.TotalSegments)

	
return &grid, nil
}

//TEST FUNCTION NOT REAL GEN LOGIC

func (gl *GridLoader) GenerateRandomTestGrid() *Grid {
    grid := &Grid{
        DimX:  gl.Width,
        DimY:  gl.Height,
        Cells: make([]Cell, 0),
    }

    idCounter := int64(1)
    for x := int64(0); x < gl.Width; x++ {
        for y := int64(0); y < gl.Height; y++ {
            cell := Cell{Xpos: x, Ypos: y, CellType: CellTypeNormal}

            if y%3 == 0 {
                seg := RoadSegment{
                    ID:     idCounter,
                    StartX: x, StartY: y,
                    EndX:   x + 1, EndY: y,
                }
                cell.RoadSegments = append(cell.RoadSegments, CellRoad{
                    RoadSegmentID: seg.ID,
                    RoadSegment:   seg,
                })
                idCounter++
            }

            grid.Cells = append(grid.Cells, cell)
        }
    }

    return grid
}
