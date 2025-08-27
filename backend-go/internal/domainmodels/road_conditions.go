package domainmodels

import "time"

type RoadCondition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	SpeedMultiplier float64 `json:"speed_multiplier"`
	FuelMultiplier  float64 `json:"fuel_multiplier"`

	IsTemporary bool       `json:"is_temporary"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Severity    string     `json:"severity"`

	VisualColor   string `json:"visual_color"`
	VisualPattern string `json:"visual_pattern"`
}

type TrafficLoadState struct {
	VehicleCount        int       `json:"vehicle_count"`
	CapacityUtilization float64   `json:"capacity_utilization"`
	AverageSpeed        float64   `json:"average_speed"`
	LastUpdated         time.Time `json:"last_updated"`
}

type SegmentVisualState struct {
	PrimaryColor   string  `json:"primary_color"`
	SecondaryColor string  `json:"secondary_color"`
	Opacity        float64 `json:"opacity"`
	AnimationSpeed float64 `json:"animation_speed"`
	ShowWarning    bool    `json:"show_warning"`
	WarningMessage string  `json:"warning_message"`
}

type RoadSegment struct {
	ID     int64 `json:"id"`
	StartX int64 `json:"start_x"`
	StartY int64 `json:"start_y"`
	EndX   int64 `json:"end_x"`
	EndY   int64 `json:"end_y"`

	LengthKM float64 `json:"length_km"` 
	BaseSpeedKPH float64 `json:"base_speed_kph"` 

	SpeedLimit *int64 `json:"speed_limit,omitempty"`
	Capacity   *int64 `json:"capacity,omitempty"`
	IsOpen     bool   `json:"is_open"`

	BaseConditions      []RoadCondition `json:"base_conditions"`
	TemporaryConditions []RoadCondition `json:"temporary_conditions"`

	CurrentTrafficLoad  TrafficLoadState `json:"current_traffic_load"`
	EffectiveSpeedLimit float64          `json:"effective_speed_limit"`
	CongestionLevel     string           `json:"congestion_level"`

	VisualState SegmentVisualState `json:"visual_state"`
}
