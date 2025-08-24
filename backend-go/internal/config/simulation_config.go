package config

import "owenvi.com/fleetsim/internal/domainmodels"


func DefaultSimulationConfig() *SimulationConfig {
	return &SimulationConfig{
		MaxActiveVehicles:       50,
		MaxVehiclesPerUser:      10,
		MaxSpawnRequestsPerMin:  30,
		SimulationSpeedMultiplier: 2.0, 
		TrafficUpdateInterval:     2000,
		VehicleCleanupInterval:    30000,
		DefaultFuelRange:          [2]float64{0.3, 0.9},
		DefaultSpeedVariation:     0.1,
		VehicleTypeDistribution: map[string]float64{
			"car":   0.6,
			"van":   0.3,
			"truck": 0.1,
		},
		BaseRoadConditions: map[string]domainmodels.RoadCondition{
			"urban_street": {
				ID: "urban_street",
				Name: "Urban Street",
				Description: "Standard city street with moderate traffic",
				SpeedMultiplier: 0.8,
				FuelMultiplier: 1.1, 
				VisualColor: "#666666",
				Severity: "minor",
			},
			"highway": {
				ID: "highway",
				Name: "Highway",
				Description: "High-speed arterial road",
				SpeedMultiplier: 1.2,
				FuelMultiplier: 0.9, 
				VisualColor: "#4A90E2",
				Severity: "minor",
			},
		},
		RandomConditionProbability: 0.05,
		ConditionDurationRange:     [2]int64{30, 300},
	}
}