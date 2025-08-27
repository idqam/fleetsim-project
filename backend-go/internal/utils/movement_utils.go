package utils

import "owenvi.com/fleetsim/internal/domainmodels"

func SegmentIsConnected(from, to *domainmodels.Cell) bool {
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
