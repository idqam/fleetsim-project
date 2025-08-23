package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type VehicleProfileDB struct {
	ID                int64     `db:"id"`
	Name              string    `db:"name"`
	VehicleType       string    `db:"vehicle_type"`
	TankLiters        float64   `db:"tank_liters"`
	ConsumptionL100KM float64   `db:"consumption_l_per_100km"`
	MaxSpeedKPH       int       `db:"max_speed_kph"`
	CargoCapacityKG   float64   `db:"cargo_capacity_kg"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

type VehicleDB struct {
	ID               string     `db:"id"`
	VehicleClass     string     `db:"vehicle_class"`
	FleetID          *uuid.UUID `db:"fleet_id"`
	VehicleProfileID int64      `db:"vehicle_profile_id"`
	Status           string     `db:"status"`

	CurrentCellID *int64  `db:"current_cell_id"`    // FK -> cells.id
	CurrentSegID  *int64  `db:"current_segment_id"` // FK -> road_segments.id
	EdgeProgress  float64 `db:"edge_progress"`

	OriginCellID      *int64 `db:"origin_cell_id"`      // FK -> cells.id
	DestinationCellID *int64 `db:"destination_cell_id"` // FK -> cells.id

	CurrentSpeedKPH float64 `db:"current_speed_kph"`
	FuelLevel       float64 `db:"fuel_level"`
	ProximityLOD    bool    `db:"proximity_lod"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
