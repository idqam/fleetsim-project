package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type VehicleType string

const (
	VehicleTypeTruck VehicleType = "truck"
	VehicleTypeVan   VehicleType = "van"
	VehicleTypeCar   VehicleType = "car"
)

type VehicleClass string

const (
	VehicleClassFleet      VehicleClass = "fleet"
	VehicleClassBackground VehicleClass = "background"
)

type VehicleStatus string

const (
	VehicleStatusRequested VehicleStatus = "requested"
	VehicleStatusQueued    VehicleStatus = "queued"
	VehicleStatusSpawning  VehicleStatus = "spawning"

	VehicleStatusMoving    VehicleStatus = "moving"
	VehicleStatusIdle      VehicleStatus = "idle"
	VehicleStatusStopped   VehicleStatus = "stopped"
	VehicleStatusRefueling VehicleStatus = "refueling"

	VehicleStatusCompleted VehicleStatus = "completed"
	VehicleStatusRemoved   VehicleStatus = "removed"
	VehicleStatusFailed    VehicleStatus = "failed"
)

type VehicleProfile struct {
	ID                int64       `json:"id"`
	Name              string      `json:"name"`
	VehicleType       VehicleType `json:"vehicle_type"`
	TankLiters        float64     `json:"tank_liters"`
	ConsumptionL100KM float64     `json:"consumption_l_per_100km"`
	MaxSpeedKPH       int         `json:"max_speed_kph"`
	CargoCapacityKG   float64     `json:"cargo_capacity_kg"`
}

type Vehicle struct {
	ID      string         `json:"id"`
	Class   VehicleClass   `json:"class"`
	FleetID *uuid.UUID     `json:"fleet_id,omitempty"`
	Profile VehicleProfile `json:"profile"`
	Status  VehicleStatus  `json:"status"`

	SpawnRequestID *string    `json:"spawn_request_id,omitempty"`
	UserSessionID  *string    `json:"user_session_id,omitempty"`
	CustomName     *string    `json:"custom_name,omitempty"`
	SpawnedAt      *time.Time `json:"spawned_at,omitempty"`

	CurrentCell     *Cell        `json:"current_cell,omitempty"`
	CurrentRoad     *RoadSegment `json:"current_road,omitempty"`
	Progress        float64      `json:"progress"`
	CurrentSpeedKPH float64      `json:"current_speed_kph"`
	PlannedPath     []int64      `json:"planned_path,omitempty"`
	NextDecisionAt  int64        `json:"next_decision_at,omitempty"`

	OriginCell      *Cell            `json:"origin_cell,omitempty"`
	DestinationCell *Cell            `json:"destination_cell,omitempty"`
	SpawnIntent     *UserSpawnIntent `json:"spawn_intent,omitempty"`

	FuelLevel          float64  `json:"fuel_level"`
	InitialFuelPercent *float64 `json:"initial_fuel_percent,omitempty"`
	SpeedMultiplier    float64  `json:"speed_multiplier"`
	ProximityLOD       bool     `json:"proximity_lod"`

	LastDecisionAt int64   `json:"last_decision_at,omitempty"`
	FailureReason  *string `json:"failure_reason,omitempty"`
}

type UserSpawnIntent struct {
	RequestedVehicleType   VehicleType `json:"requested_vehicle_type"`
	RequestedSpawnLocation string      `json:"requested_spawn_location"`
	RequestedDestination   string      `json:"requested_destination"`
	RequestedFuelPercent   *float64    `json:"requested_fuel_percent,omitempty"`
	UserNotes              *string     `json:"user_notes,omitempty"`
}
