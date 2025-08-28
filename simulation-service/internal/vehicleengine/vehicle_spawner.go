package vehicleengine

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/segmentio/ksuid"
	"owenvi.com/simsim/internal/coremodels"
)

type VehicleSpawner struct {
	grid   *coremodels.Grid
	router *coremodels.VehicleRouter
	rng    *rand.Rand
}

func NewVehicleSpawner(grid *coremodels.Grid) *VehicleSpawner {
	return &VehicleSpawner{
		grid:   grid,
		router: coremodels.NewVehicleRouter(),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (vs *VehicleSpawner) SpawnVehicle() (*coremodels.Vehicle, error) {
	spawnSegment := vs.selectRandomSegment()
	if spawnSegment == nil {
		return nil, fmt.Errorf("no valid segments for spawning")
	}

	targetNode := vs.selectRandomTargetNode(spawnSegment)
	if targetNode == -1 {
		return nil, fmt.Errorf("no valid target nodes")
	}

	baseSpeed := 30 + vs.rng.Float64()*50

	direction := int64(1)
	if vs.rng.Intn(2) == 0 {
		direction = -1
	}

	vehicle := &coremodels.Vehicle{
		ID:               ksuid.New(),
		CurrentSegmentID: spawnSegment.ID,
		SegmentProgress:  vs.rng.Float64() * 0.2,
		TargetNodeID:     targetNode,
		BaseSpeedKPH:     baseSpeed,
		CurrentSpeedKPH:  baseSpeed,
		Status:           coremodels.StatusMoving,
		SpawnTime:        time.Now(),
		LastUpdate:       time.Now(),
		TravelDirection:  direction,
		MaxTrailLength:   15,
	}

	return vehicle, nil
}

func (vs *VehicleSpawner) SpawnMultipleVehicles(count int) ([]*coremodels.Vehicle, error) {
	vehicles := make([]*coremodels.Vehicle, 0, count)

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

func (vs *VehicleSpawner) selectRandomSegment() *coremodels.RoadSegment {
	if len(vs.grid.Segments) == 0 {
		return nil
	}

	segmentList := make([]*coremodels.RoadSegment, 0, len(vs.grid.Segments))
	for _, segment := range vs.grid.Segments {
		segmentList = append(segmentList, segment)
	}

	return segmentList[vs.rng.Intn(len(segmentList))]
}


func (vs *VehicleSpawner) selectRandomTargetNode(excludeSegment *coremodels.RoadSegment) int64 {
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