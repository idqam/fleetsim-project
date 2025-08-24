package config

import "owenvi.com/fleetsim/internal/domainmodels"

type SimulationConfig struct {
	MaxActiveVehicles      int `json:"max_active_vehicles"`
	MaxVehiclesPerUser     int `json:"max_vehicles_per_user"`
	MaxSpawnRequestsPerMin int `json:"max_spawn_requests_per_min"`

	SimulationSpeedMultiplier float64 `json:"simulation_speed_multiplier"`
	TrafficUpdateInterval     int64   `json:"traffic_update_interval_ms"`
	VehicleCleanupInterval    int64   `json:"vehicle_cleanup_interval_ms"`

	DefaultFuelRange        [2]float64         `json:"default_fuel_range"`
	DefaultSpeedVariation   float64            `json:"default_speed_variation"`
	VehicleTypeDistribution map[string]float64 `json:"vehicle_type_distribution"`

	BaseRoadConditions         map[string]domainmodels.RoadCondition `json:"base_road_conditions"`
	RandomConditionProbability float64                               `json:"random_condition_probability"`
	ConditionDurationRange     [2]int64                              `json:"condition_duration_range"`
}

type VehicleSpawnRequest struct {
	VehicleType        string   `json:"vehicle_type"`
	InitialFuelPercent *float64 `json:"initial_fuel_percent,omitempty"`

	SpawnLocation *SpawnLocationRequest `json:"spawn_location,omitempty"`

	Destination *DestinationRequest `json:"destination,omitempty"`

	CustomSpeedMultiplier *float64 `json:"custom_speed_multiplier,omitempty"`
	VehicleName           *string  `json:"vehicle_name,omitempty"`
}

type SpawnLocationRequest struct {
	LocationType string `json:"location_type"`
	CellX        *int64 `json:"cell_x,omitempty"`
	CellY        *int64 `json:"cell_y,omitempty"`
}

type DestinationRequest struct {
	DestinationType string `json:"destination_type"`
	CellX           *int64 `json:"cell_x,omitempty"`
	CellY           *int64 `json:"cell_y,omitempty"`
}
