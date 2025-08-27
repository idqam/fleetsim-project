package utils

import (
	"owenvi.com/fleetsim/internal/domainmodels"
)

func ManhattanDistance(x1, y1, x2, y2 int64) int64 {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func GetDirectionString(dx, dy int64) string {
	if dx > 0 {
		return "east"
	} else if dx < 0 {
		return "west"
	} else if dy > 0 {
		return "south"
	} else if dy < 0 {
		return "north"
	}
	return "unknown"
}

func CountCellConnections(grid *domainmodels.Grid, cell *domainmodels.Cell) int {
	connections := make(map[string]bool)

	for _, cellRoad := range cell.RoadSegments {
		segment := cellRoad.RoadSegment

		if segment.StartX == cell.Xpos && segment.StartY == cell.Ypos {

			direction := GetDirectionString(segment.EndX-segment.StartX, segment.EndY-segment.StartY)
			connections[direction] = true
		} else if segment.EndX == cell.Xpos && segment.EndY == cell.Ypos {
			direction := GetDirectionString(segment.StartX-segment.EndX, segment.StartY-segment.EndY)
			connections[direction] = true
		}
	}

	return len(connections)
}


func  GetCellAtGrid(grid *domainmodels.Grid, x, y int64) *domainmodels.Cell {
	for i := range grid.Cells {
		cell := &grid.Cells[i]
		if cell.Xpos == x && cell.Ypos == y {
			return cell
		}
	}
	return nil
}