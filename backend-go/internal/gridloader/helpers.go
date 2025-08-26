package gridloader

import (
	"fmt"

	"owenvi.com/fleetsim/internal/constants"
	"owenvi.com/fleetsim/internal/domainmodels"
)


func (gl *GridLoader) findValidSpawnLocations(grid *domainmodels.Grid) []*domainmodels.Cell {
	var validLocations []*domainmodels.Cell
	
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		
		if len(cell.RoadSegments) == 0 {
			continue 
		}
		
		if cell.CellType == domainmodels.CellTypeBlocked {
			continue 
		}
		
		
		if gl.isSuitableForSpawning(cell) {
			validLocations = append(validLocations, cell)
		}
	}
	
	return validLocations
}


func (gl *GridLoader) isSuitableForSpawning(cell *domainmodels.Cell) bool {

	if cell.CellType == domainmodels.CellTypeDepot {
		return true
	}
	
	
	if cell.CellType == domainmodels.CellTypeNormal && len(cell.RoadSegments) > 0 {
		return true
	}
	
	
	if cell.CellType == domainmodels.CellTypeRefuel {
		return true
	}
	
	return false
}


func (gl *GridLoader) CreateDemoGrid(vehicleCount int, spawner *VehicleSpawner) (*DemoWorld, error) {
	
	grid, err := gl.GenerateProcedural()
	if err != nil {
		return nil, err
	}
	
	
	vehicles, err := spawner.SpawnRandomVehicles(grid, vehicleCount)
	if err != nil {
		return nil, err
	}
	
	
	return &DemoWorld{
		Grid:     grid,
		Vehicles: vehicles,
		Stats:    gl.GetGenerationStats(),
	}, nil
}


type DemoWorld struct {
	Grid     *domainmodels.Grid           `json:"grid"`
	Vehicles []domainmodels.Vehicle       `json:"vehicles"`
	Stats    *GenerationStats             `json:"stats"`
}

func (world *DemoWorld) PrintASCIIVisualization() {
	grid := world.Grid
	
	fmt.Printf("\n=== Demo World Visualization ===\n")
	fmt.Printf("Grid: %dx%d | Vehicles: %d | Segments: %d\n\n",
		grid.DimX, grid.DimY, len(world.Vehicles), world.Stats.TotalSegments)
	
	
	display := make([][]rune, grid.DimY)
	for i := range display {
		display[i] = make([]rune, grid.DimX)
		for j := range display[i] {
			display[i][j] = '.'
		}
	}
	
	
	for _, cell := range grid.Cells {
		x, y := cell.Xpos, cell.Ypos
		
		
		var symbol rune
		switch cell.CellType {
		case domainmodels.CellTypeNormal:
			if len(cell.RoadSegments) > 0 {
				symbol = '─' 
			} else {
				symbol = '.' 
			}
		case domainmodels.CellTypeRefuel:
			symbol = 'F' 
		case domainmodels.CellTypeDepot:
			symbol = 'D' 
		case domainmodels.CellTypeBlocked:
			symbol = 'X' 
		default:
			symbol = '?'
		}
		
		display[y][x] = symbol
	}
	
	for _, vehicle := range world.Vehicles {
		if vehicle.CurrentCell != nil {
			x, y := vehicle.CurrentCell.Xpos, vehicle.CurrentCell.Ypos
			
			switch vehicle.Profile.VehicleType {
			case constants.VehicleTypeCar:
				display[y][x] = 'c' 
			case constants.VehicleTypeVan:
				display[y][x] = 'v' 
			case constants.VehicleTypeTruck:
				display[y][x] = 't' 
			}
		}
	}
	fmt.Print("   ")
	for x := int64(0); x < grid.DimX; x++ {
		fmt.Printf("%2d", x%10)
	}
	fmt.Println()
	
	
	for y := int64(0); y < grid.DimY; y++ {
		fmt.Printf("%2d ", y)
		for x := int64(0); x < grid.DimX; x++ {
			fmt.Printf(" %c", display[y][x])
		}
		fmt.Println()
	}
	
	
	fmt.Printf("\nLegend: . = empty, ─ = road, F = fuel, D = depot, X = blocked\n")
	fmt.Printf("Vehicles: c = car, v = van, t = truck\n")
	fmt.Printf("=================================\n\n")
}

func (world *DemoWorld) PrintDetailedStats() {
	fmt.Printf("=== Detailed World Statistics ===\n")
	
	
	totalCells := len(world.Grid.Cells)
	roadCells := 0
	specialCells := 0
	
	cellTypeCounts := make(map[domainmodels.CellType]int)
	
	for _, cell := range world.Grid.Cells {
		cellTypeCounts[cell.CellType]++
		
		if len(cell.RoadSegments) > 0 {
			roadCells++
		}
		
		if cell.CellType != domainmodels.CellTypeNormal {
			specialCells++
		}
	}
	
	fmt.Printf("Grid Analysis:\n")
	fmt.Printf("  Total cells: %d\n", totalCells)
	fmt.Printf("  Road cells: %d (%.1f%%)\n", roadCells, float64(roadCells)/float64(totalCells)*100)
	fmt.Printf("  Special cells: %d (%.1f%%)\n", specialCells, float64(specialCells)/float64(totalCells)*100)
	
	for cellType, count := range cellTypeCounts {
		fmt.Printf("    %s: %d\n", cellType, count)
	}
	
	// Vehicle statistics
	vehicleTypeCounts := make(map[constants.VehicleType]int)
	totalFuel := 0.0
	
	for _, vehicle := range world.Vehicles {
		vehicleTypeCounts[constants.VehicleType(vehicle.Profile.VehicleType)]++
		totalFuel += vehicle.FuelLevel
	}
	
	fmt.Printf("\nVehicle Analysis:\n")
	fmt.Printf("  Total vehicles: %d\n", len(world.Vehicles))
	for vType, count := range vehicleTypeCounts {
		fmt.Printf("    %s: %d\n", vType, count)
	}
	
	if len(world.Vehicles) > 0 {
		avgFuel := totalFuel / float64(len(world.Vehicles))
		fmt.Printf("  Average fuel level: %.1f liters\n", avgFuel)
	}
	
	// Network connectivity
	fmt.Printf("\nNetwork Analysis:\n")
	fmt.Printf("  Total road segments: %d\n", world.Stats.TotalSegments)
	fmt.Printf("  Generation time: %d ms\n", world.Stats.GenerationTimeMs)
	
	fmt.Printf("==================================\n\n")
}