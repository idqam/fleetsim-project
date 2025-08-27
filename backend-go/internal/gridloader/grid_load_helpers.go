package gridloader

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"owenvi.com/fleetsim/internal/domainmodels"
)

func (gl *GridLoader) buildSpatialIndexes(grid *domainmodels.Grid) {
	fmt.Printf("Building spatial indexes for %d cells...\n", len(grid.Cells))

	grid.CoordIndex = make(map[[2]int64]*domainmodels.Cell)
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		grid.CoordIndex[[2]int64{cell.Xpos, cell.Ypos}] = cell
	}

	grid.SegmentIndex = make(map[int64]*domainmodels.Cell)
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

	grid.RoadGraph = &domainmodels.RoadGraph{Adjacency: adjacency}

	fmt.Printf(
		"Spatial indexing completed: %d coord mappings, %d road segments, %d adjacency entries\n",
		len(grid.CoordIndex), len(grid.SegmentIndex), len(adjacency),
	)
}

func (gl *GridLoader) LoadFromJSON(filepath string) (*domainmodels.Grid, error) {
	startTime := time.Now()

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read grid file %s: %w", filepath, err)
	}

	var grid domainmodels.Grid
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

func (gl *GridLoader) GetGenerationStats() *GenerationStats {
	if gl.GenerationStatsSu == nil {
		return &GenerationStats{}
	}
	return gl.GenerationStatsSu
}

func (gl *GridLoader) calculateSafeParameters(grid *domainmodels.Grid) (float64, float64, float64) {
	roadCells := gl.countRoadCells(grid)
	criticalConnections := gl.identifyCriticalConnections(grid)
	networkRedundancy := gl.calculateNetworkRedundancy(grid)

	maxSafeBlocked := gl.calculateMaxBlockedAreas(roadCells, criticalConnections, networkRedundancy)

	safeBlockedAllotment := math.Min(gl.BlockedCellsAllotment, maxSafeBlocked)
	safeRefuelAllotment := gl.RefuelCellsAllotment
	safeDepotAllotment := gl.DepotCellsAllotment

	return safeBlockedAllotment, safeRefuelAllotment, safeDepotAllotment
}

type CellStateBackup struct {
	cellStates map[[2]int64]CellBackupData
}

type CellBackupData struct {
	cellType     domainmodels.CellType
	refuelAmount *float64
}

func (gl *GridLoader) createCellStateBackup(grid *domainmodels.Grid) *CellStateBackup {
	backup := &CellStateBackup{
		cellStates: make(map[[2]int64]CellBackupData),
	}

	for i := range grid.Cells {
		cell := &grid.Cells[i]
		coords := [2]int64{cell.Xpos, cell.Ypos}

		var refuelAmountCopy *float64
		if cell.RefuelAmount != nil {
			amount := *cell.RefuelAmount
			refuelAmountCopy = &amount
		}

		backup.cellStates[coords] = CellBackupData{
			cellType:     cell.CellType,
			refuelAmount: refuelAmountCopy,
		}
	}

	fmt.Printf("Created backup of %d cell states\n", len(backup.cellStates))
	return backup
}

func (gl *GridLoader) restoreCellStateFromBackup(grid *domainmodels.Grid, backup *CellStateBackup) {
	restoredCount := 0

	for i := range grid.Cells {
		cell := &grid.Cells[i]
		coords := [2]int64{cell.Xpos, cell.Ypos}

		if backupData, exists := backup.cellStates[coords]; exists {
			cell.CellType = backupData.cellType

			if backupData.refuelAmount != nil {
				amount := *backupData.refuelAmount
				cell.RefuelAmount = &amount
			} else {
				cell.RefuelAmount = nil
			}

			restoredCount++
		} else {
			fmt.Printf("Warning: No backup data found for cell at (%d,%d)\n", cell.Xpos, cell.Ypos)
		}
	}

	fmt.Printf("Restored %d cells to their backed-up state\n", restoredCount)
}

func (gl *GridLoader) placeSpecialLocationsHybrid(grid *domainmodels.Grid, rng *rand.Rand) error {
	safeBlocked, _, _ := gl.calculateSafeParameters(grid)

	originalBlocked := gl.BlockedCellsAllotment
	originalRefuel := gl.RefuelCellsAllotment
	originalDepot := gl.DepotCellsAllotment

	maxAttempts := 3

	for attempt := 0; attempt < maxAttempts; attempt++ {
		var targetBlocked, targetRefuel, targetDepot float64

		if attempt == 0 {
			targetBlocked = math.Min(originalBlocked, safeBlocked*1.2)
			targetRefuel = originalRefuel
			targetDepot = originalDepot
			fmt.Printf("Attempt 1: Trying optimistic parameters (blocked=%.3f)\n", targetBlocked)
		} else {
			aggressiveness := 1.0 - (float64(attempt-1) * 0.3)
			targetBlocked = safeBlocked * aggressiveness
			targetRefuel = originalRefuel * (0.9 + aggressiveness*0.1)
			targetDepot = originalDepot * (0.9 + aggressiveness*0.1)
			fmt.Printf("Attempt %d: Trying conservative parameters (blocked=%.3f)\n", attempt+1, targetBlocked)
		}

		gl.BlockedCellsAllotment = targetBlocked
		gl.RefuelCellsAllotment = targetRefuel
		gl.DepotCellsAllotment = targetDepot

		cellBackup := gl.createCellStateBackup(grid)

		err := gl.placeSpecialLocations(grid, rng)
		if err == nil {
			gl.BlockedCellsAllotment = originalBlocked
			gl.RefuelCellsAllotment = originalRefuel
			gl.DepotCellsAllotment = originalDepot
			fmt.Printf("Hybrid placement succeeded on attempt %d\n", attempt+1)
			return nil
		}

		gl.restoreCellStateFromBackup(grid, cellBackup)
	}

	gl.BlockedCellsAllotment = originalBlocked
	gl.RefuelCellsAllotment = originalRefuel
	gl.DepotCellsAllotment = originalDepot

	return fmt.Errorf("hybrid placement failed after %d attempts", maxAttempts)
}

func (gl *GridLoader) countRoadCells(grid *domainmodels.Grid) int {
	roadCellCount := 0
	for _, cell := range grid.Cells {
		if len(cell.RoadSegments) > 0 {
			roadCellCount++
		}
	}
	return roadCellCount
}

func (gl *GridLoader) identifyCriticalConnections(grid *domainmodels.Grid) int {
	segmentMap := make(map[int64]bool)
	for _, cell := range grid.Cells {
		for _, cellRoad := range cell.RoadSegments {
			segmentMap[cellRoad.RoadSegment.ID] = true
		}
	}

	criticalCount := 0

	for segmentID := range segmentMap {
		tempGrid := gl.createGridWithoutSegment(grid, segmentID)
		components := gl.findConnectedComponents(tempGrid)
		if len(components) > 1 {
			criticalCount++
		}
	}

	return criticalCount
}

func (gl *GridLoader) calculateNetworkRedundancy(grid *domainmodels.Grid) float64 {
	if grid.RoadGraph == nil || len(grid.RoadGraph.Adjacency) == 0 {
		return 0.0
	}

	totalConnections := 0
	totalNodes := len(grid.RoadGraph.Adjacency)

	for _, neighbors := range grid.RoadGraph.Adjacency {
		totalConnections += len(neighbors)
	}

	actualConnections := float64(totalConnections) / 2.0
	maxPossibleConnections := float64(totalNodes*(totalNodes-1)) / 2.0

	if maxPossibleConnections == 0 {
		return 0.0
	}

	return actualConnections / maxPossibleConnections
}

func (gl *GridLoader) calculateMaxBlockedAreas(roadCells int, criticalConnections int, networkRedundancy float64) float64 {
	baseCapacity := float64(roadCells) * 0.15

	criticalPenalty := float64(criticalConnections) * 0.02
	redundancyBonus := networkRedundancy * 0.1

	maxBlocked := baseCapacity - criticalPenalty + redundancyBonus

	if maxBlocked < 0.01 {
		maxBlocked = 0.01
	}

	totalCells := float64(gl.Width * gl.Height)
	maxBlockedRatio := maxBlocked / totalCells

	if maxBlockedRatio > 0.25 {
		maxBlockedRatio = 0.25
	}

	return maxBlockedRatio
}

func (gl *GridLoader) createGridWithoutSegment(grid *domainmodels.Grid, excludeSegmentID int64) *domainmodels.Grid {
	tempGrid := &domainmodels.Grid{
		DimX:  grid.DimX,
		DimY:  grid.DimY,
		Cells: make([]domainmodels.Cell, len(grid.Cells)),
	}

	for i, cell := range grid.Cells {
		tempGrid.Cells[i] = domainmodels.Cell{
			Xpos:         cell.Xpos,
			Ypos:         cell.Ypos,
			CellType:     cell.CellType,
			RoadSegments: make([]domainmodels.CellRoad, 0),
			RefuelAmount: cell.RefuelAmount,
		}

		for _, cellRoad := range cell.RoadSegments {
			if cellRoad.RoadSegment.ID != excludeSegmentID {
				tempGrid.Cells[i].RoadSegments = append(tempGrid.Cells[i].RoadSegments, cellRoad)
			}
		}
	}

	return tempGrid
}

func (gl *GridLoader) placeSpecialLocationsWithRetry(grid *domainmodels.Grid, rng *rand.Rand) error {
	maxAttempts := 20

	originalBlockedAllotment := gl.BlockedCellsAllotment
	originalRefuelAllotment := gl.RefuelCellsAllotment
	originalDepotAllotment := gl.DepotCellsAllotment

	for attempt := 0; attempt < maxAttempts; attempt++ {
		fmt.Printf("Special location placement attempt %d/%d (blocked: %.3f, refuel: %.3f, depot: %.3f)\n",
			attempt+1, maxAttempts, gl.BlockedCellsAllotment, gl.RefuelCellsAllotment, gl.DepotCellsAllotment)

		cellStateBackup := make(map[[2]int64]struct {
			cellType     domainmodels.CellType
			refuelAmount *float64
		})

		for i := range grid.Cells {
			cell := &grid.Cells[i]
			coords := [2]int64{cell.Xpos, cell.Ypos}
			cellStateBackup[coords] = struct {
				cellType     domainmodels.CellType
				refuelAmount *float64
			}{cell.CellType, cell.RefuelAmount}
		}

		err := gl.placeSpecialLocations(grid, rng)

		if err == nil {

			gl.BlockedCellsAllotment = originalBlockedAllotment
			gl.RefuelCellsAllotment = originalRefuelAllotment
			gl.DepotCellsAllotment = originalDepotAllotment
			fmt.Printf("Special location placement succeeded on attempt %d\n", attempt+1)
			return nil
		}

		fmt.Printf("Attempt %d failed: %v\n", attempt+1, err)
		for i := range grid.Cells {
			cell := &grid.Cells[i]
			coords := [2]int64{cell.Xpos, cell.Ypos}
			if backup, exists := cellStateBackup[coords]; exists {
				cell.CellType = backup.cellType
				cell.RefuelAmount = backup.refuelAmount
			}
		}

		gl.BlockedCellsAllotment *= 0.7
		gl.RefuelCellsAllotment *= 0.95
		gl.DepotCellsAllotment *= 0.95

		if gl.BlockedCellsAllotment < 0.01 {
			gl.BlockedCellsAllotment = 0.01
		}
		if gl.RefuelCellsAllotment < 0.02 {
			gl.RefuelCellsAllotment = 0.02
		}
		if gl.DepotCellsAllotment < 0.01 {
			gl.DepotCellsAllotment = 0.01
		}
	}

	gl.BlockedCellsAllotment = originalBlockedAllotment
	gl.RefuelCellsAllotment = originalRefuelAllotment
	gl.DepotCellsAllotment = originalDepotAllotment

	return fmt.Errorf("failed to place special locations after %d attempts - network connectivity cannot be maintained with any parameter combination tried", maxAttempts)
}
func (gl *GridLoader) GenerateProcedural() (*domainmodels.Grid, error) {
	startTime := time.Now()
	maxRetries := 10

	rng := rand.New(rand.NewSource(gl.Seed))

	for attempt := 0; attempt < maxRetries; attempt++ {
		grid := gl.initializeEmptyGrid()

		if err := gl.generateRoadNetwork(grid, rng); err != nil {
			continue
		}

		if err := gl.placeSpecialLocationsHybrid(grid, rng); err != nil {
			if attempt == maxRetries-1 {
				return nil, err
			}
			continue
		}

		if err := gl.validateAndRepairConnectivity(grid); err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("connectivity validation failed: %w", err)
			}
			continue
		}

		gl.buildSpatialIndexes(grid)

		if gl.GenerationStatsSu == nil {
			gl.GenerationStatsSu = &GenerationStats{}
		}
		gl.GenerationStatsSu.TotalCells = len(grid.Cells)
		gl.GenerationStatsSu.GenerationTimeMs = time.Since(startTime).Milliseconds()

		gl.analyzeGenerationResults(grid)
		gl.logGenerationStats()
		gl.currentGrid = grid
		return grid, nil
	}

	return nil, fmt.Errorf("failed to generate valid grid after %d attempts", maxRetries)
}
func (gl *GridLoader) initializeEmptyGrid() *domainmodels.Grid {
	totalCells := int(gl.Width * gl.Height)
	cells := make([]domainmodels.Cell, 0, totalCells)

	for y := int64(0); y < gl.Height; y++ {
		for x := int64(0); x < gl.Width; x++ {
			cell := domainmodels.Cell{
				Xpos:         x,
				Ypos:         y,
				CellType:     domainmodels.CellTypeNormal,
				RoadSegments: make([]domainmodels.CellRoad, 0),
				RefuelAmount: nil,
			}
			cells = append(cells, cell)
		}
	}

	return &domainmodels.Grid{
		DimX:  gl.Width,
		DimY:  gl.Height,
		Cells: cells,
	}
}

func (gl *GridLoader) generateRoadNetwork(grid *domainmodels.Grid, rng *rand.Rand) error {
	fmt.Printf("Generating road network with density %.2f...\n", gl.RoadDensity)

	mainArteriesCreated := gl.createMainArteries(grid, rng)

	secondaryRoadsCreated := gl.createSecondaryRoads(grid, rng)

	connectivityRoadsCreated := gl.fillConnectivityGaps(grid, rng)

	gl.GenerationStatsSu.MainArteries = mainArteriesCreated
	gl.GenerationStatsSu.SecondaryRoads = secondaryRoadsCreated

	fmt.Printf("Created %d main arteries, %d secondary roads, %d connectivity segments\n",
		mainArteriesCreated, secondaryRoadsCreated, connectivityRoadsCreated)

	return nil
}

func (gl *GridLoader) createMainArteries(grid *domainmodels.Grid, rng *rand.Rand) int {
	arteriesCreated := 0

	min, max := 2, 20

	b := rand.Intn(max-min+1) + min

	a := rand.Intn(max-b+1) + b

	horizontalArteries := gl.selectMainRoadPositions(gl.Height, b, a, rng)
	for _, y := range horizontalArteries {
		if gl.createHorizontalRoad(grid, y) {
			arteriesCreated++
		}
	}

	verticalArteries := gl.selectMainRoadPositions(gl.Width, b, a, rng)
	for _, x := range verticalArteries {
		if gl.createVerticalRoad(grid, x) {
			arteriesCreated++
		}
	}

	return arteriesCreated
}

func (gl *GridLoader) selectMainRoadPositions(dimension int64, minRoads, maxRoads int, rng *rand.Rand) []int64 {
	numRoads := minRoads + rng.Intn(maxRoads-minRoads+1)
	positions := make([]int64, 0, numRoads)

	step := dimension / int64(numRoads+1)

	for i := 0; i < numRoads; i++ {

		basePos := int64(i+1) * step

		variation := step / 5
		if variation > 0 {
			offset := rng.Int63n(variation*2) - variation
			basePos += offset

			if basePos < 1 {
				basePos = 1
			}
			if basePos >= dimension-1 {
				basePos = dimension - 2
			}
		}

		positions = append(positions, basePos)
	}

	return positions
}

func (gl *GridLoader) createHorizontalRoad(grid *domainmodels.Grid, y int64) bool {
	segmentsCreated := 0

	for x := int64(0); x < gl.Width-1; x++ {
		segment := domainmodels.RoadSegment{
			ID:       gl.SegmentIDCounter,
			StartX:   x,
			StartY:   y,
			EndX:     x + 1,
			EndY:     y,
			IsOpen:   true,
			Capacity: gl.getDefaultCapacityForSegment(),
		}

		gl.addSegmentToCell(grid, x, y, segment)
		gl.addSegmentToCell(grid, x+1, y, segment)

		gl.SegmentIDCounter++
		segmentsCreated++
	}

	return segmentsCreated > 0
}

func (gl *GridLoader) createVerticalRoad(grid *domainmodels.Grid, x int64) bool {
	segmentsCreated := 0

	for y := int64(0); y < gl.Height-1; y++ {
		segment := domainmodels.RoadSegment{
			ID:       gl.SegmentIDCounter,
			StartX:   x,
			StartY:   y,
			EndX:     x,
			EndY:     y + 1,
			IsOpen:   true,
			Capacity: gl.getDefaultCapacityForSegment(),
		}

		gl.addSegmentToCell(grid, x, y, segment)
		gl.addSegmentToCell(grid, x, y+1, segment)

		gl.SegmentIDCounter++
		segmentsCreated++
	}

	return segmentsCreated > 0
}

func (gl *GridLoader) createSecondaryRoads(grid *domainmodels.Grid, rng *rand.Rand) int {
	roadsCreated := 0

	for i := range grid.Cells {
		cell := &grid.Cells[i]

		if len(cell.RoadSegments) > 0 {
			continue
		}

		roadProbability := gl.calculateSecondaryRoadProbability(grid, cell, rng)

		if rng.Float64() < roadProbability {
			connectionsAdded := gl.addLocalRoadConnections(grid, cell, rng)
			if connectionsAdded > 0 {
				roadsCreated++
			}
		}
	}

	return roadsCreated
}

func (gl *GridLoader) calculateSecondaryRoadProbability(grid *domainmodels.Grid, cell *domainmodels.Cell, rng *rand.Rand) float64 {
	baseProbability := gl.RoadDensity * 0.3

	nearbyRoadBonus := gl.calculateNearbyRoadInfluence(grid, cell.Xpos, cell.Ypos)

	centerBias := gl.calculateCenterBias(cell.Xpos, cell.Ypos) * 0.1

	totalProbability := baseProbability + nearbyRoadBonus + centerBias
	if totalProbability > 1.0 {
		totalProbability = 1.0
	}

	return totalProbability
}

func (gl *GridLoader) calculateNearbyRoadInfluence(grid *domainmodels.Grid, x, y int64) float64 {
	roadsNearby := 0
	totalChecked := 0

	for dx := int64(-1); dx <= 1; dx++ {
		for dy := int64(-1); dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}

			adjX, adjY := x+dx, y+dy

			if adjX >= 0 && adjX < gl.Width && adjY >= 0 && adjY < gl.Height {
				adjCell := gl.getCellAtGrid(grid, adjX, adjY)
				if adjCell != nil && len(adjCell.RoadSegments) > 0 {
					roadsNearby++
				}
				totalChecked++
			}
		}
	}

	if totalChecked == 0 {
		return 0.0
	}

	influence := float64(roadsNearby) / float64(totalChecked)
	return influence * 0.4
}

func (gl *GridLoader) calculateCenterBias(x, y int64) float64 {
	centerX := gl.Width / 2
	centerY := gl.Height / 2

	maxDistance := float64(centerX*centerX + centerY*centerY)
	actualDistance := float64((x-centerX)*(x-centerX) + (y-centerY)*(y-centerY))
	normalizedDistance := actualDistance / maxDistance

	centerBias := 1.0 - normalizedDistance

	return centerBias
}

func (gl *GridLoader) addLocalRoadConnections(grid *domainmodels.Grid, cell *domainmodels.Cell, rng *rand.Rand) int {
	connectionsAdded := 0

	directions := []struct{ dx, dy int64 }{
		{0, -1}, // North
		{0, 1},  // South
		{1, 0},  // East
		{-1, 0}, // West
	}

	maxConnections := 2 + rng.Intn(3)
	connectionsAttempted := 0

	rng.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		if connectionsAttempted >= maxConnections {
			break
		}

		targetX := cell.Xpos + dir.dx
		targetY := cell.Ypos + dir.dy

		if targetX >= 0 && targetX < gl.Width && targetY >= 0 && targetY < gl.Height {
			if gl.createConnectionSegment(grid, cell.Xpos, cell.Ypos, targetX, targetY) {
				connectionsAdded++
			}
		}

		connectionsAttempted++
	}

	return connectionsAdded
}

func (gl *GridLoader) createConnectionSegment(grid *domainmodels.Grid, fromX, fromY, toX, toY int64) bool {
	segment := domainmodels.RoadSegment{
		ID:       gl.SegmentIDCounter,
		StartX:   fromX,
		StartY:   fromY,
		EndX:     toX,
		EndY:     toY,
		IsOpen:   true,
		Capacity: gl.getDefaultCapacityForSegment(),
	}

	gl.addSegmentToCell(grid, fromX, fromY, segment)
	gl.addSegmentToCell(grid, toX, toY, segment)

	gl.SegmentIDCounter++
	return true
}

func (gl *GridLoader) fillConnectivityGaps(grid *domainmodels.Grid, rng *rand.Rand) int {
	connectivitySegmentsAdded := 0

	totalPossibleConnections := int(gl.Width * gl.Height * 2)
	currentConnections := gl.countRoadSegments(grid)
	targetConnections := int(float64(totalPossibleConnections) * gl.RoadDensity)

	connectionsNeeded := targetConnections - currentConnections
	if connectionsNeeded <= 0 {
		return 0
	}

	fmt.Printf("Need %d more connections to reach target density\n", connectionsNeeded)

	attempts := 0
	maxAttempts := connectionsNeeded * 3

	for connectivitySegmentsAdded < connectionsNeeded && attempts < maxAttempts {
		attempts++

		cellIndex := rng.Intn(len(grid.Cells))
		cell := &grid.Cells[cellIndex]

		if gl.addRandomConnection(grid, cell, rng) {
			connectivitySegmentsAdded++
		}
	}

	return connectivitySegmentsAdded
}

func (gl *GridLoader) addRandomConnection(grid *domainmodels.Grid, cell *domainmodels.Cell, rng *rand.Rand) bool {
	directions := []struct{ dx, dy int64 }{
		{0, -1}, {0, 1}, {1, 0}, {-1, 0},
	}

	rng.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		targetX := cell.Xpos + dir.dx
		targetY := cell.Ypos + dir.dy

		if targetX >= 0 && targetX < gl.Width && targetY >= 0 && targetY < gl.Height {

			if !gl.connectionExists(grid, cell.Xpos, cell.Ypos, targetX, targetY) {
				return gl.createConnectionSegment(grid, cell.Xpos, cell.Ypos, targetX, targetY)
			}
		}
	}

	return false
}

func (gl *GridLoader) getCellAtGrid(grid *domainmodels.Grid, x, y int64) *domainmodels.Cell {
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		if cell.Xpos == x && cell.Ypos == y {
			return cell
		}
	}
	return nil
}

func (gl *GridLoader) addSegmentToCell(grid *domainmodels.Grid, x, y int64, segment domainmodels.RoadSegment) {
	cell := gl.getCellAtGrid(grid, x, y)
	if cell != nil {
		cellRoad := domainmodels.CellRoad{
			RoadSegmentID: segment.ID,
			RoadSegment:   segment,
		}
		cell.RoadSegments = append(cell.RoadSegments, cellRoad)
	}
}

func (gl *GridLoader) connectionExists(grid *domainmodels.Grid, fromX, fromY, toX, toY int64) bool {
	fromCell := gl.getCellAtGrid(grid, fromX, fromY)
	if fromCell == nil {
		return false
	}

	for _, cellRoad := range fromCell.RoadSegments {
		segment := cellRoad.RoadSegment
		if (segment.StartX == fromX && segment.StartY == fromY && segment.EndX == toX && segment.EndY == toY) ||
			(segment.StartX == toX && segment.StartY == toY && segment.EndX == fromX && segment.EndY == fromY) {
			return true
		}
	}

	return false
}

func (gl *GridLoader) getDefaultCapacityForSegment() *int64 {
	capacity := int64(15)
	return &capacity
}
