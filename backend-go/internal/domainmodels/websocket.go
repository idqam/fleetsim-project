package domainmodels

import "time"

type WSMessageType string

const (
	WSMsgVehiclePosition WSMessageType = "vehicle_position"
	WSMsgTrafficUpdate   WSMessageType = "traffic_update"
	WSMsgSimulationStart WSMessageType = "simulation_start"
	WSMsgSimulationStop  WSMessageType = "simulation_stop"
	WSMsgSimulationError WSMessageType = "simulation_error"
	WSMsgRoutingDecision WSMessageType = "routing_decision"
)

type WebSocketMessage struct {
	Type      WSMessageType `json:"type"`
	Timestamp time.Time     `json:"timestamp"`
	Data      interface{}   `json:"data"`
}

type RoutingDecisionUpdate struct {
	VehicleID   string `json:"vehicle_id"`
	FromSegment int64  `json:"from_segment"`
	ToSegment   int64  `json:"to_segment"`
	Reason      string `json:"reason,omitempty"`
}

type VehiclePositionUpdate struct {
	VehicleID     string        `json:"vehicle_id"`
	Xpos          int64         `json:"xpos"`
	Ypos          int64         `json:"ypos"`
	SpeedKPH      float64       `json:"speed_kph"`
	Status        VehicleStatus `json:"status"`
	RoadSegmentID *int64        `json:"road_segment_id,omitempty"`
	EdgeProgress  float64       `json:"edge_progress"`
	FuelLevel     float64       `json:"fuel_level"`
}

type TrafficUpdate struct {
	RoadSegmentID   int64   `json:"road_segment_id"`
	FleetCount      int     `json:"fleet_count"`
	BackgroundCount int     `json:"background_count"`
	Capacity        int     `json:"capacity"`
	CongestionRatio float64 `json:"congestion_ratio"` // fleet+background / capacity
}
