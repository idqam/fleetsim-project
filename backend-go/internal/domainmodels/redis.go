package domainmodels

import "time"

type RedisVehicleState struct {
	VehicleID      string
	Lat            float64
	Lng            float64
	SpeedKPH       float64
	FuelLevel      float64
	CurrentSegment *int64
	EdgeProgress   float64
	Status         string
	UpdatedAt      time.Time

	UserSessionID  *string    `json:"user_session_id,omitempty"`
	SpawnRequestID *string    `json:"spawn_request_id,omitempty"`
	CustomName     *string    `json:"custom_name,omitempty"`
	SpawnedAt      *time.Time `json:"spawned_at,omitempty"`

	SpeedMultiplier float64 `json:"speed_multiplier"`
	FuelMultiplier  float64 `json:"fuel_multiplier"`
}

type RedisRoadLoad struct {
	SegmentID       int64
	FleetCount      int
	BackgroundCount int
	Capacity        int
	CongestionRatio float64
	UpdatedAt       time.Time

	ActiveConditions    []string `json:"active_conditions"`
	EffectiveSpeedLimit float64  `json:"effective_speed_limit"`
	AverageSpeed        float64  `json:"average_speed"`

	VisualColor     string `json:"visual_color"`
	CongestionLevel string `json:"congestion_level"`
	ShowWarning     bool   `json:"show_warning"`
}

type RedisSpawnQueue struct {
	QueuedRequests  []string  `json:"queued_requests"`
	ProcessingCount int       `json:"processing_count"`
	LastProcessed   time.Time `json:"last_processed"`
	TotalProcessed  int64     `json:"total_processed"`
}
