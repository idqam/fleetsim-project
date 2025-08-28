package movement

import (
	"fmt"

	"owenvi.com/fleetsim/internal/constants"
	"owenvi.com/fleetsim/internal/domainmodels"
)

type VehicleLifecycleManager struct {
	grid     *domainmodels.Grid
	vehicles map[string]*domainmodels.Vehicle
}

func NewVehicleLifecycleManager(grid *domainmodels.Grid, vehicles []domainmodels.Vehicle) *VehicleLifecycleManager {
    vehicleMap := make(map[string]*domainmodels.Vehicle)
    
    for i := range vehicles {
        vehicleCopy := vehicles[i]
        vehicle := &vehicleCopy
        
        
        fmt.Printf("Debug - Vehicle %s: CurrentCell=%v, DestinationCell=%v\n", 
            vehicle.ID, 
            vehicle.CurrentCell != nil,
            vehicle.DestinationCell != nil)

        
        if vehicle.CurrentCell == nil {
            fmt.Printf("Warning: Vehicle %s has no current cell\n", vehicle.ID)
            continue
        }
        
        if vehicle.DestinationCell == nil {
            fmt.Printf("Warning: Vehicle %s has no destination cell\n", vehicle.ID)
            continue
        }
        
        if len(vehicle.CurrentCell.RoadSegments) > 0 {
            vehicle.CurrentSegment = &vehicle.CurrentCell.RoadSegments[0].RoadSegment
            vehicle.SegmentProgress = 0.0
            vehicle.Status = constants.VehicleStatusMoving
            fmt.Printf("Successfully initialized vehicle %s\n", vehicle.ID)
        } else {
            fmt.Printf("Warning: Vehicle %s has no road segments at (%d,%d)\n", 
                vehicle.ID, vehicle.CurrentCell.Xpos, vehicle.CurrentCell.Ypos)
        }
        vehicleMap[vehicle.ID] = vehicle
    }
    
    fmt.Printf("Initialized %d out of %d vehicles\n", len(vehicleMap), len(vehicles))
    
    return &VehicleLifecycleManager{
        grid:     grid,
        vehicles: vehicleMap,
    }
}

func (vlm *VehicleLifecycleManager) UpdateAllVehicles(timeStepSeconds float64) {
	for _, vehicle := range vlm.vehicles {
		if vehicle.Status == constants.VehicleStatusMoving {
			vlm.updateSingleVehicle(vehicle, timeStepSeconds)
		}
	}
}

func (vlm *VehicleLifecycleManager) updateSingleVehicle(vehicle *domainmodels.Vehicle, timeStepSeconds float64) {
    if vehicle.CurrentSegment == nil {
        fmt.Printf("Vehicle %s has no current segment\n", vehicle.ID)
        vehicle.Status = constants.VehicleStatusCompleted
        return
    }
    
    result := vehicle.UpdatePosition(timeStepSeconds, vlm.grid)
    
    if result.ReachedDestination {
        fmt.Printf("Vehicle %s reached destination!\n", vehicle.ID)
        return
    }
    
    if result.ReachedSegmentEnd {
        vehicle.Status = constants.VehicleStatusCompleted
        fmt.Printf("Vehicle %s reached end of segment\n", vehicle.ID)
    }
}

func (vlm *VehicleLifecycleManager) GetActiveVehicles() []*domainmodels.Vehicle {
	var active []*domainmodels.Vehicle
	for _, vehicle := range vlm.vehicles {
		if vehicle.Status == constants.VehicleStatusMoving {
			active = append(active, vehicle)
		}
	}
	return active
}

func (vlm *VehicleLifecycleManager) PrintCurrentState() {
	
	fmt.Println("Current vehicle positions:")
	for _, vehicle := range vlm.vehicles {
		if vehicle.Status == constants.VehicleStatusMoving && vehicle.CurrentCell != nil {
			fmt.Printf("  %s at (%d,%d) progress: %.1f%%\n",
				vehicle.ID, vehicle.CurrentCell.Xpos, vehicle.CurrentCell.Ypos,
				vehicle.SegmentProgress*100)
		}
	}
}
