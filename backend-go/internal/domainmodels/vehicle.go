package domainmodels

import (
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
	VehicleStatusMoving    VehicleStatus = "moving"
	VehicleStatusIdle      VehicleStatus = "idle"
	VehicleStatusStopped   VehicleStatus = "stopped"
	VehicleStatusRefueling VehicleStatus = "refueling"
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

	CurrentCell     *Cell        `json:"current_cell,omitempty"`
	CurrentRoad     *RoadSegment `json:"current_road,omitempty"`
	Progress        float64      `json:"progress"`
	CurrentSpeedKPH float64      `json:"current_speed_kph"`
	PlannedPath     []int64      `json:"planned_path,omitempty"`     // sequence of RoadSegment IDs
	NextDecisionAt  int64        `json:"next_decision_at,omitempty"` // simulated

	OriginCell      *Cell `json:"origin_cell,omitempty"`
	DestinationCell *Cell `json:"destination_cell,omitempty"`

	FuelLevel    float64 `json:"fuel_level"`
	ProximityLOD bool    `json:"proximity_lod"`

	LastDecisionAt int64 `json:"last_decision_at,omitempty"`
}
