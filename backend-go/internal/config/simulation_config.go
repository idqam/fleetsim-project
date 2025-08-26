package config

import (
	"fmt"

	"owenvi.com/fleetsim/internal/domainmodels"
)

type SimulationConfig struct {
	MaxActiveVehicles       int     `json:"max_active_vehicles"`     
	MaxSpawnRequestsPerMin  int     `json:"max_spawn_requests_per_min"` 
	
	SimulationSpeedMultiplier float64 `json:"simulation_speed_multiplier"` 
	TrafficUpdateInterval     int64   `json:"traffic_update_interval_ms"`  
	VehicleCleanupInterval    int64   `json:"vehicle_cleanup_interval_ms"` 
	MovementUpdateInterval    int64   `json:"movement_update_interval_ms"`
	
	DefaultFuelRange         [2]float64         `json:"default_fuel_range"`     
	DefaultSpeedVariation    float64            `json:"default_speed_variation"`
	VehicleTypeDistribution  map[string]float64 `json:"vehicle_type_distribution"` 
	
	DefaultSpawnLocation     string             `json:"default_spawn_location"`
	DefaultDestinationType   string             `json:"default_destination_type"`
	PreferEdgeSpawn         bool               `json:"prefer_edge_spawn"`
	AvoidCongestion         bool               `json:"avoid_congestion"`
	MinDistanceFromOthers   int64              `json:"min_distance_from_others"`
	RequireFuelStop         bool               `json:"require_fuel_stop"`
	
	BaseRoadConditions map[string]domainmodels.RoadCondition `json:"base_road_conditions"`
	RandomConditionProbability float64 `json:"random_condition_probability"`
	ConditionDurationRange     [2]int64 `json:"condition_duration_range"`   
	
	RedisCommandBudgetPerHour int `json:"redis_command_budget_per_hour"` 
	TimescaleRetentionHours   int `json:"timescale_retention_hours"`     
}




func Config() *SimulationConfig {
	return &SimulationConfig{
		MaxActiveVehicles:       50,  
		MaxSpawnRequestsPerMin:  30,  
		
		SimulationSpeedMultiplier: 2.0,  
		TrafficUpdateInterval:     2000, 
		VehicleCleanupInterval:    30000,
		MovementUpdateInterval:    800,  
		
		DefaultFuelRange:     [2]float64{0.3, 0.9}, 
		DefaultSpeedVariation: 0.1, 
		
		VehicleTypeDistribution: map[string]float64{
			"car":   0.6,  
			"van":   0.3,  
			"truck": 0.1,  
		},
		
		DefaultSpawnLocation:     "random",
		DefaultDestinationType:   "random", 
		PreferEdgeSpawn:         false,
		AvoidCongestion:         true,
		MinDistanceFromOthers:   2,
		RequireFuelStop:         false,
		
		BaseRoadConditions: map[string]domainmodels.RoadCondition{
			"urban_street": {
				ID:              "urban_street",
				Name:            "Urban Street",
				Description:     "City street with moderate traffic patterns",
				SpeedMultiplier: 0.8,
				FuelMultiplier:  1.1, 
				IsTemporary:     false,
				Severity:        "minor",
				VisualColor:     "#666666",
				VisualPattern:   "solid",
			},
			"arterial_road": {
				ID:              "arterial_road", 
				Name:            "Arterial Road",
				Description:     "Main thoroughfare with higher speed limits",
				SpeedMultiplier: 1.2,
				FuelMultiplier:  0.9, 
				IsTemporary:     false,
				Severity:        "minor",
				VisualColor:     "#4A90E2",
				VisualPattern:   "solid",
			},
		},
		
		RandomConditionProbability: 0.05, 
		ConditionDurationRange:     [2]int64{60, 300}, 
		
		RedisCommandBudgetPerHour: 7000,  
		TimescaleRetentionHours:   2,     
	}
}

func (config *SimulationConfig) ValidateConfig() error {
	total := 0.0
	for _, ratio := range config.VehicleTypeDistribution {
		if ratio < 0.0 || ratio > 1.0 {
			return fmt.Errorf("vehicle type ratio must be between 0.0 and 1.0, got %.2f", ratio)
		}
		total += ratio
	}
	
	if total < 0.95 || total > 1.05 {
		return fmt.Errorf("vehicle type distribution ratios must sum to 1.0, got %.2f", total)
	}
	
	if config.DefaultFuelRange[0] < 0.0 || config.DefaultFuelRange[1] > 1.0 {
		return fmt.Errorf("fuel range must be between 0.0 and 1.0")
	}
	if config.DefaultFuelRange[0] >= config.DefaultFuelRange[1] {
		return fmt.Errorf("fuel range minimum must be less than maximum")
	}
	
	if config.TrafficUpdateInterval <= 0 || config.MovementUpdateInterval <= 0 {
		return fmt.Errorf("update intervals must be positive")
	}
	
	return nil
}