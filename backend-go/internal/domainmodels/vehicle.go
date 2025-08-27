package domainmodels

import (
	"time"

	"owenvi.com/fleetsim/internal/constants"
)

type VehicleProfile struct {
	ID                int64                 `json:"id"`
	Name              string                `json:"name"`
	VehicleType       constants.VehicleType `json:"vehicle_type"`
	TankLiters        float64               `json:"tank_liters"`
	ConsumptionL100KM float64               `json:"consumption_l_per_100km"`
	MaxSpeedKPH       int                   `json:"max_speed_kph"`
	CargoCapacityKG   float64               `json:"cargo_capacity_kg"`
}

type Vehicle struct {
    ID      string                  `json:"id"`
    Class   constants.VehicleClass  `json:"class"`
    Profile VehicleProfile          `json:"profile"`
    Status  constants.VehicleStatus `json:"status"`
    
    
    CurrentCell     *Cell        `json:"current_cell,omitempty"`
    CurrentSegment  *RoadSegment `json:"current_segment,omitempty"`  
    SegmentProgress float64      `json:"segment_progress"`           
    
    
    CurrentSpeedKPH    float64 `json:"current_speed_kph"`
    TargetSpeedKPH     float64 `json:"target_speed_kph"`    
    LastMovementUpdate int64   `json:"last_movement_update"`
    
    
    PlannedPath       []int64 `json:"planned_path,omitempty"`
    NextDecisionAt    int64   `json:"next_decision_at,omitempty"`
    OriginCell        *Cell   `json:"origin_cell,omitempty"`
    DestinationCell   *Cell   `json:"destination_cell,omitempty"`
	ProximityLOD bool 
    
    
    FuelLevel          float64 `json:"fuel_level"`
    InitialFuelPercent *float64 `json:"initial_fuel_percent,omitempty"`
    SpeedMultiplier    float64 `json:"speed_multiplier"`
    
    SpawnedAt       *time.Time `json:"spawned_at,omitempty"`
    LastDecisionAt  int64      `json:"last_decision_at,omitempty"`
    FailureReason   *string    `json:"failure_reason,omitempty"`
    
    
    TotalDistanceTraveled float64 `json:"total_distance_traveled"`
    TotalFuelConsumed    float64 `json:"total_fuel_consumed"`
}


func (v *Vehicle) TankLiters() float64 {
	return v.Profile.TankLiters
}

func (v *Vehicle) MaxSpeedKPH() int {
	return v.Profile.MaxSpeedKPH
}

func (v *Vehicle) VehicleType() constants.VehicleType {
	return v.Profile.VehicleType
}

func (v *Vehicle) ConsumptionL100KM() float64 {
	return v.Profile.ConsumptionL100KM
}

func (v *Vehicle) GetFuelRange() float64 {
	if v.Profile.TankLiters == 0 {
		return 0
	}
	return (v.Profile.TankLiters / v.Profile.ConsumptionL100KM) * 100
}

func (v *Vehicle) GetFuelPercentage() float64 {
	if v.Profile.TankLiters == 0 {
		return 0
	}
	return v.FuelLevel / v.Profile.TankLiters
}

func (v *Vehicle) IsLowFuel() bool {
	return v.GetFuelPercentage() < 0.25
}

func (v *Vehicle) CanUseSegment(segment *RoadSegment) bool {
	if segment.SpeedLimit == nil {
		return true
	}
	return v.Profile.MaxSpeedKPH >= int(*segment.SpeedLimit)
}

func (v *Vehicle) EstimatedTravelTime(distanceKM float64) time.Duration {
	speedKPH := float64(v.Profile.MaxSpeedKPH) * v.SpeedMultiplier
	if speedKPH <= 0 {
		return 0
	}
	hours := distanceKM / speedKPH
	return time.Duration(hours * float64(time.Hour))
}
