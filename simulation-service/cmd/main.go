package main

import (
	"fmt"
	"log"

	"github.com/segmentio/ksuid"
	"owenvi.com/simsim/internal/coremodels"
	"owenvi.com/simsim/internal/gridengine"
)

func main() {
	seed := ksuid.New()
	fmt.Println(seed)
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))

	
	spaceGrid := gridengine.NewGrid(gridengine.WithDimensions(50,50), gridengine.WithAlgorithm(coremodels.Suburban), gridengine.WithSeed(seed))
	

	if err := gridengine.PlotGrid(spaceGrid, "space_grid.svg"); err != nil {
		log.Fatal(err)
	}
	
	
	log.Println("Space Colonization grid plotted to grid_space.svg")
}