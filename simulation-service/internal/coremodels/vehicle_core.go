package coremodels

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/segmentio/ksuid"
)

type VehicleStatus int

const (
	StatusMoving VehicleStatus = iota
	StatusWaitingForPermission
	StatusReachedDestination
	StatusDeadEnd
	StatusError
)

func (vs VehicleStatus) String() string {
	switch vs {
	case StatusMoving:
		return "moving"
	case StatusWaitingForPermission:
		return "waiting_for_permission"
	case StatusReachedDestination:
		return "reached_destination"
	case StatusDeadEnd:
		return "dead_end"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

type Vehicle struct {
	ID ksuid.KSUID `json:"id"`
	
	CurrentSegmentID int64   `json:"current_segment_id"`
	SegmentProgress  float64 `json:"segment_progress"`
	
	TargetNodeID int64 `json:"target_node_id"`
	
	BaseSpeedKPH    float64       `json:"base_speed_kph"`
	CurrentSpeedKPH float64       `json:"current_speed_kph"`
	Status          VehicleStatus `json:"status"`
	
	NextSegmentID     int64 `json:"next_segment_id,omitempty"`
	RoutingDecisionID int64 `json:"routing_decision_id,omitempty"`
	
	PendingMovementRequestID ksuid.KSUID `json:"pending_movement_request_id,omitempty"`
	LastMovementRequest      time.Time   `json:"last_movement_request"`
	MovementDenialCount      int         `json:"movement_denial_count"`
	
	SpawnTime       time.Time `json:"spawn_time"`
	LastUpdate      time.Time `json:"last_update"`
	TotalDistanceKM float64   `json:"total_distance_km"`
	
	AverageSpeedKPH     float64 `json:"average_speed_kph"`
	TimeStuckSeconds    int64   `json:"time_stuck_seconds"`
	RouteChanges        int     `json:"route_changes"`
	IntersectionsCrossed int     `json:"intersections_crossed"`
}

func (v *Vehicle) GetCurrentPosition(grid *Grid) (float64, float64, error) {
	segment, exists := grid.Segments[v.CurrentSegmentID]
	if !exists {
		return 0, 0, fmt.Errorf("segment %d not found for vehicle %s", v.CurrentSegmentID, v.ID.String())
	}
	
	startNode, startExists := grid.Nodes[segment.StartNode]
	endNode, endExists := grid.Nodes[segment.EndNode]
	
	if !startExists || !endExists {
		return 0, 0, fmt.Errorf("nodes for segment %d not found", v.CurrentSegmentID)
	}
	
	x := startNode.Pos_X + (endNode.Pos_X-startNode.Pos_X)*v.SegmentProgress
	y := startNode.Pos_Y + (endNode.Pos_Y-startNode.Pos_Y)*v.SegmentProgress
	
	return x, y, nil
}

func (v *Vehicle) GetCurrentEffectiveSpeed(grid *Grid) float64 {
	segment, exists := grid.Segments[v.CurrentSegmentID]
	if !exists {
		return v.BaseSpeedKPH
	}
	
	effectiveSpeed := v.BaseSpeedKPH / segment.CongestionFactor
	v.CurrentSpeedKPH = effectiveSpeed
	return effectiveSpeed
}

func (v *Vehicle) GetNextNodeID(grid *Grid) (int64, error) {
	segment, exists := grid.Segments[v.CurrentSegmentID]
	if !exists {
		return 0, fmt.Errorf("segment %d not found", v.CurrentSegmentID)
	}
	
	if v.SegmentProgress > 0.5 {
		return segment.EndNode, nil
	}
	return segment.StartNode, nil
}

func (v *Vehicle) IsAtIntersection() bool {
	const intersectionThreshold = 0.05
	return v.SegmentProgress <= intersectionThreshold || v.SegmentProgress >= (1.0-intersectionThreshold)
}

func (v *Vehicle) HasReachedTarget(grid *Grid) bool {
	if v.Status == StatusReachedDestination {
		return true
	}
	
	nextNode, err := v.GetNextNodeID(grid)
	if err != nil {
		return false
	}
	
	return nextNode == v.TargetNodeID && v.IsAtIntersection()
}

func (v *Vehicle) CanMakeMovementRequest() bool {
	if v.Status == StatusWaitingForPermission {
		return false
	}
	
	if v.Status == StatusReachedDestination || v.Status == StatusDeadEnd {
		return false
	}
	
	timeSinceLastRequest := time.Since(v.LastMovementRequest)
	backoffDuration := time.Duration(v.MovementDenialCount*100) * time.Millisecond
	minWaitTime := 50 * time.Millisecond
	
	if backoffDuration > minWaitTime {
		return timeSinceLastRequest > backoffDuration
	}
	
	return timeSinceLastRequest > minWaitTime
}

func (v *Vehicle) UpdateProgress(deltaTimeSeconds float64, grid *Grid) error {
	if v.Status != StatusMoving {
		return nil
	}
	
	segment, exists := grid.Segments[v.CurrentSegmentID]
	if !exists {
		return fmt.Errorf("cannot update progress: segment %d not found", v.CurrentSegmentID)
	}
	
	effectiveSpeed := v.GetCurrentEffectiveSpeed(grid)
	distanceMovedKM := (effectiveSpeed / 3600.0) * deltaTimeSeconds
	
	if segment.LengthKM > 0 {
		progressDelta := distanceMovedKM / segment.LengthKM
		v.SegmentProgress += progressDelta
		v.TotalDistanceKM += distanceMovedKM
	}
	
	if v.SegmentProgress > 1.0 {
		v.SegmentProgress = 1.0
	}
	if v.SegmentProgress < 0.0 {
		v.SegmentProgress = 0.0
	}
	
	v.LastUpdate = time.Now()
	return nil
}

func (v *Vehicle) PrepareMovementRequest(targetSegmentID int64) {
	v.NextSegmentID = targetSegmentID
	v.Status = StatusWaitingForPermission
	v.PendingMovementRequestID = ksuid.New()
	v.LastMovementRequest = time.Now()
}

func (v *Vehicle) HandleMovementResponse(accepted bool, reason string, alternativeSegmentID int64) {
	if accepted {
		v.CurrentSegmentID = v.NextSegmentID
		v.SegmentProgress = 0.0
		v.Status = StatusMoving
		v.MovementDenialCount = 0
		v.IntersectionsCrossed++
		v.NextSegmentID = 0
		v.PendingMovementRequestID = ksuid.Nil
	} else {
		v.MovementDenialCount++
		v.Status = StatusMoving
		
		switch reason {
		case "dead_end":
			v.Status = StatusDeadEnd
		case "capacity_full", "heavy_congestion":
			if alternativeSegmentID != 0 {
				v.NextSegmentID = alternativeSegmentID
				v.RouteChanges++
			}
		case "segment_blocked":
			v.TimeStuckSeconds += int64(time.Since(v.LastMovementRequest).Seconds())
		}
		
		v.NextSegmentID = 0
		v.PendingMovementRequestID = ksuid.Nil
	}
	
	v.LastUpdate = time.Now()
}

func (v *Vehicle) UpdateAverageSpeed() {
	if v.AverageSpeedKPH == 0 {
		v.AverageSpeedKPH = v.CurrentSpeedKPH
	} else {
		alpha := 0.1
		v.AverageSpeedKPH = alpha*v.CurrentSpeedKPH + (1-alpha)*v.AverageSpeedKPH
	}
}

func (v *Vehicle) GetDistanceToTarget(grid *Grid) (float64, error) {
	currentX, currentY, err := v.GetCurrentPosition(grid)
	if err != nil {
		return math.Inf(1), err
	}
	
	targetNode, exists := grid.Nodes[v.TargetNodeID]
	if !exists {
		return math.Inf(1), fmt.Errorf("target node %d not found", v.TargetNodeID)
	}
	
	dx := targetNode.Pos_X - currentX
	dy := targetNode.Pos_Y - currentY
	return math.Sqrt(dx*dx + dy*dy) / 1000.0, nil
}

func (v *Vehicle) IsStuck() bool {
	return v.MovementDenialCount > 5 || v.TimeStuckSeconds > 30
}

func (v *Vehicle) Clone() *Vehicle {
	return &Vehicle{
		ID:                       v.ID,
		CurrentSegmentID:         v.CurrentSegmentID,
		SegmentProgress:          v.SegmentProgress,
		TargetNodeID:             v.TargetNodeID,
		BaseSpeedKPH:             v.BaseSpeedKPH,
		CurrentSpeedKPH:          v.CurrentSpeedKPH,
		Status:                   v.Status,
		NextSegmentID:            v.NextSegmentID,
		RoutingDecisionID:        v.RoutingDecisionID,
		PendingMovementRequestID: v.PendingMovementRequestID,
		LastMovementRequest:      v.LastMovementRequest,
		MovementDenialCount:      v.MovementDenialCount,
		SpawnTime:                v.SpawnTime,
		LastUpdate:               v.LastUpdate,
		TotalDistanceKM:          v.TotalDistanceKM,
		AverageSpeedKPH:          v.AverageSpeedKPH,
		TimeStuckSeconds:         v.TimeStuckSeconds,
		RouteChanges:             v.RouteChanges,
		IntersectionsCrossed:     v.IntersectionsCrossed,
	}
}

type RoutingDecision struct {
	FromNodeID     int64   `json:"from_node_id"`
	ToSegmentID    int64   `json:"to_segment_id"`
	ToNodeID       int64   `json:"to_node_id"`
	DistanceCost   float64 `json:"distance_cost"`
	CongestionCost float64 `json:"congestion_cost"`
	TotalCost      float64 `json:"total_cost"`
	Reason         string  `json:"reason"`
}

type VehicleRouter struct {
	DistanceWeight    float64
	CongestionWeight  float64
	ExplorationRate   float64
}

func NewVehicleRouter() *VehicleRouter {
	return &VehicleRouter{
		DistanceWeight:   0.6,
		CongestionWeight: 0.4,
		ExplorationRate:  0.15,
	}
}

func (r *VehicleRouter) GetNextSegment(vehicle *Vehicle, grid *Grid) (*RoutingDecision, error) {
	currentNode, err := vehicle.GetNextNodeID(grid)
	if err != nil {
		return nil, err
	}
	
	if currentNode == vehicle.TargetNodeID {
		return &RoutingDecision{
			FromNodeID: currentNode,
			Reason:     "reached_destination",
		}, nil
	}
	
	candidateSegments := r.getCandidateSegments(currentNode, grid)
	if len(candidateSegments) == 0 {
		return &RoutingDecision{
			FromNodeID: currentNode,
			Reason:     "dead_end",
		}, nil
	}
	
	bestDecision := r.evaluateSegments(candidateSegments, currentNode, vehicle.TargetNodeID, grid)
	
	if rand.Float64() < r.ExplorationRate && len(candidateSegments) > 1 {
		randomIdx := rand.Intn(len(candidateSegments))
		randomSeg := candidateSegments[randomIdx]
		bestDecision = r.createDecision(randomSeg, currentNode, vehicle.TargetNodeID, grid)
		bestDecision.Reason = "exploration"
	}
	
	return bestDecision, nil
}

func (r *VehicleRouter) getCandidateSegments(fromNodeID int64, grid *Grid) []*RoadSegment {
	segmentIDs, exists := grid.Adjacency[fromNodeID]
	if !exists {
		return nil
	}
	
	var candidates []*RoadSegment
	for _, segID := range segmentIDs {
		if segment, exists := grid.Segments[segID]; exists {
			candidates = append(candidates, segment)
		}
	}
	
	return candidates
}

func (r *VehicleRouter) evaluateSegments(segments []*RoadSegment, fromNodeID, targetNodeID int64, grid *Grid) *RoutingDecision {
	var bestDecision *RoutingDecision
	bestCost := math.Inf(1)
	
	for _, segment := range segments {
		decision := r.createDecision(segment, fromNodeID, targetNodeID, grid)
		if decision.TotalCost < bestCost {
			bestCost = decision.TotalCost
			bestDecision = decision
		}
	}
	
	if bestDecision != nil {
		bestDecision.Reason = "optimal"
	}
	
	return bestDecision
}

func (r *VehicleRouter) createDecision(segment *RoadSegment, fromNodeID, targetNodeID int64, grid *Grid) *RoutingDecision {
	toNodeID := segment.EndNode
	if segment.EndNode == fromNodeID {
		toNodeID = segment.StartNode
	}
	
	distanceCost := segment.LengthKM
	congestionCost := segment.CongestionFactor * segment.LengthKM
	
	distanceToTarget := r.calculateDistanceToTarget(toNodeID, targetNodeID, grid)
	heuristicCost := distanceToTarget * 0.1
	
	totalCost := (r.DistanceWeight * distanceCost) + 
				(r.CongestionWeight * congestionCost) + 
				heuristicCost
	
	return &RoutingDecision{
		FromNodeID:     fromNodeID,
		ToSegmentID:    segment.ID,
		ToNodeID:       toNodeID,
		DistanceCost:   distanceCost,
		CongestionCost: congestionCost,
		TotalCost:      totalCost,
	}
}

func (r *VehicleRouter) calculateDistanceToTarget(fromNodeID, targetNodeID int64, grid *Grid) float64 {
	fromNode, fromExists := grid.Nodes[fromNodeID]
	targetNode, targetExists := grid.Nodes[targetNodeID]
	
	if !fromExists || !targetExists {
		return math.Inf(1)
	}
	
	dx := targetNode.Pos_X - fromNode.Pos_X
	dy := targetNode.Pos_Y - fromNode.Pos_Y
	return math.Sqrt(dx*dx + dy*dy) / 1000.0
}

type VehicleSpawner struct {
	grid   *Grid
	router *VehicleRouter
	rng    *rand.Rand
}

func NewVehicleSpawner(grid *Grid) *VehicleSpawner {
	return &VehicleSpawner{
		grid:   grid,
		router: NewVehicleRouter(),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (vs *VehicleSpawner) SpawnVehicle() (*Vehicle, error) {
	spawnSegment := vs.selectRandomSegment()
	if spawnSegment == nil {
		return nil, fmt.Errorf("no valid segments for spawning")
	}
	
	targetNode := vs.selectRandomTargetNode(spawnSegment)
	if targetNode == -1 {
		return nil, fmt.Errorf("no valid target nodes")
	}
	
	baseSpeed := 30 + vs.rng.Float64()*50
	
	vehicle := &Vehicle{
		ID:               ksuid.New(),
		CurrentSegmentID: spawnSegment.ID,
		SegmentProgress:  vs.rng.Float64() * 0.2,
		TargetNodeID:     targetNode,
		BaseSpeedKPH:     baseSpeed,
		CurrentSpeedKPH:  baseSpeed,
		Status:           StatusMoving,
		SpawnTime:        time.Now(),
		LastUpdate:       time.Now(),
	}
	
	return vehicle, nil
}

func (vs *VehicleSpawner) SpawnMultipleVehicles(count int) ([]*Vehicle, error) {
	vehicles := make([]*Vehicle, 0, count)
	
	for i := 0; i < count; i++ {
		vehicle, err := vs.SpawnVehicle()
		if err != nil {
			continue
		}
		vehicles = append(vehicles, vehicle)
	}
	
	if len(vehicles) == 0 {
		return nil, fmt.Errorf("failed to spawn any vehicles")
	}
	
	return vehicles, nil
}

func (vs *VehicleSpawner) selectRandomSegment() *RoadSegment {
	if len(vs.grid.Segments) == 0 {
		return nil
	}
	
	segmentList := make([]*RoadSegment, 0, len(vs.grid.Segments))
	for _, segment := range vs.grid.Segments {
		segmentList = append(segmentList, segment)
	}
	
	return segmentList[vs.rng.Intn(len(segmentList))]
}

func (vs *VehicleSpawner) selectRandomTargetNode(excludeSegment *RoadSegment) int64 {
	if len(vs.grid.Nodes) < 3 {
		return -1
	}
	
	nodeList := make([]int64, 0, len(vs.grid.Nodes))
	for nodeID := range vs.grid.Nodes {
		if nodeID != excludeSegment.StartNode && nodeID != excludeSegment.EndNode {
			if len(vs.grid.Adjacency[nodeID]) > 0 {
				nodeList = append(nodeList, nodeID)
			}
		}
	}
	
	if len(nodeList) == 0 {
		return -1
	}
	
	return nodeList[vs.rng.Intn(len(nodeList))]
}