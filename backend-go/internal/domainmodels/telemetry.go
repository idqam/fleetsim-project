package domainmodels

import (
	"time"

	"github.com/google/uuid"
)

type TelemetryEvent struct {
	EventID         string    `json:"event_id"`
	SimulationRunID uuid.UUID `json:"simulation_run_id"`
	Timestamp       time.Time `json:"timestamp"`
	EventType       string    `json:"event_type"`

	VehicleID    *string        `json:"vehicle_id,omitempty"`
	CellX        *int64         `json:"cell_x,omitempty"`
	CellY        *int64         `json:"cell_y,omitempty"`
	SpeedKPH     *float64       `json:"speed_kph,omitempty"`
	FuelLevel    *float64       `json:"fuel_level,omitempty"`
	CurrentSegID *int64         `json:"current_segment_id,omitempty"`
	EdgeProgress *float64       `json:"edge_progress,omitempty"`
	Status       *VehicleStatus `json:"status,omitempty"`

	SegmentID       *int64 `json:"segment_id,omitempty"`
	FleetCount      *int   `json:"fleet_count,omitempty"`
	BackgroundCount *int   `json:"background_count,omitempty"`
	Capacity        *int   `json:"capacity,omitempty"`

	UserSessionID  *string `json:"user_session_id,omitempty"`
	SpawnRequestID *string `json:"spawn_request_id,omitempty"`
	UserAction     *string `json:"user_action,omitempty"`
	ActionResult   *string `json:"action_result,omitempty"`

	ConditionID     *string  `json:"condition_id,omitempty"`
	ConditionChange *string  `json:"condition_change,omitempty"`
	SpeedMultiplier *float64 `json:"speed_multiplier,omitempty"`
}
type VehicleState struct {
	VehicleID     string        `json:"vehicle_id"`
	Timestamp     time.Time     `json:"timestamp"`
	CellX         int64         `json:"cell_x"`
	CellY         int64         `json:"cell_y"`
	SpeedKPH      float64       `json:"speed_kph"`
	FuelLevel     float64       `json:"fuel_level"`
	CurrentSegID  *int64        `json:"current_segment_id"`
	EdgeProgress  float64       `json:"edge_progress"`
	Status        VehicleStatus `json:"status"`
	SimulatedTime int64         `json:"simulated_time_ms"`
}

type RoadLoadEvent struct {
	SegmentID       int64 `json:"segment_id"`
	FleetCount      int   `json:"fleet_count"`
	BackgroundCount int   `json:"background_count"`
	Capacity        int   `json:"capacity"`
}

type RoutingDecisionEvent struct {
	EventID         string    `json:"event_id"`
	SimulationRunID uuid.UUID `json:"simulation_run_id"`
	Timestamp       time.Time `json:"timestamp"`

	VehicleID    string `json:"vehicle_id"`
	FromSegment  int64  `json:"from_segment"`
	ToSegment    int64  `json:"to_segment"`
	DecisionType string `json:"decision_type"`
	Reason       string `json:"reason,omitempty"`
}
