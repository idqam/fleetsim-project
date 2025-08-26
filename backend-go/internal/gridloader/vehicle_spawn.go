package gridloader

import (
	"fmt"
	"math/rand"

	"owenvi.com/fleetsim/internal/config"
	"owenvi.com/fleetsim/internal/constants"
	"owenvi.com/fleetsim/internal/domainmodels"
	"owenvi.com/fleetsim/internal/utils"

	"time"
)


type VehicleSpawner struct {
		config *config.SimulationConfig
	
	
	vehicleProfiles map[string]*domainmodels.VehicleProfile
	
	
	vehicleCounter int64
	spawnedVehicles []domainmodels.Vehicle
	
	
	rng *rand.Rand
}


func NewVehicleSpawner(config *config.SimulationConfig, seed int64) *VehicleSpawner {
	spawner := &VehicleSpawner{
		config:          config,
		vehicleProfiles: make(map[string]*domainmodels.VehicleProfile),
		vehicleCounter:  1,
		spawnedVehicles: make([]domainmodels.Vehicle, 0),
		rng:             rand.New(rand.NewSource(seed)),
	}
	

	spawner.initializeVehicleProfiles()
	
	return spawner
}

func (vs *VehicleSpawner) initializeVehicleProfiles() {
	
	carProfile := &domainmodels.VehicleProfile{
		ID:                1,
		Name:              "Standard Car",
		VehicleType:       constants.VehicleTypeCar,
		TankLiters:        60.0,  
		ConsumptionL100KM: 8.0,   
		MaxSpeedKPH:       120,   
		CargoCapacityKG:   500.0, 
	}
	vs.vehicleProfiles["car"] = carProfile
	
	vanProfile := &domainmodels.VehicleProfile{
		ID:                2,
		Name:              "Delivery Van",
		VehicleType:       constants.VehicleTypeVan,
		TankLiters:        80.0,  
		ConsumptionL100KM: 12.0,  
		MaxSpeedKPH:       100,   
		CargoCapacityKG:   1500.0,
	}
	vs.vehicleProfiles["van"] = vanProfile
	
	
	truckProfile := &domainmodels.VehicleProfile{
		ID:                3,
		Name:              "Heavy Truck",
		VehicleType:       constants.VehicleTypeTruck,
		TankLiters:        200.0,
		ConsumptionL100KM: 25.0, 
		MaxSpeedKPH:       80,   
		CargoCapacityKG:   8000.0,
	}
	vs.vehicleProfiles["truck"] = truckProfile
	
	fmt.Printf("Initialized %d vehicle profiles: car, van, truck\n", len(vs.vehicleProfiles))
}


func (vs *VehicleSpawner) SpawnRandomVehicles(grid *domainmodels.Grid, count int) ([]domainmodels.Vehicle, error) {
	fmt.Printf("Spawning %d random vehicles...\n", count)
	
	validSpawnPoints := vs.findValidSpawnLocations(grid)
	if len(validSpawnPoints) == 0 {
		return nil, fmt.Errorf("no valid spawn locations found in grid")
	}
	
	fmt.Printf("Found %d valid spawn locations\n", len(validSpawnPoints))
	
	var spawnedVehicles []domainmodels.Vehicle
	spawnAttempts := 0
	maxAttempts := count * 3
	
	for len(spawnedVehicles) < count && spawnAttempts < maxAttempts {
		spawnAttempts++
		
		spawnPoint := validSpawnPoints[vs.rng.Intn(len(validSpawnPoints))]
		vehicleType := vs.selectRandomVehicleType()
		
		
		vehicle := vs.createVehicle(vehicleType, spawnPoint)
		
		
		destination := vs.selectRandomDestination(grid, spawnPoint)
		if destination != nil {
			vehicle.DestinationCell = destination
		}
		
		
		spawnedVehicles = append(spawnedVehicles, vehicle)
		vs.spawnedVehicles = append(vs.spawnedVehicles, vehicle)
		
		fmt.Printf("Spawned %s '%s' at (%d,%d) -> (%d,%d)\n",
			vehicle.Profile.VehicleType,
			vehicle.ID,
			spawnPoint.Xpos, spawnPoint.Ypos,
			vehicle.DestinationCell.Xpos, vehicle.DestinationCell.Ypos)
	}
	
	if len(spawnedVehicles) < count {
		fmt.Printf("Warning: Only spawned %d of %d requested vehicles after %d attempts\n",
			len(spawnedVehicles), count, spawnAttempts)
	}
	
	return spawnedVehicles, nil
}


func (vs *VehicleSpawner) findValidSpawnLocations(grid *domainmodels.Grid) []*domainmodels.Cell {
	var validLocations []*domainmodels.Cell
	
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		
		
		if len(cell.RoadSegments) == 0 {
			continue 
		}
		
		if cell.CellType == domainmodels.CellTypeBlocked {
			continue 
		}
		
		if vs.isSuitableSpawnLocation(grid, cell) {
			validLocations = append(validLocations, cell)
		}
	}
	
	return validLocations
}

func (vs *VehicleSpawner) isSuitableSpawnLocation(grid *domainmodels.Grid, cell *domainmodels.Cell) bool {

	if cell.CellType == domainmodels.CellTypeDepot {
		return true
	}
	
	
	if cell.CellType == domainmodels.CellTypeNormal {
		
		connectionCount := utils.CountCellConnections(grid, cell)
		return connectionCount >= 1 
	}
	
	if cell.CellType == domainmodels.CellTypeRefuel {
		return true
	}
	
	return false
}


func (vs *VehicleSpawner) getConnectionDirection(dx, dy int64) string {
	if dx > 0 {
		return "east"
	} else if dx < 0 {
		return "west"
	} else if dy > 0 {
		return "south"
	} else if dy < 0 {
		return "north"
	}
	return "center"
}

func (vs *VehicleSpawner) selectRandomVehicleType() string {
	random := vs.rng.Float64()
	
	if random < vs.config.VehicleTypeDistribution["car"] {
		return "car"
	} else if random < vs.config.VehicleTypeDistribution["car"] + vs.config.VehicleTypeDistribution["van"] {
		return "van"
	} else {
		return "truck"
	}
}


func (vs *VehicleSpawner) createVehicle(vehicleType string, spawnLocation *domainmodels.Cell) domainmodels.Vehicle {

	profile := vs.vehicleProfiles[vehicleType]
	if profile == nil {

		profile = vs.vehicleProfiles["car"]
	}
	

	vehicleID := fmt.Sprintf("v%d_%s_%d", vs.vehicleCounter, vehicleType, time.Now().Unix()%10000)
	vs.vehicleCounter++
	

	fuelMin := vs.config.DefaultFuelRange[0]
	fuelMax := vs.config.DefaultFuelRange[1]
	initialFuelPercent := fuelMin + vs.rng.Float64()*(fuelMax-fuelMin)
	initialFuelAmount := profile.TankLiters * initialFuelPercent
	

	vehicle := domainmodels.Vehicle{
		ID:      vehicleID,
		Class:   constants.VehicleClassFleet, 
		Profile: *profile,
		Status:  constants.VehicleStatusIdle, 
		
		
		CurrentCell:     spawnLocation,
		OriginCell:     spawnLocation,
		Progress:       0.0, 
		CurrentSpeedKPH: 0.0,
		
		
		FuelLevel:       initialFuelAmount,
		SpeedMultiplier: 1.0 + (vs.rng.Float64()-0.5)*vs.config.DefaultSpeedVariation,
		ProximityLOD:    false, 
		
		SpawnedAt: &[]time.Time{time.Now()}[0],
		SpawnIntent: &domainmodels.UserSpawnIntent{
			RequestedVehicleType:   profile.VehicleType,
			RequestedSpawnLocation: "random",
			RequestedDestination:   "random",
			RequestedFuelPercent:   &initialFuelPercent,
			UserNotes:             &[]string{"Week 1 random spawn"}[0],
		},
	}
	
	return vehicle
}

func (vs *VehicleSpawner) selectRandomDestination(grid *domainmodels.Grid, origin *domainmodels.Cell) *domainmodels.Cell {
	var fuelStations []*domainmodels.Cell
	var depots []*domainmodels.Cell
	var normalCells []*domainmodels.Cell
	
	minDistance := int64(5) 
	
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		
		if len(cell.RoadSegments) == 0 {
			continue
		}
		
		distance := utils.ManhattanDistance(origin.Xpos, origin.Ypos, cell.Xpos, cell.Ypos)
		if distance < minDistance {
			continue
		}
		
		
		switch cell.CellType {
		case domainmodels.CellTypeRefuel:
			fuelStations = append(fuelStations, cell)
		case domainmodels.CellTypeDepot:
			depots = append(depots, cell)
		case domainmodels.CellTypeNormal:
			normalCells = append(normalCells, cell)
		}
	}
	

	random := vs.rng.Float64()
	
	if random < 0.4 && len(depots) > 0 {
		return depots[vs.rng.Intn(len(depots))]
	} else if random < 0.7 && len(fuelStations) > 0 {
		return fuelStations[vs.rng.Intn(len(fuelStations))]
	} else if len(normalCells) > 0 {
		return normalCells[vs.rng.Intn(len(normalCells))]
	}
	
	
	allAccessible := append(append(fuelStations, depots...), normalCells...)
	if len(allAccessible) > 0 {
		return allAccessible[vs.rng.Intn(len(allAccessible))]
	}
	
	return nil 
}


func (vs *VehicleSpawner) GetSpawnedVehicles() []domainmodels.Vehicle {
	return vs.spawnedVehicles
}


func (vs *VehicleSpawner) GetSpawnStatistics() SpawnStatistics {
	stats := SpawnStatistics{
		TotalVehiclesSpawned: len(vs.spawnedVehicles),
		VehicleTypeCounts:    make(map[string]int),
	}
	
	var totalFuel, totalSpeed float64
	
	for _, vehicle := range vs.spawnedVehicles {
		typeKey := string(vehicle.Profile.VehicleType)
		stats.VehicleTypeCounts[typeKey]++
		
		totalFuel += vehicle.FuelLevel
		totalSpeed += vehicle.SpeedMultiplier
	}
	
	
	if len(vs.spawnedVehicles) > 0 {
		stats.AverageFuelLevel = totalFuel / float64(len(vs.spawnedVehicles))
		stats.AverageSpeedMultiplier = totalSpeed / float64(len(vs.spawnedVehicles))
	}
	
	return stats
}

type SpawnStatistics struct {
	TotalVehiclesSpawned    int                `json:"total_vehicles_spawned"`
	VehicleTypeCounts       map[string]int     `json:"vehicle_type_counts"`
	AverageFuelLevel        float64            `json:"average_fuel_level"`
	AverageSpeedMultiplier  float64            `json:"average_speed_multiplier"`
}