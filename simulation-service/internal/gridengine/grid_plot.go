package gridengine

import (
	"log"
	"math"
	"os"

	svg "github.com/ajstarks/svgo"
	"owenvi.com/simsim/internal/coremodels"
)

func PlotGrid(g *coremodels.Grid, filename string) error {
	
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	
	minX, minY, maxX, maxY := math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, n := range g.Nodes {
		minX = math.Min(minX, n.Pos_X)
		minY = math.Min(minY, n.Pos_Y)
		maxX = math.Max(maxX, n.Pos_X)
		maxY = math.Max(maxY, n.Pos_Y)
	}

	padding := 50.0
	contentWidth := maxX - minX
	contentHeight := maxY - minY
	
	
	canvasWidth := int(contentWidth + padding*2)
	canvasHeight := int(contentHeight + padding*2)

	canvas := svg.New(f)
	canvas.Start(canvasWidth, canvasHeight)

	
	canvas.Rect(0, 0, canvasWidth, canvasHeight, "fill:white")

	
	offsetX := padding - minX
	offsetY := padding - minY

	
	lineStyle := "stroke:gray;stroke-width:2;stroke-linecap:round"
	for _, seg := range g.Segments {
		n1, ok1 := g.Nodes[seg.StartNode]
		n2, ok2 := g.Nodes[seg.EndNode]
		if !ok1 || !ok2 {
			continue
		}

		x1 := int(n1.Pos_X + offsetX)
		y1 := int(n1.Pos_Y + offsetY)
		x2 := int(n2.Pos_X + offsetX)
		y2 := int(n2.Pos_Y + offsetY)

		canvas.Line(x1, y1, x2, y2, lineStyle)
	}

	
	nodeStyle := "fill:red;stroke:black;stroke-width:1"
	for _, n := range g.Nodes {
		x := int(n.Pos_X + offsetX)
		y := int(n.Pos_Y + offsetY)

		canvas.Circle(x, y, 5, nodeStyle)
	}

	
	canvas.End()

	return nil
}
