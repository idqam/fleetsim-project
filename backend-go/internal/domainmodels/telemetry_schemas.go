package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type TelemetryEventDB struct {
	ID              int64     `db:"id"` // PK
	SimulationRunID uuid.UUID `db:"simulation_run_id"`
	Timestamp       time.Time `db:"timestamp"`  // hypertable partition key
	EventType       string    `db:"event_type"` // "position", "fuel", "traffic_load"

	VehicleID *string `db:"vehicle_id"` // NULL for road load events
	SegmentID *int64  `db:"segment_id"` // NULL for vehicle fuel events

	// Metrics (flatten for Timescale efficiency)
	Lat             *float64 `db:"lat"`
	Lng             *float64 `db:"lng"`
	SpeedKPH        *float64 `db:"speed_kph"`
	FuelLevel       *float64 `db:"fuel_level"`
	EdgeProgress    *float64 `db:"edge_progress"`
	Status          *string  `db:"status"`
	FleetCount      *int     `db:"fleet_count"`
	BackgroundCount *int     `db:"background_count"`
	Capacity        *int     `db:"capacity"`

	CreatedAt time.Time `db:"created_at"`
}
