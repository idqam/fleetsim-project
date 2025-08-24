package domainmodels

import (
	"time"
)


type SpawnRequestStatus string

const (
	SpawnRequestStatusPending    SpawnRequestStatus = "pending"    
	SpawnRequestStatusValidating SpawnRequestStatus = "validating" 
	SpawnRequestStatusQueued     SpawnRequestStatus = "queued"      
	SpawnRequestStatusProcessing SpawnRequestStatus = "processing"  
	SpawnRequestStatusCompleted  SpawnRequestStatus = "completed"   
	SpawnRequestStatusFailed     SpawnRequestStatus = "failed"      
	SpawnRequestStatusCancelled  SpawnRequestStatus = "cancelled"   
)

type VehicleSpawnRequest struct {

	RequestID       string                `json:"request_id"`
	UserSessionID   string                `json:"user_session_id"`
	Status          SpawnRequestStatus    `json:"status"`
	CreatedAt       time.Time            `json:"created_at"`
	ProcessedAt     *time.Time           `json:"processed_at,omitempty"`
	

	VehicleType         VehicleType      `json:"vehicle_type"`
	CustomName          *string          `json:"custom_name,omitempty"`
	InitialFuelPercent  *float64         `json:"initial_fuel_percent,omitempty"`
	SpeedMultiplier     *float64         `json:"speed_multiplier,omitempty"`
	

	SpawnLocation       SpawnLocationRequest `json:"spawn_location"`
	Destination         DestinationRequest   `json:"destination"`
	

	UserNotes           *string          `json:"user_notes,omitempty"`
	Priority            int              `json:"priority"`               
	

	SpawnedVehicleID    *string          `json:"spawned_vehicle_id,omitempty"`
	FailureReason       *string          `json:"failure_reason,omitempty"`
	ValidationErrors    []string         `json:"validation_errors,omitempty"`
}

type SpawnLocationRequest struct {
	LocationType string `json:"location_type"`
	CellX        *int64 `json:"cell_x,omitempty"`
	CellY        *int64 `json:"cell_y,omitempty"`
	
	PreferEdgeSpawn    bool     `json:"prefer_edge_spawn"`   
	AvoidCongestion    bool     `json:"avoid_congestion"`      
	MinDistanceFromOthers int64 `json:"min_distance_from_others"` 
}

type DestinationRequest struct {
	DestinationType string `json:"destination_type"` 
	CellX           *int64 `json:"cell_x,omitempty"`
	CellY           *int64 `json:"cell_y,omitempty"`
	
	MaxDistanceFromSpawn *int64  `json:"max_distance_from_spawn,omitempty"` 
	RequireFuelStop      bool    `json:"require_fuel_stop"`                 
}