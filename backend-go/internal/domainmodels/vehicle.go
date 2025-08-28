package domainmodels

import (
	"math"
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

type MovementResult struct {
	NewProgress        float64 `json:"new_progress"`
	DistanceTraveled   float64 `json:"distance_traveled"`
	FuelConsumed       float64 `json:"fuel_consumed"`
	EffectiveSpeed     float64 `json:"effective_speed"`
	ReachedSegmentEnd  bool    `json:"reached_segment_end"`
	RemainingFuel      float64 `json:"remaining_fuel"`
	ReachedDestination bool    `json:"reached_destination"`
	Error              string  `json:"error,omitempty"`
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

	PlannedPath     []int64 `json:"planned_path,omitempty"`
	NextDecisionAt  int64   `json:"next_decision_at,omitempty"`
	OriginCell      *Cell   `json:"origin_cell,omitempty"`
	DestinationCell *Cell   `json:"destination_cell,omitempty"`
	ProximityLOD    bool
	Progress        float64

	FuelLevel          float64  `json:"fuel_level"`
	InitialFuelPercent *float64 `json:"initial_fuel_percent,omitempty"`
	SpeedMultiplier    float64  `json:"speed_multiplier"`

	SpawnedAt      *time.Time `json:"spawned_at,omitempty"`
	LastDecisionAt int64      `json:"last_decision_at,omitempty"`
	FailureReason  *string    `json:"failure_reason,omitempty"`

	TotalDistanceTraveled float64 `json:"total_distance_traveled"`
	TotalFuelConsumed     float64 `json:"total_fuel_consumed"`
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

func (v *Vehicle) CalculateMovementForTimeStep(timeStepSeconds float64, currentSegment *RoadSegment) MovementResult {
	if currentSegment == nil {
		return MovementResult{Error: "vehicle not on a segment"}
	}
	effectiveSpeed, distanceTraveledKM, segmentProgressIncrement := v.calculateMovementStep(timeStepSeconds, currentSegment)
	newProgress := v.SegmentProgress + segmentProgressIncrement
	fuelConsumed := v.calculateFuelConsumption(distanceTraveledKM, currentSegment)
	return MovementResult{
		NewProgress:       newProgress,
		DistanceTraveled:  distanceTraveledKM,
		FuelConsumed:      fuelConsumed,
		EffectiveSpeed:    effectiveSpeed,
		ReachedSegmentEnd: newProgress >= 1.0,
		RemainingFuel:     v.FuelLevel - fuelConsumed,
	}
}

func (v *Vehicle) calculateTrafficMultipliers(load TrafficLoadState) (speedMultiplier, fuelMultiplier float64) {
	if load.CapacityUtilization <= 0 {
		return 1.0, 1.0
	}

	if load.CapacityUtilization >= 1.0 {
		speedMultiplier = 0.3
	} else {
		avgSpeedFactor := math.Min(1.0, load.AverageSpeed/float64(v.Profile.MaxSpeedKPH))
		congestionEffect := 1.0 - 0.6*math.Pow(load.CapacityUtilization, 2)
		speedMultiplier = math.Min(avgSpeedFactor, congestionEffect)
	}

	trafficFactor := 1.0 + 0.5*math.Pow(load.CapacityUtilization, 1.5)
	speedFactor := 1.0
	if load.AverageSpeed < float64(v.Profile.MaxSpeedKPH)*0.5 {
		speedFactor += 0.2 * (1.0 - load.AverageSpeed/(float64(v.Profile.MaxSpeedKPH)*0.5))
	}
	fuelMultiplier = trafficFactor * speedFactor

	return speedMultiplier, fuelMultiplier
}
func (v *Vehicle) calculateEffectiveSpeed(segment *RoadSegment) float64 {
	maxSpeed := float64(v.Profile.MaxSpeedKPH) * v.SpeedMultiplier
	segmentLimit := segment.BaseSpeedKPH
	if segment.SpeedLimit != nil {
		segmentLimit = math.Min(segmentLimit, float64(*segment.SpeedLimit))
	}
	baseSpeed := math.Min(maxSpeed, segmentLimit)
	speedMultiplier := 1.0
	for _, condition := range segment.BaseConditions {
		speedMultiplier *= condition.SpeedMultiplier
	}
	for _, condition := range segment.TemporaryConditions {
		speedMultiplier *= condition.SpeedMultiplier
	}
	trafficMultiplier, _ := v.calculateTrafficMultipliers(segment.CurrentTrafficLoad)
	return baseSpeed * speedMultiplier * trafficMultiplier
}

func (v *Vehicle) calculateFuelConsumption(distanceKM float64, segment *RoadSegment) float64 {
	baseFuelPer100KM := v.Profile.ConsumptionL100KM
	fuelMultiplier := 1.0
	for _, condition := range segment.BaseConditions {
		fuelMultiplier *= condition.FuelMultiplier
	}
	for _, condition := range segment.TemporaryConditions {
		fuelMultiplier *= condition.FuelMultiplier
	}
	_, trafficFuelMultiplier := v.calculateTrafficMultipliers(segment.CurrentTrafficLoad)
	effectiveConsumption := baseFuelPer100KM * fuelMultiplier * trafficFuelMultiplier
	return (effectiveConsumption * distanceKM) / 100.0
}

func (v *Vehicle) CanEnterSegment(segment *RoadSegment) bool {

	if !segment.IsOpen {
		return false
	}

	if segment.Capacity != nil {
		currentCount := segment.CurrentTrafficLoad.VehicleCount
		if currentCount >= int(*segment.Capacity) {
			return false
		}
	}

	return true
}

func (v *Vehicle) HasReachedDestination() bool {
	if v.DestinationCell == nil || v.CurrentCell == nil {
		return false
	}

	return v.CurrentCell.Xpos == v.DestinationCell.Xpos &&
		v.CurrentCell.Ypos == v.DestinationCell.Ypos
}

func (v *Vehicle) UpdatePosition(timeStepSeconds float64, grid *Grid) MovementResult {
	if v.CurrentSegment == nil {
		return MovementResult{Error: "vehicle not on any segment"}
	}

	currentSpeed, distanceTraveled, progressIncrement := v.calculateMovementStep(timeStepSeconds, v.CurrentSegment)

	v.CurrentSpeedKPH = currentSpeed

	v.Progress += progressIncrement

	v.updateCurrentCellFromProgress(grid)

	if v.HasReachedDestination() {
		v.Status = constants.VehicleStatusCompleted
		return MovementResult{
			NewProgress:        v.Progress,
			DistanceTraveled:   distanceTraveled,
			ReachedDestination: true,
		}
	}

	return MovementResult{
		NewProgress:       v.Progress,
		DistanceTraveled:  distanceTraveled,
		ReachedSegmentEnd: v.Progress >= 1.0,
	}
}

func (v *Vehicle) updateCurrentCellFromProgress(grid *Grid) {
	if v.CurrentSegment == nil {
		return
	}

	startX := float64(v.CurrentSegment.StartX)
	startY := float64(v.CurrentSegment.StartY)
	endX := float64(v.CurrentSegment.EndX)
	endY := float64(v.CurrentSegment.EndY)

	currentX := startX + v.Progress*(endX-startX)
	currentY := startY + v.Progress*(endY-startY)

	cellX := int64(math.Round(currentX))
	cellY := int64(math.Round(currentY))

	coords := [2]int64{cellX, cellY}
	v.CurrentCell = grid.CoordIndex[coords]
}

func (v *Vehicle) calculateMovementStep(timeStepSeconds float64, segment *RoadSegment) (effectiveSpeed, distanceTraveled, progressIncrement float64) {
	effectiveSpeed = v.calculateEffectiveSpeed(segment)
	timeStepHours := timeStepSeconds / 3600.0
	distanceTraveled = effectiveSpeed * timeStepHours
	progressIncrement = distanceTraveled / segment.LengthKM
	return
}
