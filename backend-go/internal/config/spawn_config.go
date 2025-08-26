package config

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
