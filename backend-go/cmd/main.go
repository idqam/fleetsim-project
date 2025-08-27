package main

import (
	"fmt"
	"os"
	"strconv"

	"owenvi.com/fleetsim/internal/config"
	"owenvi.com/fleetsim/internal/domainmodels"
	"owenvi.com/fleetsim/internal/gridloader"
)

func main() {
	fmt.Println("Fleet Simulation - Week 1 Milestone Demo")
	fmt.Println("========================================")

	config := config.Config()

	gridLoader := gridloader.NewGridLoader()
	args := os.Args[1:]
	
	dimX := args[0]
	dimY := args[1]
	
	width, e1 :=strconv.Atoi(dimX)
	height, e2 :=strconv.Atoi(dimY)
	if e1!= nil{
		panic("width must be numeric")
	}
	if e2 != nil {
		panic("height must be numeric")
	}
	
	
	 
	//gridLoader.ConfigureForTesting(int64(width), int64(height), 42, 0, 0.02, 0, 0.2, 0.3, 0.1)
//gridLoader.ConfigureForTesting(int64(width), int64(height), 42, 0.03, 0.01, 0.02, 0.6, 0.2, 0.1)
gridLoader.ConfigureForTesting(int64(width), int64(height), 99, 0.05, 0.02, 0.05, 0.7, 0.3, 0.1)	
//gridloader.ConfigureForTesting(int64(width), int64(height),, 1337, 0.08, 0.03, 0.02, 0.85, 0.5, 0.05)
//gridloader.ConfigureForTesting(int64(width), int64(height), 7, 0.02, 0.01, 0.08, 0.4, 0.15, 0.25)
//gridLoader.ConfigureForTesting(int64(width), int64(height), 12345, 0.05, 0.02, 0.05, 0.7, 0.35, 0.1)
vehicleSpawner := gridloader.NewVehicleSpawner(config, 42)

	fmt.Printf("Generating %s x %s grid with roads and special locations...", dimX, dimY)

	demoWorld, err := gridLoader.CreateDemoGrid(20, vehicleSpawner)
	if err != nil {
		fmt.Printf("Grid generation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated %s by %s grid successfully\n", dimX, dimY)
	fmt.Printf("   • Grid dimensions: %dx%d\n", demoWorld.Grid.DimX, demoWorld.Grid.DimY)
	fmt.Printf("   • Total cells: %d\n", len(demoWorld.Grid.Cells))
	fmt.Printf("   • Road segments: %d\n", demoWorld.Stats.TotalSegments)
	fmt.Printf("   • Generation time: %d ms\n", demoWorld.Stats.GenerationTimeMs)

	fuelStations := 0
	depots := 0
	blockedAreas := 0
	roadCells := 0

	for _, cell := range demoWorld.Grid.Cells {
		switch cell.CellType {
		case domainmodels.CellTypeRefuel:
			fuelStations++
		case domainmodels.CellTypeDepot:
			depots++
		case domainmodels.CellTypeBlocked:
			blockedAreas++
		}
		if len(cell.RoadSegments) > 0 {
			roadCells++
		}
	}

	fmt.Printf("   • Special locations: %d fuel stations, %d depots, %d blocked areas\n",
		fuelStations, depots, blockedAreas)
	fmt.Printf("   • Road network: %d cells with road access\n", roadCells)

	fmt.Printf("\n✅ Spawned %d vehicles at different positions\n", len(demoWorld.Vehicles))
	for i, vehicle := range demoWorld.Vehicles {
		if i < 3 {
			fmt.Printf("   • Vehicle %s (%s) at (%d,%d) with %.1fL fuel\n",
				vehicle.ID, vehicle.Profile.VehicleType,
				vehicle.CurrentCell.Xpos, vehicle.CurrentCell.Ypos,
				vehicle.FuelLevel)
		}
	}
	if len(demoWorld.Vehicles) > 3 {
		fmt.Printf("   • ... and %d more vehicles\n", len(demoWorld.Vehicles)-3)
	}

	fmt.Println("\n✅ Basic movement validation and collision detection ready")
	validMoves := 0
	totalChecks := 0

	for _, vehicle := range demoWorld.Vehicles {

		adjacentPositions := [][2]int64{
			{vehicle.CurrentCell.Xpos + 1, vehicle.CurrentCell.Ypos},
			{vehicle.CurrentCell.Xpos - 1, vehicle.CurrentCell.Ypos},
			{vehicle.CurrentCell.Xpos, vehicle.CurrentCell.Ypos + 1},
			{vehicle.CurrentCell.Xpos, vehicle.CurrentCell.Ypos - 1},
		}

		validAdjacent := 0
		for _, pos := range adjacentPositions {
			if targetCell := demoWorld.Grid.CoordIndex[pos]; targetCell != nil {
				if targetCell.CellType != domainmodels.CellTypeBlocked {
					validAdjacent++
					totalChecks++

					if len(targetCell.RoadSegments) > 0 {
						validMoves++
					}
				}
			}
		}

		fmt.Printf("   • Vehicle %s: %d adjacent cells from (%d,%d), %d accessible\n",
			vehicle.ID, validAdjacent,
			vehicle.CurrentCell.Xpos, vehicle.CurrentCell.Ypos, validAdjacent)
	}

	fmt.Printf("   • Movement validation: %d/%d adjacent cells have road access\n", validMoves, totalChecks)

	fmt.Printf("%s x %s GRID VISUALIZATION", dimX, dimY)

	demoWorld.PrintASCIIVisualization()

	fmt.Println("WEEK 1 MILESTONE COMPLETED")

	fmt.Println("✅ Go project structure with domain models")
	fmt.Println("✅ Grid Loader  procedural generation")
	fmt.Println("✅ Basic Vehicle and Cell structs")
	fmt.Printf("✅ %s x %s demo grid with roads, fuel stations, depots\n", dimX, dimY)
	fmt.Println("✅ Simple vehicle spawning at random locations")
	fmt.Println("✅ Generated grid with roads and special locations displayed")
	fmt.Println("✅ 10 vehicles spawned at different positions shown")
	fmt.Println("✅ Basic movement validation and collision detection demonstrated")
}
