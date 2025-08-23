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
}

type RedisRoadLoad struct {
	SegmentID       int64
	FleetCount      int
	BackgroundCount int
	Capacity        int
	CongestionRatio float64
	UpdatedAt       time.Time
}
