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

	WSMsgSpawnRequest     WSMessageType = "spawn_request"
	WSMsgSpawnResponse    WSMessageType = "spawn_response"
	WSMsgVehicleRemove    WSMessageType = "vehicle_remove"
	WSMsgUserVehiclesList WSMessageType = "user_vehicles_list"
	WSMsgConditionUpdate  WSMessageType = "condition_update"
)

type WebSocketMessage struct {
	Type      WSMessageType `json:"type"`
	Timestamp time.Time     `json:"timestamp"`
	Data      interface{}   `json:"data"`

	UserSessionID *string `json:"user_session_id,omitempty"`
	RequestID     *string `json:"request_id,omitempty"`
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

	UserSessionID  *string `json:"user_session_id,omitempty"`
	CustomName     *string `json:"custom_name,omitempty"`
	SpawnRequestID *string `json:"spawn_request_id,omitempty"`
}

type SpawnRequestMessage struct {
	RequestID     string              `json:"request_id"`
	UserSessionID string              `json:"user_session_id"`
	SpawnRequest  VehicleSpawnRequest `json:"spawn_request"`
}

type SpawnResponseMessage struct {
	RequestID        string   `json:"request_id"`
	Success          bool     `json:"success"`
	SpawnedVehicleID *string  `json:"spawned_vehicle_id,omitempty"`
	ErrorMessage     *string  `json:"error_message,omitempty"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
}

type UserVehiclesListMessage struct {
	UserSessionID string                  `json:"user_session_id"`
	Vehicles      []VehiclePositionUpdate `json:"vehicles"`
	TotalCount    int                     `json:"total_count"`
}

type ConditionUpdateMessage struct {
	SegmentID         int64              `json:"segment_id"`
	AddedConditions   []RoadCondition    `json:"added_conditions,omitempty"`
	RemovedConditions []string           `json:"removed_conditions,omitempty"`
	NewVisualState    SegmentVisualState `json:"new_visual_state"`
}

type TrafficUpdate struct {
	RoadSegmentID   int64   `json:"road_segment_id"`
	FleetCount      int     `json:"fleet_count"`
	BackgroundCount int     `json:"background_count"`
	Capacity        int     `json:"capacity"`
	CongestionRatio float64 `json:"congestion_ratio"`

	ActiveConditions    []string `json:"active_conditions"`
	EffectiveSpeedLimit float64  `json:"effective_speed_limit"`
	AverageSpeed        float64  `json:"average_speed"`
	CongestionLevel     string   `json:"congestion_level"`
	VisualColor         string   `json:"visual_color"`
}
