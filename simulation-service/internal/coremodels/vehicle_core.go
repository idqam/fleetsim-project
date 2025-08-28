package coremodels

import (
	"container/heap"
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

type Position struct {
	X float64
	Y float64
	T time.Time
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
	IntersectionsCrossed int    `json:"intersections_crossed"`

	TravelDirection int64 `json:"travel_direction"`
	PreviousNodeID  int64 `json:"previous_node_id"`

	RecentPositions []Position `json:"recent_positions"`
	MaxTrailLength  int        `json:"max_trail_length"`
}

func (v *Vehicle) recordPosition(x, y float64) {
	v.RecentPositions = append(v.RecentPositions, Position{X: x, Y: y, T: time.Now()})
	if len(v.RecentPositions) > v.MaxTrailLength {
		v.RecentPositions = v.RecentPositions[len(v.RecentPositions)-v.MaxTrailLength:]
	}
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

	var x, y float64
	if v.TravelDirection >= 0 {
		x = startNode.Pos_X + (endNode.Pos_X-startNode.Pos_X)*v.SegmentProgress
		y = startNode.Pos_Y + (endNode.Pos_Y-startNode.Pos_Y)*v.SegmentProgress
	} else {
		x = endNode.Pos_X + (startNode.Pos_X-endNode.Pos_X)*(1.0-v.SegmentProgress)
		y = endNode.Pos_Y + (startNode.Pos_Y-endNode.Pos_Y)*(1.0-v.SegmentProgress)
	}

	v.recordPosition(x, y)
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

	if v.TravelDirection >= 0 {
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

func (v *Vehicle) UpdateProgress(deltaTimeSeconds float64, grid *Grid, router *VehicleRouter) error {
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
		if v.TravelDirection >= 0 {
			v.SegmentProgress += progressDelta
		} else {
			v.SegmentProgress -= progressDelta
		}
		v.TotalDistanceKM += distanceMovedKM
	}

	if v.SegmentProgress >= 1.0 || v.SegmentProgress <= 0.0 {
		nextNode, _ := v.GetNextNodeID(grid)
		if nextNode == v.TargetNodeID {
			v.Status = StatusReachedDestination
			return nil
		}
		decision, err := router.GetNextSegment(v, grid)
		if err != nil || decision.Reason == "dead_end" {
			v.Status = StatusDeadEnd
			return nil
		}
		v.PrepareMovementRequest(decision.ToSegmentID, nextNode)
		v.HandleMovementResponse(true, "", 0, grid)
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

func (v *Vehicle) PrepareMovementRequest(targetSegmentID int64, fromNodeID int64) {
	v.NextSegmentID = targetSegmentID
	v.Status = StatusWaitingForPermission
	v.PendingMovementRequestID = ksuid.New()
	v.LastMovementRequest = time.Now()
	v.PreviousNodeID = fromNodeID
}

func (v *Vehicle) HandleMovementResponse(accepted bool, reason string, alternativeSegmentID int64, grid *Grid) {
	if accepted {
		v.CurrentSegmentID = v.NextSegmentID
		if segment, exists := grid.Segments[v.CurrentSegmentID]; exists {
			if segment.StartNode == v.PreviousNodeID {
				v.TravelDirection = 1
				v.SegmentProgress = 0.0
			} else if segment.EndNode == v.PreviousNodeID {
				v.TravelDirection = -1
				v.SegmentProgress = 1.0
			} else {
				v.TravelDirection = 1
				v.SegmentProgress = 0.0
			}
		} else {
			v.SegmentProgress = 0.0
			v.TravelDirection = 1
		}
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
		TravelDirection:          v.TravelDirection,
		PreviousNodeID:           v.PreviousNodeID,
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

	candidateSegments := r.getCandidateSegments(currentNode, grid, vehicle)
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

func (r *VehicleRouter) getCandidateSegments(fromNodeID int64, grid *Grid, vehicle *Vehicle) []*RoadSegment {
	segmentIDs, exists := grid.Adjacency[fromNodeID]
	if !exists {
		return nil
	}

	var candidates []*RoadSegment
	for _, segID := range segmentIDs {
		if segID == vehicle.CurrentSegmentID {
			continue
		}
		if segment, exists := grid.Segments[segID]; exists {
			candidates = append(candidates, segment)
		}
	}

	if len(candidates) == 0 {
		for _, segID := range segmentIDs {
			if segment, exists := grid.Segments[segID]; exists {
				candidates = append(candidates, segment)
			}
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


type astarItem struct {
	node    int64
	priority float64
	index   int
}

type priorityQueue []*astarItem

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].priority < pq[j].priority }
func (pq priorityQueue) Swap(i, j int) { pq[i], pq[j] = pq[j], pq[i]; pq[i].index = i; pq[j].index = j }
func (pq *priorityQueue) Push(x interface{}) { item := x.(*astarItem); item.index = len(*pq); *pq = append(*pq, item) }
func (pq *priorityQueue) Pop() interface{} { old := *pq; n := len(old); item := old[n-1]; old[n-1] = nil; *pq = old[0 : n-1]; return item }

func (r *VehicleRouter) astarNextSegment(vehicle *Vehicle, startNodeID, goalNodeID int64, grid *Grid) *RoutingDecision {
	_, ok := grid.Nodes[startNodeID]
	if !ok {
		return nil
	}
	_, ok = grid.Nodes[goalNodeID]
	if !ok {
		return nil
	}

	openSet := &priorityQueue{}
	heap.Init(openSet)
	heap.Push(openSet, &astarItem{node: startNodeID, priority: 0})

	cameFromNode := make(map[int64]int64)
	cameFromSeg := make(map[int64]int64)
	gScore := make(map[int64]float64)
	fScore := make(map[int64]float64)

	for nid := range grid.Nodes {
		gScore[nid] = math.Inf(1)
		fScore[nid] = math.Inf(1)
	}
	gScore[startNodeID] = 0
	fScore[startNodeID] = r.heuristic(startNodeID, goalNodeID, grid)

	for openSet.Len() > 0 {
		item := heap.Pop(openSet).(*astarItem)
		current := item.node
		if current == goalNodeID {
			
			path := []int64{current}
			for path[len(path)-1] != startNodeID {
				prev := cameFromNode[path[len(path)-1]]
				path = append(path, prev)
			}
			
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			if len(path) < 2 {
				return nil
			}
			nextNode := path[1]
			segID := cameFromSeg[nextNode]
			segment, exists := grid.Segments[segID]
			if !exists {
				
				for _, sid := range grid.Adjacency[startNodeID] {
					s := grid.Segments[sid]
					if (s.StartNode == startNodeID && s.EndNode == nextNode) || (s.EndNode == startNodeID && s.StartNode == nextNode) {
						segID = s.ID
						segment = s
						break
					}
				}
			}
			decision := r.createDecision(segment, startNodeID, goalNodeID, grid)
			decision.ToSegmentID = segID
			decision.ToNodeID = nextNode
			decision.Reason = "optimal"
			decision.TotalCost = gScore[goalNodeID]
			return decision
		}

		adjSegIDs := grid.Adjacency[current]
		for _, segID := range adjSegIDs {
			seg, exists := grid.Segments[segID]
			if !exists {
				continue
			}
			var neighbor int64
			if seg.StartNode == current {
				neighbor = seg.EndNode
			} else {
				neighbor = seg.StartNode
			}
			if segID == vehicle.CurrentSegmentID {
				
				otherCount := 0
				for _, sID := range grid.Adjacency[current] {
					if sID != vehicle.CurrentSegmentID {
						otherCount++
					}
				}
				if otherCount > 0 {
					continue
				}
			}

			edgeCost := (r.DistanceWeight * seg.LengthKM) + (r.CongestionWeight * seg.CongestionFactor * seg.LengthKM)
			tentativeG := gScore[current] + edgeCost
			if tentativeG < gScore[neighbor] {
				cameFromNode[neighbor] = current
				cameFromSeg[neighbor] = segID
				gScore[neighbor] = tentativeG
				fScore[neighbor] = tentativeG + r.heuristic(neighbor, goalNodeID, grid)
				heap.Push(openSet, &astarItem{node: neighbor, priority: fScore[neighbor]})
			}
		}
	}

	return nil
}

func (r *VehicleRouter) heuristic(fromNodeID, targetNodeID int64, grid *Grid) float64 {
	fromNode, fromExists := grid.Nodes[fromNodeID]
	targetNode, targetExists := grid.Nodes[targetNodeID]
	if !fromExists || !targetExists {
		return 0
	}
	dx := targetNode.Pos_X - fromNode.Pos_X
	dy := targetNode.Pos_Y - fromNode.Pos_Y
	return math.Sqrt(dx*dx+dy*dy) / 1000.0
}