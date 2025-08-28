package gridloader

import (
	"math/rand"

	"owenvi.com/fleetsim/internal/domainmodels"
)

type undoOp struct {
	cellIndex int
	prevType  domainmodels.CellType
	prevSegs  []domainmodels.CellRoad
}

type DSU struct {
	parent map[int64]int64
}

func newDSU() *DSU {
	return &DSU{parent: make(map[int64]int64)}
}

func (d *DSU) find(x int64) int64 {
	if d.parent[x] != x {
		d.parent[x] = d.find(d.parent[x])
	}
	return d.parent[x]
}

func (d *DSU) union(x, y int64) {
	px, py := d.find(x), d.find(y)
	if px != py {
		d.parent[px] = py
	}
}

func buildDSU(grid *domainmodels.Grid) *DSU {
	d := newDSU()
	for _, cell := range grid.Cells {
		for _, seg := range cell.RoadSegments {
			if _, ok := d.parent[seg.RoadSegment.ID]; !ok {
				d.parent[seg.RoadSegment.ID] = seg.RoadSegment.ID
			}
		}
	}
	for _, cell := range grid.Cells {
		for _, seg := range cell.RoadSegments {
			for _, other := range cell.RoadSegments {
				if seg.RoadSegment.ID != other.RoadSegment.ID {
					d.union(seg.RoadSegment.ID, other.RoadSegment.ID)
				}
			}
		}
	}
	return d
}

func shuffledCandidates(cells []*domainmodels.Cell, rng *rand.Rand) []*domainmodels.Cell {
	shuffled := make([]*domainmodels.Cell, len(cells))
	copy(shuffled, cells)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}

func (gl *GridLoader) quickConnectivityCheck(dsu *DSU, grid *domainmodels.Grid) bool {

	visited := make(map[int64]bool)
	components := 0
	for segID := range dsu.parent {
		root := dsu.find(segID)
		if !visited[root] {
			visited[root] = true
			components++
			if components > 1 {
				return false
			}
		}
	}
	return true
}
