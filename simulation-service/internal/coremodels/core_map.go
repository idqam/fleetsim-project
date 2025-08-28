package coremodels

import (
	"github.com/segmentio/ksuid"
)

type Grid struct {
	ID ksuid.KSUID
	//bounds for n by m grid
	DimX  int64  
	DimY  int64  
	Segments map[int64]*RoadSegment 
	Adjacency map[int64][]int64 //adjacency map for O(1) lookup, nodeID -> list of connected segment ID
	Nodes map[int64]*Node 

	
}
type GenerationAlgorithmType int


const (
    Varonoi GenerationAlgorithmType = iota
    LForm
    Space
    Lorenz
    LSystem
	Hierarchical
	Suburban
	CityLike
)
type GridConfig struct {
    DimX int64
    DimY int64
    Algo GenerationAlgorithmType
    Seed ksuid.KSUID
}


//like intersection 
type Node struct {
	ID int64
	Pos_X, Pos_Y float64 //geo coord

}

type RoadSegment struct {
	ID int64
	StartNode, EndNode int64
	LengthKM float64
	CongestionFactor float64 //ratio 

}

func (g *Grid) getGridAdjacencyMatrix() map[int64][]int64 {
	if g.Adjacency == nil {
		
		return make(map[int64][]int64)
	}

	result := make(map[int64][]int64)
	for nodeID, neighbors := range g.Adjacency{
		neighborsCopy := make([]int64, len(neighbors))
		copy(neighborsCopy, neighbors)
		result[nodeID] = neighborsCopy

	}

	return result
}


func (g *Grid) getRoadSegments() map[int64]*RoadSegment {
    //DO NOT MODIFY SEGMENTS 
    return g.Segments
}


