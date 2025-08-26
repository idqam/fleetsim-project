package gridloader

import (
	"fmt"

	"owenvi.com/fleetsim/internal/domainmodels"
)

func (gl *GridLoader) validateAndRepairConnectivity(grid *domainmodels.Grid) error {
	fmt.Printf("Validating and repairing network connectivity...\n")
	
	
	components := gl.findConnectedComponents(grid)
	
	if len(components) <= 1 {
		fmt.Printf("Network is fully connected (%d components)\n", len(components))
		return nil 
	}
	
	fmt.Printf("Found %d disconnected components, attempting repair...\n", len(components))
	
	
	connectionsAdded := gl.connectDisconnectedComponents(grid, components)
	
	
	componentsAfterRepair := gl.findConnectedComponents(grid)
	
	if len(componentsAfterRepair) > 1 {
		return fmt.Errorf("connectivity repair failed: %d components remain after adding %d connections",
			len(componentsAfterRepair), connectionsAdded)
	}
	
	fmt.Printf("Connectivity repair successful: added %d bridge connections\n", connectionsAdded)
	return nil
}


func (gl *GridLoader) findConnectedComponents(grid *domainmodels.Grid) [][]int64 {
	
	adjacency := make(map[int64][]int64)
	allSegments := make(map[int64]bool)
	
	
	for _, cell := range grid.Cells {
		for _, cellRoad := range cell.RoadSegments {
			segment := cellRoad.RoadSegment
			allSegments[segment.ID] = true
			
			
			connections := gl.findConnectedSegments(segment, grid)
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
			gl.dfsCollectComponent(segmentID, adjacency, visited, &component)
			components = append(components, component)
		}
	}
	
	return components
}

func (gl *GridLoader) dfsCollectComponent(segmentID int64, adjacency map[int64][]int64, visited map[int64]bool, component *[]int64) {
	visited[segmentID] = true
	*component = append(*component, segmentID)
	
	
	for _, connectedID := range adjacency[segmentID] {
		if !visited[connectedID] {
			gl.dfsCollectComponent(connectedID, adjacency, visited, component)
		}
	}
}


func (gl *GridLoader) connectDisconnectedComponents(grid *domainmodels.Grid, components [][]int64) int {
	if len(components) <= 1 {
		return 0 
	}
	
	connectionsAdded := 0

	largestComponent := gl.findLargestComponent(components)
	
	for i, component := range components {
		if i == largestComponent {
			continue 
		}
		
		bridgeConnection := gl.findBestBridgeConnection(grid, components[largestComponent], component)
		
		if bridgeConnection != nil {
			
			if gl.createBridgeSegment(grid, bridgeConnection) {
				connectionsAdded++
				fmt.Printf("Added bridge connection: (%d,%d) -> (%d,%d)\n",
					bridgeConnection.FromX, bridgeConnection.FromY,
					bridgeConnection.ToX, bridgeConnection.ToY)
			}
		}
	}
	
	return connectionsAdded
}


type BridgeConnection struct {
	FromX, FromY int64  
	ToX, ToY     int64  
	Distance     int64  
	FromSegment  int64  
	ToSegment    int64  
}

func (gl *GridLoader) findLargestComponent(components [][]int64) int {
	largestIndex := 0
	largestSize := len(components[0])
	
	for i, component := range components {
		if len(component) > largestSize {
			largestSize = len(component)
			largestIndex = i
		}
	}
	
	return largestIndex
}

func (gl *GridLoader) findBestBridgeConnection(grid *domainmodels.Grid, componentA, componentB []int64) *BridgeConnection {
	var bestConnection *BridgeConnection
	shortestDistance := int64(500) 
	
	cellsA := gl.getSegmentCells(grid, componentA)
	cellsB := gl.getSegmentCells(grid, componentB)
	
	for _, cellA := range cellsA {
		for _, cellB := range cellsB {
			distance := gl.manhattanDistance(cellA.Xpos, cellA.Ypos, cellB.Xpos, cellB.Ypos)
		
			if distance == 1 && distance < shortestDistance {
				connection := &BridgeConnection{
					FromX:    cellA.Xpos,
					FromY:    cellA.Ypos,
					ToX:      cellB.Xpos,
					ToY:      cellB.Ypos,
					Distance: distance,
				}
				
				
				if gl.isValidBridgeConnection(grid, connection) {
					bestConnection = connection
					shortestDistance = distance
				}
			}
		}
	}
	
	return bestConnection
}

func (gl *GridLoader) getSegmentCells(grid *domainmodels.Grid, segmentIDs []int64) []*domainmodels.Cell {
	segmentSet := make(map[int64]bool)
	for _, id := range segmentIDs {
		segmentSet[id] = true
	}
	
	var cells []*domainmodels.Cell
	cellsSeen := make(map[[2]int64]bool) 
	
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		
		for _, cellRoad := range cell.RoadSegments {
			if segmentSet[cellRoad.RoadSegment.ID] {
				coords := [2]int64{cell.Xpos, cell.Ypos}
				if !cellsSeen[coords] {
					cells = append(cells, cell)
					cellsSeen[coords] = true
				}
				break
			}
		}
	}
	
	return cells
}

func (gl *GridLoader) isValidBridgeConnection(grid *domainmodels.Grid, connection *BridgeConnection) bool {
	fromCell := gl.getCellAt(grid, connection.FromX, connection.FromY)
	toCell := gl.getCellAt(grid, connection.ToX, connection.ToY)
	
	if fromCell == nil || toCell == nil {
		return false
	}
	
	if fromCell.CellType == domainmodels.CellTypeBlocked || toCell.CellType == domainmodels.CellTypeBlocked {
		return false
	}
	
	if len(fromCell.RoadSegments) == 0 || len(toCell.RoadSegments) == 0 {
		return false
	}
	
	
	if gl.connectionExists(grid, connection.FromX, connection.FromY, connection.ToX, connection.ToY) {
		return false
	}
	
	dx := connection.ToX - connection.FromX
	dy := connection.ToY - connection.FromY
	
	if (dx == 0 && (dy == 1 || dy == -1)) || (dy == 0 && (dx == 1 || dx == -1)) {
		return true 
	}
	
	return false
}

func (gl *GridLoader) createBridgeSegment(grid *domainmodels.Grid, connection *BridgeConnection) bool {
	segment := domainmodels.RoadSegment{
		ID:       gl.SegmentIDCounter,
		StartX:   connection.FromX,
		StartY:   connection.FromY,
		EndX:     connection.ToX,
		EndY:     connection.ToY,
		IsOpen:   true,
		Capacity: gl.getDefaultCapacityForSegment(),
	}
	
	
	gl.addSegmentToCell(grid, connection.FromX, connection.FromY, segment)
	gl.addSegmentToCell(grid, connection.ToX, connection.ToY, segment)
	
	gl.SegmentIDCounter++
	return true
}


func (gl *GridLoader) validateConnectedAccessibility(grid *domainmodels.Grid) error {
	var specialLocations []*domainmodels.Cell
	
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		if cell.CellType == domainmodels.CellTypeRefuel || cell.CellType == domainmodels.CellTypeDepot {
			specialLocations = append(specialLocations, cell)
		}
	}
	
	for _, location := range specialLocations {
		if len(location.RoadSegments) == 0 {
			return fmt.Errorf("special location at (%d,%d) type %s has no road access",
				location.Xpos, location.Ypos, location.CellType)
		}
	}
	
	return nil
}


func (gl *GridLoader) generateConnectivityReport(grid *domainmodels.Grid) *ConnectivityReport {
	components := gl.findConnectedComponents(grid)
	
	report := &ConnectivityReport{
		TotalSegments:       gl.countRoadSegments(grid),
		ConnectedComponents: len(components),
		IsFullyConnected:    len(components) <= 1,
		ComponentSizes:      make([]int, len(components)),
	}
	
	for i, component := range components {
		report.ComponentSizes[i] = len(component)
	}
	
	
	for _, cell := range grid.Cells {
		switch cell.CellType {
		case domainmodels.CellTypeRefuel:
			report.AccessibleFuelStations++
		case domainmodels.CellTypeDepot:
			report.AccessibleDepots++
		}
	}
	
	return report
}

type ConnectivityReport struct {
	TotalSegments            int     `json:"total_segments"`
	ConnectedComponents      int     `json:"connected_components"`
	IsFullyConnected         bool    `json:"is_fully_connected"`
	ComponentSizes           []int   `json:"component_sizes"`
	AccessibleFuelStations   int     `json:"accessible_fuel_stations"`
	AccessibleDepots         int     `json:"accessible_depots"`
	NetworkDensity           float64 `json:"network_density"`
	AverageConnectivity      float64 `json:"average_connectivity"`
}