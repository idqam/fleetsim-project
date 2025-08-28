package gridengine

import (
	"math"
	"math/rand"
	"sort"

	"owenvi.com/simsim/internal/coremodels"
)


type LatticeParams struct {
	BaseParams
	CellSize     float64
	DeleteProb   float64
	AddDiagonals bool
	TwoWay       bool
}

type RadialParams struct {
	BaseParams
	NumRays      int
	NumRings     int
	RingSpacing  float64
}

type SpaceColonizationParams struct {
	BaseParams
	Attractions int
	StepSize    float64
	CaptureRadius float64
}

type KNNMeshParams struct {
	BaseParams
	Sites int
	K     int
}

type RandomParams struct {
	BaseParams
	NodeCount int
	ExtraEdges int
}

type LorenzAttractorParams struct {
	BaseParams
	NumSteps  int
	StepSize  float64
	Sigma     float64
	Rho       float64
	Beta      float64
	ScaleX    float64
	ScaleY    float64
}

type LSystemParams struct {
	BaseParams
	Axiom      string
	Rules      map[rune]string
	Iterations int
	Angle      float64
	Length     float64
}

func GenerateLattice(g *coremodels.Grid, r *rand.Rand, p LatticeParams) {
	counter := &NodeSegmentCounters{}
	
	
	nodeGrid := AddNodesGrid(g, g.DimY+1, g.DimX+1, p.CellSize, p.JitterMax, r, counter)
	
	
	for y := int64(0); y <= g.DimY; y++ {
		for x := int64(0); x <= g.DimX; x++ {
			u := nodeGrid[y][x]
			
			
			if x < g.DimX && r.Float64() > p.DeleteProb {
				v := nodeGrid[y][x+1]
				AddSegmentWithCounter(g, u, v, 1.0, counter)
			}
			
			 
			if y < g.DimY && r.Float64() > p.DeleteProb {
				v := nodeGrid[y+1][x]
				AddSegmentWithCounter(g, u, v, 1.0, counter)
			}
			
			
			if p.AddDiagonals {
				if x < g.DimX && y < g.DimY && r.Float64() > p.DeleteProb {
					v := nodeGrid[y+1][x+1]
					AddSegmentWithCounter(g, u, v, 1.0, counter)
				}
				if x < g.DimX && y > 0 && r.Float64() > p.DeleteProb {
					v := nodeGrid[y-1][x+1] 
					AddSegmentWithCounter(g, u, v, 1.0, counter)
				}
			}
		}
	}
}

func GenerateRadial(g *coremodels.Grid, r *rand.Rand, p RadialParams) {
	counter := &NodeSegmentCounters{}
	
	centerIdx, ringNodes := AddNodesRadial(g, p.CenterX, p.CenterY, p.NumRays, p.NumRings, p.RingSpacing, p.JitterMax, r, counter)
	
	
	rayNodes := make([][]int64, p.NumRays)
	for i := 0; i < p.NumRays; i++ {
		rayNodes[i] = append(rayNodes[i], centerIdx)
		for ring := 0; ring < p.NumRings; ring++ {
			nodeIdx := ringNodes[ring][i]
			rayNodes[i] = append(rayNodes[i], nodeIdx)
			
			
			if len(rayNodes[i]) > 1 {
				prev := rayNodes[i][len(rayNodes[i])-2]
				AddSegmentWithCounter(g, prev, nodeIdx, 1.0, counter)
			}
		}
	}
	
	
	for _, ringNodeList := range ringNodes {
		for i := 0; i < len(ringNodeList); i++ {
			u := ringNodeList[i]
			v := ringNodeList[(i+1)%len(ringNodeList)]
			AddSegmentWithCounter(g, u, v, 1.0, counter)
		}
	}
}

func GenerateSpaceColonization(g *coremodels.Grid, r *rand.Rand, p SpaceColonizationParams) {
    g.Nodes = make(map[int64]*coremodels.Node)
    g.Segments = make(map[int64]*coremodels.RoadSegment)
    g.Adjacency = make(map[int64][]int64)
    
    counter := &NodeSegmentCounters{}
    center := AddNodeWithCounter(g, p.CenterX, p.CenterY, counter)
    
    attractions := make([]Point, p.Attractions)
    
    
    for i := 0; i < min(20, p.Attractions); i++ {
        angle := r.Float64() * 2 * math.Pi
        dist := r.Float64() * p.CaptureRadius * 0.8
        attractions[i] = Point{
            X: p.CenterX + math.Cos(angle) * dist,
            Y: p.CenterY + math.Sin(angle) * dist,
            Index: i, Alive: true,
        }
    }
    
    gridCells := 10
    cellW := p.BoxWidth / float64(gridCells)
    cellH := p.BoxHeight / float64(gridCells)
    attrsPerCell := p.Attractions / (gridCells * gridCells)
    
    idx := min(20, p.Attractions)
    for gy := 0; gy < gridCells && idx < p.Attractions; gy++ {
        for gx := 0; gx < gridCells && idx < p.Attractions; gx++ {
            for c := 0; c < attrsPerCell && idx < p.Attractions; c++ {
                x := p.CenterX - p.BoxWidth/2 + float64(gx)*cellW + r.Float64()*cellW
                y := p.CenterY - p.BoxHeight/2 + float64(gy)*cellH + r.Float64()*cellH
                attractions[idx] = Point{X: x, Y: y, Index: idx, Alive: true}
                idx++
            }
        }
    }
    
    for i := idx; i < p.Attractions; i++ {
        gx := r.Intn(gridCells)
        gy := r.Intn(gridCells)
        x := p.CenterX - p.BoxWidth/2 + float64(gx)*cellW + r.Float64()*cellW
        y := p.CenterY - p.BoxHeight/2 + float64(gy)*cellH + r.Float64()*cellH
        attractions[i] = Point{X: x, Y: y, Index: i, Alive: true}
    }
    
    spatialGrid := NewSpatialGrid(attractions, p.CaptureRadius, p.BaseParams)
    
    frontier := []int64{center}
    aliveCount := p.Attractions
    maxIterations := p.Attractions * 3
    iteration := 0
    
    for aliveCount > 0 && len(frontier) > 0 && iteration < maxIterations {
        newFrontier := make([]int64, 0)
        
        for _, u := range frontier {
            n := g.Nodes[u]
            
            nearbyIndices := spatialGrid.QueryRadius(n.Pos_X, n.Pos_Y, p.CaptureRadius)
            
            closest := -1
            bestDist := p.CaptureRadius
            
            for _, attrIdx := range nearbyIndices {
                if !attractions[attrIdx].Alive {
                    continue
                }
                
                dx := attractions[attrIdx].X - n.Pos_X
                dy := attractions[attrIdx].Y - n.Pos_Y
                dist := math.Sqrt(dx*dx + dy*dy)
                
                if dist < bestDist {
                    bestDist = dist
                    closest = attrIdx
                }
            }
            
            if closest != -1 {
                attr := attractions[closest]
                dx := attr.X - n.Pos_X
                dy := attr.Y - n.Pos_Y
                dist := math.Sqrt(dx*dx + dy*dy)
                
                stepRatio := p.StepSize / dist
                nx := n.Pos_X + dx * stepRatio
                ny := n.Pos_Y + dy * stepRatio
                
                v := AddNodeWithCounter(g, nx, ny, counter)
                AddSegmentWithCounter(g, u, v, 1.0, counter)
                newFrontier = append(newFrontier, v)
                
                if dist <= p.StepSize * 1.2 {
                    attractions[closest].Alive = false
                    aliveCount--
                }
            }
        }
        
        if len(newFrontier) > 0 {
            frontier = newFrontier
        } else {
            frontier = make([]int64, 0)
            for nodeID := range g.Nodes {
                n := g.Nodes[nodeID]
                for _, attr := range attractions {
                    if !attr.Alive {
                        continue
                    }
                    dx := attr.X - n.Pos_X
                    dy := attr.Y - n.Pos_Y
                    if dx*dx + dy*dy <= p.CaptureRadius * p.CaptureRadius {
                        frontier = append(frontier, nodeID)
                        break
                    }
                }
                if len(frontier) > 20 {
                    break
                }
            }
        }
        
        iteration++
        
        if iteration%15 == 0 {
            spatialGrid.UpdatePoints(attractions)
        }
    }
}

func GenerateKNNMesh(g *coremodels.Grid, r *rand.Rand, p KNNMeshParams) {
	counter := &NodeSegmentCounters{}
	
	
	siteNodes := make([]int64, p.Sites)
	for i := 0; i < p.Sites; i++ {
		x := p.CenterX - p.BoxWidth/2 + r.Float64()*p.BoxWidth
		y := p.CenterY - p.BoxHeight/2 + r.Float64()*p.BoxHeight
		siteNodes[i] = AddNodeWithCounter(g, x, y, counter)
	}
	
	
	type neighbor struct {
		nodeID int64
		dist2  float64
	}
	
	for _, u := range siteNodes {
		var neighbors []neighbor
		nu := g.Nodes[u]
		
		for _, v := range siteNodes {
			if v == u {
				continue
			}
			nv := g.Nodes[v]
			dx := nu.Pos_X - nv.Pos_X
			dy := nu.Pos_Y - nv.Pos_Y
			d2 := dx*dx + dy*dy
			neighbors = append(neighbors, neighbor{nodeID: v, dist2: d2})
		}
		
		sort.Slice(neighbors, func(i, j int) bool {
			return neighbors[i].dist2 < neighbors[j].dist2
		})
		
		limit := min(p.K, len(neighbors))
		for i := 0; i < limit; i++ {
			v := neighbors[i].nodeID
			AddSegmentWithCounter(g, u, v, 1.0, counter)
		}
	}
}

func GenerateRandom(g *coremodels.Grid, r *rand.Rand, p RandomParams) {
	counter := &NodeSegmentCounters{}
	
	
	nodes := make([]int64, p.NodeCount)
	for i := 0; i < p.NodeCount; i++ {
		x := p.CenterX - p.BoxWidth/2 + r.Float64()*p.BoxWidth
		y := p.CenterY - p.BoxHeight/2 + r.Float64()*p.BoxHeight  
		nodes[i] = AddNodeWithCounter(g, x, y, counter)
	}
	
	
	for i := 1; i < len(nodes); i++ {
		AddSegmentWithCounter(g, nodes[i-1], nodes[i], 1.0, counter)
	}
	
	
	for i := 0; i < p.ExtraEdges; i++ {
		u := nodes[r.Intn(len(nodes))]
		v := nodes[r.Intn(len(nodes))]
		if u == v {
			continue
		}
		congestion := 1.0 + r.Float64()*0.25
		AddSegmentWithCounter(g, u, v, congestion, counter)
	}
}

func GenerateLorenzAttractor(g *coremodels.Grid, r *rand.Rand, p LorenzAttractorParams) {
	counter := &NodeSegmentCounters{}
	
	
	x, y, z := r.Float64()*10-5, r.Float64()*10-5, r.Float64()*10-5
	
	prevNode := AddNodeWithCounter(g, x*p.ScaleX+p.CenterX, y*p.ScaleY+p.CenterY, counter)
	
	
	for i := 0; i < p.NumSteps; i++ {
		
		dx := p.Sigma * (y - x)
		dy := x*(p.Rho-z) - y
		dz := x*y - p.Beta*z
		
		
		k1x, k1y, k1z := dx, dy, dz
		
		x2, y2, z2 := x+p.StepSize*k1x/2, y+p.StepSize*k1y/2, z+p.StepSize*k1z/2
		k2x := p.Sigma * (y2 - x2)
		k2y := x2*(p.Rho-z2) - y2
		k2z := x2*y2 - p.Beta*z2
		
		x3, y3, z3 := x+p.StepSize*k2x/2, y+p.StepSize*k2y/2, z+p.StepSize*k2z/2
		k3x := p.Sigma * (y3 - x3)
		k3y := x3*(p.Rho-z3) - y3
		k3z := x3*y3 - p.Beta*z3
		
		x4, y4, z4 := x+p.StepSize*k3x, y+p.StepSize*k3y, z+p.StepSize*k3z
		k4x := p.Sigma * (y4 - x4)
		k4y := x4*(p.Rho-z4) - y4
		k4z := x4*y4 - p.Beta*z4
		
		x += (p.StepSize/6) * (k1x + 2*k2x + 2*k3x + k4x)
		y += (p.StepSize/6) * (k1y + 2*k2y + 2*k3y + k4y)
		z += (p.StepSize/6) * (k1z + 2*k2z + 2*k3z + k4z)
		
		currentNode := AddNodeWithCounter(g, x*p.ScaleX+p.CenterX, y*p.ScaleY+p.CenterY, counter)
		AddSegmentWithCounter(g, prevNode, currentNode, 1.0, counter)
		prevNode = currentNode
	}
}

func GenerateLSystem(g *coremodels.Grid, r *rand.Rand, p LSystemParams) {
	counter := &NodeSegmentCounters{}
	
	type turtle struct {
		x, y  float64
		angle float64
	}
	
	
	current := p.Axiom
	for i := 0; i < p.Iterations; i++ {
		next := ""
		for _, char := range current {
			if rule, ok := p.Rules[char]; ok {
				next += rule
			} else {
				next += string(char)
			}
		}
		current = next
	}
	
	state := turtle{x: p.CenterX, y: p.CenterY, angle: -math.Pi / 2}
	stack := []turtle{}
	
	prevNode := AddNodeWithCounter(g, state.x, state.y, counter)
	
	for _, char := range current {
		switch char {
		case 'F', 'G':
			
			newX := state.x + p.Length*math.Cos(state.angle)
			newY := state.y + p.Length*math.Sin(state.angle)
			
			currentNode := AddNodeWithCounter(g, newX, newY, counter)
			AddSegmentWithCounter(g, prevNode, currentNode, 1.0, counter)
			
			state.x, state.y = newX, newY
			prevNode = currentNode
			
		case '+':
			state.angle += p.Angle * math.Pi / 180
			
		case '-':
			state.angle -= p.Angle * math.Pi / 180
			
		case '[':
			stack = append(stack, state)
			
		case ']':
			if len(stack) > 0 {
				state = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				
				
				for nodeID, node := range g.Nodes {
					if math.Abs(node.Pos_X-state.x) < 0.001 && math.Abs(node.Pos_Y-state.y) < 0.001 {
						prevNode = nodeID
						break
					}
				}
			}
		}
	}
}

type HierarchicalParams struct {
	BaseParams
	MajorCellSize   float64
	LocalCellSize   float64
	MajorDeleteProb float64
	LocalDeleteProb float64
}

type CityLikeParams struct {
	BaseParams
	NumRays      int
	NumRings     int
	RingSpacing  float64
}

type SuburbanParams struct {
	BaseParams
	CellSize     float64
	DeleteProb   float64
	AddDiagonals bool
}

func GenerateHierarchical(g *coremodels.Grid, r *rand.Rand, p HierarchicalParams) {
	counter := &NodeSegmentCounters{}
	
	majorNodeGrid := AddNodesGrid(g, int64(p.BoxHeight/p.MajorCellSize)+1, int64(p.BoxWidth/p.MajorCellSize)+1, p.MajorCellSize, p.JitterMax, r, counter)
	
	for y := 0; y < len(majorNodeGrid); y++ {
		for x := 0; x < len(majorNodeGrid[y]); x++ {
			u := majorNodeGrid[y][x]
			
			if x < len(majorNodeGrid[y])-1 && r.Float64() > p.MajorDeleteProb {
				v := majorNodeGrid[y][x+1]
				AddSegmentWithCounter(g, u, v, 1.0, counter)
			}
			
			if y < len(majorNodeGrid)-1 && r.Float64() > p.MajorDeleteProb {
				v := majorNodeGrid[y+1][x]
				AddSegmentWithCounter(g, u, v, 1.0, counter)
			}
		}
	}
	
	localNodeGrid := AddNodesGrid(g, int64(p.BoxHeight/p.LocalCellSize)+1, int64(p.BoxWidth/p.LocalCellSize)+1, p.LocalCellSize, p.JitterMax/2, r, counter)
	
	for y := 0; y < len(localNodeGrid); y++ {
		for x := 0; x < len(localNodeGrid[y]); x++ {
			u := localNodeGrid[y][x]
			
			if x < len(localNodeGrid[y])-1 && r.Float64() > p.LocalDeleteProb {
				v := localNodeGrid[y][x+1]
				AddSegmentWithCounter(g, u, v, 0.8, counter)
			}
			
			if y < len(localNodeGrid)-1 && r.Float64() > p.LocalDeleteProb {
				v := localNodeGrid[y+1][x]
				AddSegmentWithCounter(g, u, v, 0.8, counter)
			}
			
			if x < len(localNodeGrid[y])-1 && y < len(localNodeGrid)-1 && r.Float64() > p.LocalDeleteProb+0.2 {
				v := localNodeGrid[y+1][x+1]
				AddSegmentWithCounter(g, u, v, 0.7, counter)
			}
		}
	}
}

func GenerateCityLike(g *coremodels.Grid, r *rand.Rand, p CityLikeParams) {
	counter := &NodeSegmentCounters{}
	
	centerIdx, ringNodes := AddNodesRadial(g, p.CenterX, p.CenterY, p.NumRays, p.NumRings, p.RingSpacing, p.JitterMax, r, counter)
	
	rayNodes := make([][]int64, p.NumRays)
	for i := 0; i < p.NumRays; i++ {
		rayNodes[i] = append(rayNodes[i], centerIdx)
		for ring := 0; ring < p.NumRings; ring++ {
			nodeIdx := ringNodes[ring][i]
			rayNodes[i] = append(rayNodes[i], nodeIdx)
			
			if len(rayNodes[i]) > 1 {
				prev := rayNodes[i][len(rayNodes[i])-2]
				AddSegmentWithCounter(g, prev, nodeIdx, 1.2, counter)
			}
		}
	}
	
	for ringIndex, ringNodeList := range ringNodes {
		congestion := 1.0 + float64(ringIndex)*0.2
		for i := 0; i < len(ringNodeList); i++ {
			u := ringNodeList[i]
			v := ringNodeList[(i+1)%len(ringNodeList)]
			AddSegmentWithCounter(g, u, v, congestion, counter)
		}
	}
	
	for ring := 1; ring < p.NumRings; ring++ {
		for ray := 0; ray < p.NumRays; ray++ {
			if r.Float64() < 0.3 {
				u := ringNodes[ring][ray]
				nextRay := (ray + 1) % p.NumRays
				v := ringNodes[ring][nextRay]
				AddSegmentWithCounter(g, u, v, 0.9, counter)
			}
		}
	}
}

func GenerateSuburban(g *coremodels.Grid, r *rand.Rand, p SuburbanParams) {
	counter := &NodeSegmentCounters{}
	
	nodeGrid := AddNodesGrid(g, int64(p.BoxHeight/p.CellSize)+1, int64(p.BoxWidth/p.CellSize)+1, p.CellSize, p.JitterMax, r, counter)
	
	for y := 0; y < len(nodeGrid); y++ {
		for x := 0; x < len(nodeGrid[y]); x++ {
			u := nodeGrid[y][x]
			
			if x < len(nodeGrid[y])-1 && r.Float64() > p.DeleteProb {
				v := nodeGrid[y][x+1]
				congestion := 1.0 + r.Float64()*0.3
				AddSegmentWithCounter(g, u, v, congestion, counter)
			}
			
			if y < len(nodeGrid)-1 && r.Float64() > p.DeleteProb {
				v := nodeGrid[y+1][x]
				congestion := 1.0 + r.Float64()*0.3
				AddSegmentWithCounter(g, u, v, congestion, counter)
			}
			
			if p.AddDiagonals {
				if x < len(nodeGrid[y])-1 && y < len(nodeGrid)-1 && r.Float64() > p.DeleteProb+0.3 {
					v := nodeGrid[y+1][x+1]
					congestion := 0.8 + r.Float64()*0.2
					AddSegmentWithCounter(g, u, v, congestion, counter)
				}
				
				if x < len(nodeGrid[y])-1 && y > 0 && r.Float64() > p.DeleteProb+0.3 {
					v := nodeGrid[y-1][x+1]
					congestion := 0.8 + r.Float64()*0.2
					AddSegmentWithCounter(g, u, v, congestion, counter)
				}
			}
			
			if r.Float64() < 0.1 {
				for i := 0; i < 3; i++ {
					angle := r.Float64() * 2 * math.Pi
					dist := p.CellSize * (0.5 + r.Float64()*0.5)
					nx := g.Nodes[u].Pos_X + math.Cos(angle)*dist
					ny := g.Nodes[u].Pos_Y + math.Sin(angle)*dist
					cul := AddNodeWithCounter(g, nx, ny, counter)
					AddSegmentWithCounter(g, u, cul, 0.6, counter)
				}
			}
		}
	}
}