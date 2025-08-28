package gridengine

import (
	"github.com/segmentio/ksuid"
	"owenvi.com/simsim/internal/coremodels"
)

type GridOption func(*coremodels.GridConfig)

type GridEngine struct{}

func NewGrid(opts ...GridOption) *coremodels.Grid {
	cfg := &coremodels.GridConfig{
		DimX: 10,
		DimY: 10,
		Algo: coremodels.Varonoi,
		Seed: ksuid.New(),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	g := &coremodels.Grid{
		ID:        cfg.Seed,
		DimX:      cfg.DimX,
		DimY:      cfg.DimY,
		Segments:  make(map[int64]*coremodels.RoadSegment),
		Adjacency: make(map[int64][]int64),
		Nodes:     make(map[int64]*coremodels.Node),
	}

	r := NewRandFromSeed(*cfg)
	
	baseParams := BaseParams{
		BoxWidth:  float64(cfg.DimX) * 100.0,
		BoxHeight: float64(cfg.DimY) * 100.0,
		CenterX:   float64(cfg.DimX) * 50,
		CenterY:   float64(cfg.DimY) * 50,
		JitterMax: 12.0,
	}

	switch cfg.Algo {
	case coremodels.Varonoi:
		GenerateKNNMesh(g, r, KNNMeshParams{
			BaseParams: baseParams,
			Sites:      int(4*(cfg.DimX+cfg.DimY)/2 + 12),
			K:          5,
		})
		
	case coremodels.LForm:
		GenerateLattice(g, r, LatticeParams{
			BaseParams:   baseParams,
			CellSize:    80.0,
			DeleteProb:   0.1,
			AddDiagonals: false,
			TwoWay:       true,
		})
		
		GenerateRadial(g, r, RadialParams{
			BaseParams:  baseParams,
			NumRays:     8,
			NumRings:    2,
			RingSpacing: 120.0,
		})
		
	case coremodels.Space:
		GenerateSpaceColonization(g, r, SpaceColonizationParams{
        BaseParams:    BaseParams{
            BoxWidth:  float64(cfg.DimX) * 80.0,  
            BoxHeight: float64(cfg.DimY) * 80.0,  
            CenterX:   float64(cfg.DimX) * 40,    
            CenterY:   float64(cfg.DimY) * 40,
            JitterMax: 15.0,
        },
		Attractions:   int(float64(cfg.DimX*cfg.DimY) * 0.6), 
        StepSize:      25,   
        CaptureRadius: 80,   
    })
	case coremodels.CityLike:
    GenerateRadial(g, r, RadialParams{
        BaseParams:  baseParams,
        NumRays:     8,
        NumRings:    4,
        RingSpacing: 200.0,
    })
	case coremodels.Hierarchical:
    
    GenerateLattice(g, r, LatticeParams{
        BaseParams:   baseParams,
        CellSize:     200.0,  
        DeleteProb:   0.1,    
        AddDiagonals: false,
    })

    
    
    GenerateLattice(g, r, LatticeParams{
        BaseParams:   baseParams,
        CellSize:    60.0, 
        DeleteProb:   0.3,    
        AddDiagonals: true,
    })
	case coremodels.Suburban:
    GenerateLattice(g, r, LatticeParams{
        BaseParams:   baseParams,
        CellSize:     80.0,
        DeleteProb:   0,  
        AddDiagonals: false,
    })
		
	case coremodels.Lorenz:
		GenerateLorenzAttractor(g, r, LorenzAttractorParams{
			BaseParams: baseParams,
			NumSteps:   5000,
			StepSize:   0.01,
			Sigma:      10.0,
			Rho:        28.0,
			Beta:       8.0 / 3.0,
			ScaleX:     10.0,
			ScaleY:     10.0,
		})
		
	case coremodels.LSystem:
		GenerateLSystem(g, r, LSystemParams{
			BaseParams: baseParams,
			Axiom:      "F",
			Rules:      map[rune]string{'F': "F[-F][+F]"},
			Iterations: 4,
			Angle:      25.0,
			Length:     20.0,
		})
		
	default:
		GenerateRandom(g, r, RandomParams{
			BaseParams: baseParams,
			NodeCount:  int(cfg.DimX*cfg.DimY/2 + 8),
			ExtraEdges: int(cfg.DimX + cfg.DimY),
		})
	}
	
	return g
}

func WithDimensions(x, y int64) GridOption {
	return func(cfg *coremodels.GridConfig) {
		cfg.DimX, cfg.DimY = x, y
	}
}

func WithAlgorithm(algo coremodels.GenerationAlgorithmType) GridOption {
	return func(cfg *coremodels.GridConfig) {
		cfg.Algo = algo
	}
}

func WithSeed(seed ksuid.KSUID) GridOption {
	return func(cfg *coremodels.GridConfig) {
		cfg.Seed = seed
	}
}