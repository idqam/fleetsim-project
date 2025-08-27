package utils

import (
	"fmt"

	"owenvi.com/fleetsim/internal/domainmodels"
)

func BFSMarkConnected(segmentID int64, adjacency map[int64][]int64, explored map[int64]bool) {
	size := len(adjacency)
	queue := NewIntQueue(size)

	explored[segmentID] = true
	queue.Enqueue(segmentID)
	for !queue.IsEmpty() {
		currSeg := queue.Dequeue()
		for _, neighbor := range adjacency[currSeg] {
			if !explored[neighbor] {
				explored[neighbor] = true
				queue.Enqueue(neighbor)
			}
		}
	}
}

func FindConnectedComponents(grid *domainmodels.Grid) [][]int64 {

	adjacency := make(map[int64][]int64)
	allSegments := make(map[int64]bool)

	for _, cell := range grid.Cells {
		for _, cellRoad := range cell.RoadSegments {
			segment := cellRoad.RoadSegment
			allSegments[segment.ID] = true

			connections := FindConnectedSegments(segment, grid)
			adjacency[segment.ID] = connections
		}
	}

	if len(allSegments) == 0 {
		return [][]int64{}
	}

	visited := make(map[int64]bool)
	var components [][]int64

	for segmentID := range allSegments {
		if !visited[segmentID] {

			component := []int64{}
			DfsCollectComponent(segmentID, adjacency, visited, &component)
			components = append(components, component)
		}
	}

	return components
}

func FindConnectedSegments(target domainmodels.RoadSegment, grid *domainmodels.Grid) []int64 {
	var connections []int64

	targetStartX, targetStartY := target.StartX, target.StartY
	targetEndX, targetEndY := target.EndX, target.EndY

	for _, cell := range grid.Cells {
		for _, cellRoad := range cell.RoadSegments {
			other := cellRoad.RoadSegment
			if other.ID == target.ID {
				continue
			}

			if (other.StartX == targetStartX && other.StartY == targetStartY) ||
				(other.StartX == targetEndX && other.StartY == targetEndY) ||
				(other.EndX == targetStartX && other.EndY == targetStartY) ||
				(other.EndX == targetEndX && other.EndY == targetEndY) {
				connections = append(connections, other.ID)
			}
		}
	}

	return connections
}

func DfsCollectComponent(segmentID int64, adjacency map[int64][]int64, visited map[int64]bool, component *[]int64) {
	visited[segmentID] = true
	*component = append(*component, segmentID)

	for _, connectedID := range adjacency[segmentID] {
		if !visited[connectedID] {
			DfsCollectComponent(connectedID, adjacency, visited, component)
		}
	}
}
func (ei EndpointIndex) PrintIndexStats() {
	totalPoints := len(ei)
	totalSegmentEndpoints := 0
	maxSegmentsAtPoint := 0
	intersectionCount := 0
	deadEndCount := 0

	for _, segments := range ei {
		totalSegmentEndpoints += len(segments)
		if len(segments) > maxSegmentsAtPoint {
			maxSegmentsAtPoint = len(segments)
		}
		if len(segments) >= 3 {
			intersectionCount++
		}
		if len(segments) == 1 {
			deadEndCount++
		}
	}

	fmt.Printf("Endpoint Index Statistics:\n")
	fmt.Printf("  • Total indexed points: %d\n", totalPoints)
	fmt.Printf("  • Total segment endpoints: %d\n", totalSegmentEndpoints)
	fmt.Printf("  • Max segments at one point: %d\n", maxSegmentsAtPoint)
	fmt.Printf("  • Major intersections (3+ segments): %d\n", intersectionCount)
	fmt.Printf("  • Dead ends (1 segment): %d\n", deadEndCount)

	if totalPoints > 0 {
		avgSegmentsPerPoint := float64(totalSegmentEndpoints) / float64(totalPoints)
		fmt.Printf("  • Average segments per point: %.1f\n", avgSegmentsPerPoint)
	}
}

func (ei EndpointIndex) ValidateEndpointIndex(grid *domainmodels.Grid) []string {
	var issues []string

	segmentsSeen := make(map[int64]bool)
	for _, segments := range ei {
		for _, segmentID := range segments {
			segmentsSeen[segmentID] = true
		}
	}

	actualSegments := make(map[int64]bool)
	for _, cell := range grid.Cells {
		for _, cellRoad := range cell.RoadSegments {
			actualSegments[cellRoad.RoadSegment.ID] = true
		}
	}

	for segmentID := range actualSegments {
		if !segmentsSeen[segmentID] {
			issues = append(issues, fmt.Sprintf("Segment %d not found in endpoint index", segmentID))
		}
	}

	for segmentID := range segmentsSeen {
		if !actualSegments[segmentID] {
			issues = append(issues, fmt.Sprintf("Segment %d in endpoint index but not in grid", segmentID))
		}
	}

	return issues
}
func FindConnectedSegmentsFast(target domainmodels.RoadSegment, endpointIndex EndpointIndex) []int64 {
	connectionSet := make(map[int64]bool)

	startPoint := [2]int64{target.StartX, target.StartY}
	if segmentsAtStart, exists := endpointIndex[startPoint]; exists {
		for _, segmentID := range segmentsAtStart {
			if segmentID != target.ID {
				connectionSet[segmentID] = true
			}
		}
	}

	endPoint := [2]int64{target.EndX, target.EndY}
	if segmentsAtEnd, exists := endpointIndex[endPoint]; exists {
		for _, segmentID := range segmentsAtEnd {
			if segmentID != target.ID {
				connectionSet[segmentID] = true
			}
		}
	}

	connections := make([]int64, 0, len(connectionSet))
	for segmentID := range connectionSet {
		connections = append(connections, segmentID)
	}

	return connections
}

func (ei EndpointIndex) GetEndpointCount(x, y int64) int {
	point := [2]int64{x, y}
	return len(ei[point])
}

func (ei EndpointIndex) GetIntersectionPoints() [][2]int64 {
	var intersections [][2]int64

	for point, segments := range ei {
		if len(segments) >= 3 {
			intersections = append(intersections, point)
		}
	}

	return intersections
}

func (ei EndpointIndex) GetDeadEnds() [][2]int64 {
	var deadEnds [][2]int64

	for point, segments := range ei {
		if len(segments) == 1 {
			deadEnds = append(deadEnds, point)
		}
	}

	return deadEnds
}

type EndpointIndex map[[2]int64][]int64

func BuildEndpointIndex(grid *domainmodels.Grid) EndpointIndex {
	index := make(EndpointIndex)

	processedSegments := make(map[int64]bool)

	for _, cell := range grid.Cells {
		for _, cellRoad := range cell.RoadSegments {
			segment := cellRoad.RoadSegment

			if processedSegments[segment.ID] {
				continue
			}
			processedSegments[segment.ID] = true

			startPoint := [2]int64{segment.StartX, segment.StartY}
			index[startPoint] = append(index[startPoint], segment.ID)

			endPoint := [2]int64{segment.EndX, segment.EndY}
			index[endPoint] = append(index[endPoint], segment.ID)
		}
	}

	return index
}
