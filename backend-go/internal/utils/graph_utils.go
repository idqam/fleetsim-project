package utils

import "owenvi.com/fleetsim/internal/domainmodels"

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
