package gridloader

import (
	"fmt"
	"math/rand"

	"owenvi.com/fleetsim/internal/domainmodels"
	"owenvi.com/fleetsim/internal/utils"
)


func (gl *GridLoader) placeSpecialLocations(grid *domainmodels.Grid, rng *rand.Rand) error {
	fmt.Printf("Placing special locations (%.1f%% fuel, %.1f%% depot, %.1f%% blocked)...\n",
		gl.RefuelCellsAllotment*100, gl.DepotCellsAllotment*100, gl.BlockedCellsAllotment*100)
	

	eligibleCells := gl.findEligibleCells(grid)
	if len(eligibleCells) == 0 {
		return fmt.Errorf("no eligible cells found for special location placement")
	}
	
	totalCells := len(grid.Cells)
	fuelStationsNeeded := int(float64(totalCells) * gl.RefuelCellsAllotment)
	depotsNeeded := int(float64(totalCells) * gl.DepotCellsAllotment)
	blockedAreasNeeded := int(float64(totalCells) * gl.BlockedCellsAllotment)
	
	fmt.Printf("Creating %d fuel stations, %d depots, %d blocked areas from %d eligible cells\n",
		fuelStationsNeeded, depotsNeeded, blockedAreasNeeded, len(eligibleCells))
	
	if err := gl.placeFuelStations(grid, eligibleCells, fuelStationsNeeded, rng); err != nil {
		return fmt.Errorf("fuel station placement failed: %w", err)
	}
	
	if err := gl.placeDepots(grid, eligibleCells, depotsNeeded, rng); err != nil {
		return fmt.Errorf("depot placement failed: %w", err)
	}
	
	if err := gl.placeBlockedAreas(grid, eligibleCells, blockedAreasNeeded, rng); err != nil {
		return fmt.Errorf("blocked area placement failed: %w", err)
	}
	
	if err := gl.validatePostPlacementConnectivity(grid); err != nil {
		return fmt.Errorf("special location placement broke network connectivity: %w", err)
	}
	
	return nil
}

func (gl *GridLoader) findEligibleCells(grid *domainmodels.Grid) []*domainmodels.Cell {
	var eligible []*domainmodels.Cell
	
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		
		if cell.CellType != domainmodels.CellTypeNormal {
			continue 
		}
		
		if len(cell.RoadSegments) == 0 {
			continue 
		}
		
		if gl.isCellStrategicallyEligible(grid, cell) {
			eligible = append(eligible, cell)
		}
	}
	
	return eligible
}

func (gl *GridLoader) isCellStrategicallyEligible(grid *domainmodels.Grid, cell *domainmodels.Cell) bool {

	margin := int64(1)
	if cell.Xpos < margin || cell.Xpos >= gl.Width-margin ||
	   cell.Ypos < margin || cell.Ypos >= gl.Height-margin {
		return false
	}

	connectionCount := utils.CountCellConnections(grid, cell)
	return connectionCount >= 2

}




func (gl *GridLoader) placeFuelStations(grid *domainmodels.Grid, eligibleCells []*domainmodels.Cell, count int, rng *rand.Rand) error {
	placed := 0
	attempts := 0
	maxAttempts := count * 5 
	
	
	candidates := make([]*domainmodels.Cell, len(eligibleCells))
	copy(candidates, eligibleCells)
	
	for placed < count && attempts < maxAttempts && len(candidates) > 0 {
		attempts++
		
		
		candidateIndex := rng.Intn(len(candidates))
		candidate := candidates[candidateIndex]
		
		
		if gl.hasGoodFuelStationSpacing(grid, candidate) {
			
			candidate.CellType = domainmodels.CellTypeRefuel
			
			refuelAmount := 1000.0 + rng.Float64()*2000.0 
			candidate.RefuelAmount = &refuelAmount
			
			placed++
			
			candidates = gl.removeNearbyFromCandidates(candidates, candidate, 3)
		} else {
			
			candidates = append(candidates[:candidateIndex], candidates[candidateIndex+1:]...)
		}
	}
	
	if placed < count {
		fmt.Printf("Warning: Only placed %d of %d requested fuel stations\n", placed, count)
	}
	
	return nil
}


func (gl *GridLoader) hasGoodFuelStationSpacing(grid *domainmodels.Grid, candidate *domainmodels.Cell) bool {
	minSpacing := int64(4) 
	for _, cell := range grid.Cells {
		if cell.CellType == domainmodels.CellTypeRefuel {
			distance := utils.ManhattanDistance(candidate.Xpos, candidate.Ypos, cell.Xpos, cell.Ypos)
			if distance < minSpacing {
				return false 
			}
		}
	}
	
	return true
}


func (gl *GridLoader) placeDepots(grid *domainmodels.Grid, eligibleCells []*domainmodels.Cell, count int, rng *rand.Rand) error {
	placed := 0
	attempts := 0
	maxAttempts := count * 5
	
	candidates := make([]*domainmodels.Cell, len(eligibleCells))
	copy(candidates, eligibleCells)
	
	for placed < count && attempts < maxAttempts && len(candidates) > 0 {
		attempts++
		
		candidateIndex := rng.Intn(len(candidates))
		candidate := candidates[candidateIndex]
		
		if candidate.CellType == domainmodels.CellTypeNormal && 
		   gl.hasGoodDepotSpacing(grid, candidate) &&
		   utils.CountCellConnections(grid, candidate) >= 2 {
			
			candidate.CellType = domainmodels.CellTypeDepot
			placed++
			
			candidates = gl.removeNearbyFromCandidates(candidates, candidate, 5) 
		} else {
			candidates = append(candidates[:candidateIndex], candidates[candidateIndex+1:]...)
		}
	}
	
	if placed < count {
		fmt.Printf("Warning: Only placed %d of %d requested depots\n", placed, count)
	}
	
	return nil
}


func (gl *GridLoader) hasGoodDepotSpacing(grid *domainmodels.Grid, candidate *domainmodels.Cell) bool {
	minSpacing := int64(6) 
	
	for _, cell := range grid.Cells {
		if cell.CellType == domainmodels.CellTypeDepot {
			distance := utils.ManhattanDistance(candidate.Xpos, candidate.Ypos, cell.Xpos, cell.Ypos)
			if distance < minSpacing {
				return false
			}
		}
	}
	
	return true
}

func (gl *GridLoader) placeBlockedAreas(grid *domainmodels.Grid, eligibleCells []*domainmodels.Cell, count int, rng *rand.Rand) error {
	placed := 0
	attempts := 0
	maxAttempts := count * 3 
	
	candidates := make([]*domainmodels.Cell, len(eligibleCells))
	copy(candidates, eligibleCells)
	
	for placed < count && attempts < maxAttempts && len(candidates) > 0 {
		attempts++
		
		candidateIndex := rng.Intn(len(candidates))
		candidate := candidates[candidateIndex]
		
		if candidate.CellType == domainmodels.CellTypeNormal &&
		   gl.canSafelyBlockCell(grid, candidate) {
			
			
			originalSegments := candidate.RoadSegments
			candidate.CellType = domainmodels.CellTypeBlocked
			candidate.RoadSegments = []domainmodels.CellRoad{} 
			
			
			if gl.validateRoadConnectivity(grid) == nil {
				
				placed++
				candidates = gl.removeNearbyFromCandidates(candidates, candidate, 1)
			} else {
			
				candidate.CellType = domainmodels.CellTypeNormal
				candidate.RoadSegments = originalSegments
				
				candidates = append(candidates[:candidateIndex], candidates[candidateIndex+1:]...)
			}
		} else {
			candidates = append(candidates[:candidateIndex], candidates[candidateIndex+1:]...)
		}
	}
	
	if placed < count {
		fmt.Printf("Warning: Only placed %d of %d requested blocked areas (connectivity constraints)\n", placed, count)
	}
	
	return nil
}

func (gl *GridLoader) canSafelyBlockCell(grid *domainmodels.Grid, candidate *domainmodels.Cell) bool {
	
	connectionCount := utils.CountCellConnections(grid, candidate)
	if connectionCount > 2 {
		return false
	}
	
	minDistanceFromSpecial := int64(2)
	for _, cell := range grid.Cells {
		if cell.CellType == domainmodels.CellTypeRefuel || cell.CellType == domainmodels.CellTypeDepot {
			distance := utils.ManhattanDistance(candidate.Xpos, candidate.Ypos, cell.Xpos, cell.Ypos)
			if distance < minDistanceFromSpecial {
				return false
			}
		}
	}
	
	return true
}


func (gl *GridLoader) validatePostPlacementConnectivity(grid *domainmodels.Grid) error {
	return gl.validateRoadConnectivity(grid)
}




func (gl *GridLoader) removeNearbyFromCandidates(candidates []*domainmodels.Cell, placed *domainmodels.Cell, minDistance int64) []*domainmodels.Cell {
	var filtered []*domainmodels.Cell
	
	for _, candidate := range candidates {
		distance := utils.ManhattanDistance(candidate.Xpos, candidate.Ypos, placed.Xpos, placed.Ypos)
		if distance >= minDistance {
			filtered = append(filtered, candidate)
		}
	}
	
	return filtered
}