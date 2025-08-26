package runtime

import "owenvi.com/fleetsim/internal/domainmodels"

type MoveRequest struct {
	VehicleID string
	FromCell  *domainmodels.Cell
	ToCell    *domainmodels.Cell
	SegmentID int64
	SimTime   int64
}

type MovementValidator struct {
	grid *domainmodels.Grid
}

func NewMovementValidator(grid *domainmodels.Grid) *MovementValidator {
	return &MovementValidator{grid: grid}
}

func (mv *MovementValidator) ValidateMove(request MoveRequest) bool {
	if request.FromCell == nil || request.ToCell == nil {
		return false
	}

	if request.ToCell.CellType == domainmodels.CellTypeBlocked {
		return false
	}

	return mv.isConnected(request.FromCell, request.ToCell)
}

func (mv *MovementValidator) isConnected(from, to *domainmodels.Cell) bool {
	for _, cellRoad := range from.RoadSegments {
		segment := cellRoad.RoadSegment
		
		if (segment.StartX == from.Xpos && segment.StartY == from.Ypos &&
			segment.EndX == to.Xpos && segment.EndY == to.Ypos) ||
		   (segment.StartX == to.Xpos && segment.StartY == to.Ypos &&
			segment.EndX == from.Xpos && segment.EndY == from.Ypos) {
			return true
		}
	}
	return false
}

func (mv *MovementValidator) GetConnectedCells(cell *domainmodels.Cell) []*domainmodels.Cell {
	var connected []*domainmodels.Cell
	
	for _, cellRoad := range cell.RoadSegments {
		segment := cellRoad.RoadSegment
		var targetX, targetY int64
		
		if segment.StartX == cell.Xpos && segment.StartY == cell.Ypos {
			targetX, targetY = segment.EndX, segment.EndY
		} else if segment.EndX == cell.Xpos && segment.EndY == cell.Ypos {
			targetX, targetY = segment.StartX, segment.StartY
		} else {
			continue
		}
		
		if targetCell := mv.getCellAt(targetX, targetY); targetCell != nil {
			connected = append(connected, targetCell)
		}
	}
	
	return connected
}

func (mv *MovementValidator) getCellAt(x, y int64) *domainmodels.Cell {
	coords := [2]int64{x, y}
	return mv.grid.CoordIndex[coords]
}