package gridengine

import (
	"fmt"
	"log"
	"math"
	"os"

	svg "github.com/ajstarks/svgo"
	"owenvi.com/simsim/internal/coremodels"
)

func PlotGridWithVehicles(g *coremodels.Grid, vehicles []*coremodels.Vehicle, filename string) error {
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

	padding := 80.0
	contentWidth := maxX - minX
	contentHeight := maxY - minY

	canvasWidth := int(contentWidth + padding*2)
	canvasHeight := int(contentHeight + padding*2)

	canvas := svg.New(f)
	canvas.Start(canvasWidth, canvasHeight)

	canvas.Rect(0, 0, canvasWidth, canvasHeight, "fill:#f8f9fa")

	canvas.Def()
	canvas.Marker("arrowhead", 4, 0, 4, 4, "auto")
canvas.Path("M 0,0 L 0,4 L 4,2 z", "fill:#666;stroke:#666")
	canvas.MarkerEnd()
	canvas.DefEnd()

	offsetX := padding - minX
	offsetY := padding - minY

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

		congestionColor := getCongestionColor(seg.CongestionFactor)
		style := fmt.Sprintf("stroke:%s;stroke-width:6;stroke-linecap:round;opacity:0.7", congestionColor)
		canvas.Line(x1, y1, x2, y2, style)

		midX := (x1 + x2) / 2
		midY := (y1 + y2) / 2
		canvas.Text(midX, midY-5, fmt.Sprintf("%d", seg.ID),
			"font-family:Arial;font-size:10px;fill:#666;text-anchor:middle")
	}

	intersectionStyle := "fill:#495057;stroke:#343a40;stroke-width:2"
	for nodeID, n := range g.Nodes {
		x := int(n.Pos_X + offsetX)
		y := int(n.Pos_Y + offsetY)

		radius := 8
		if len(g.Adjacency[nodeID]) > 4 {
			radius = 12
		}

		canvas.Circle(x, y, radius, intersectionStyle)
		canvas.Text(x, y+4, fmt.Sprintf("%d", nodeID),
			"font-family:Arial;font-size:9px;fill:white;text-anchor:middle")
	}

	for _, vehicle := range vehicles {
		drawVehicle(canvas, vehicle, g, offsetX, offsetY)
	}

	for _, vehicle := range vehicles {
		if targetNode, exists := g.Nodes[vehicle.TargetNodeID]; exists {
			x := int(targetNode.Pos_X + offsetX)
			y := int(targetNode.Pos_Y + offsetY)

			canvas.Circle(x, y, 15, "fill:none;stroke:#dc3545;stroke-width:3;stroke-dasharray:5,3")
			canvas.Text(x, y-20, "TARGET", "font-family:Arial;font-size:8px;fill:#dc3545;text-anchor:middle;font-weight:bold")
		}
	}

	legend := []struct {
		color  string
		status string
		y      int
	}{
		{"#28a745", "Moving", 30},
		{"#ffc107", "Waiting", 50},
		{"#dc3545", "Stuck/Dead End", 70},
		{"#6f42c1", "Reached Target", 90},
	}

	canvas.Rect(10, 10, 160, 120, "fill:white;stroke:#dee2e6;stroke-width:1;opacity:0.9")
	canvas.Text(20, 25, "Vehicle Status", "font-family:Arial;font-size:12px;fill:#495057;font-weight:bold")

	for _, item := range legend {
		canvas.Circle(25, item.y, 6, fmt.Sprintf("fill:%s;stroke:#343a40;stroke-width:1", item.color))
		canvas.Text(40, item.y+4, item.status, "font-family:Arial;font-size:10px;fill:#495057")
	}

	statsY := 140
	canvas.Rect(10, statsY, 200, 80, "fill:white;stroke:#dee2e6;stroke-width:1;opacity:0.9")
	canvas.Text(20, statsY+15, "Simulation Stats", "font-family:Arial;font-size:12px;fill:#495057;font-weight:bold")

	activeCount := 0
	completedCount := 0
	stuckCount := 0

	for _, vehicle := range vehicles {
		switch vehicle.Status {
		case coremodels.StatusMoving, coremodels.StatusWaitingForPermission:
			activeCount++
		case coremodels.StatusReachedDestination:
			completedCount++
		case coremodels.StatusDeadEnd, coremodels.StatusError:
			stuckCount++
		}
	}

	canvas.Text(20, statsY+35, fmt.Sprintf("Active: %d", activeCount),
		"font-family:Arial;font-size:10px;fill:#28a745")
	canvas.Text(20, statsY+50, fmt.Sprintf("Completed: %d", completedCount),
		"font-family:Arial;font-size:10px;fill:#6f42c1")
	canvas.Text(20, statsY+65, fmt.Sprintf("Stuck: %d", stuckCount),
		"font-family:Arial;font-size:10px;fill:#dc3545")

	canvas.End()
	return nil
}

func PlotGridOnly(g *coremodels.Grid, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	minX, minY, maxX, maxY := math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, n := range g.Nodes {
		minX = math.Min(minX, n.Pos_X)
		minY = math.Min(minY, n.Pos_Y)
		maxX = math.Max(maxX, n.Pos_X)
		maxY = math.Max(maxY, n.Pos_Y)
	}

	padding := 60.0
	contentWidth := maxX - minX
	contentHeight := maxY - minY

	canvasWidth := int(contentWidth + padding*2)
	canvasHeight := int(contentHeight + padding*2)

	canvas := svg.New(f)
	canvas.Start(canvasWidth, canvasHeight)

	canvas.Rect(0, 0, canvasWidth, canvasHeight, "fill:#ffffff")

	offsetX := padding - minX
	offsetY := padding - minY

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

		canvas.Line(x1, y1, x2, y2, "stroke:#333;stroke-width:3;stroke-linecap:round")

		midX := (x1 + x2) / 2
		midY := (y1 + y2) / 2
		canvas.Text(midX, midY-5, fmt.Sprintf("%d", seg.ID),
			"font-family:Arial;font-size:9px;fill:#666;text-anchor:middle")
	}

	for nodeID, n := range g.Nodes {
		x := int(n.Pos_X + offsetX)
		y := int(n.Pos_Y + offsetY)

		radius := 6
		if len(g.Adjacency[nodeID]) > 4 {
			radius = 10
		}

		canvas.Circle(x, y, radius, "fill:#444;stroke:#222;stroke-width:1")
		canvas.Text(x, y+3, fmt.Sprintf("%d", nodeID),
			"font-family:Arial;font-size:8px;fill:white;text-anchor:middle")
	}

	canvas.Text(canvasWidth/2, 30, "Road Network Layout",
		"font-family:Arial;font-size:16px;fill:#333;text-anchor:middle;font-weight:bold")

	canvas.End()
	return nil
}

func PlotTrafficHeatmap(g *coremodels.Grid, vehicles []*coremodels.Vehicle, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	segmentTraffic := make(map[int64]int)
	for _, vehicle := range vehicles {
		segmentTraffic[vehicle.CurrentSegmentID]++
	}

	minX, minY, maxX, maxY := math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, n := range g.Nodes {
		minX = math.Min(minX, n.Pos_X)
		minY = math.Min(minY, n.Pos_Y)
		maxX = math.Max(maxX, n.Pos_X)
		maxY = math.Max(maxY, n.Pos_Y)
	}

	padding := 60.0
	contentWidth := maxX - minX
	contentHeight := maxY - minY

	canvasWidth := int(contentWidth + padding*2)
	canvasHeight := int(contentHeight + padding*2)

	canvas := svg.New(f)
	canvas.Start(canvasWidth, canvasHeight)

	canvas.Rect(0, 0, canvasWidth, canvasHeight, "fill:#f0f0f0")

	offsetX := padding - minX
	offsetY := padding - minY

	maxTraffic := 1
	for _, count := range segmentTraffic {
		if count > maxTraffic {
			maxTraffic = count
		}
	}

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

		traffic := segmentTraffic[seg.ID]
		intensity := float64(traffic) / float64(maxTraffic)

		heatColor := getHeatmapColor(intensity)
		width := 3 + int(intensity*8)

		style := fmt.Sprintf("stroke:%s;stroke-width:%d;stroke-linecap:round;opacity:0.8", heatColor, width)
		canvas.Line(x1, y1, x2, y2, style)

		if traffic > 0 {
			midX := (x1 + x2) / 2
			midY := (y1 + y2) / 2
			canvas.Text(midX, midY, fmt.Sprintf("%d", traffic),
				"font-family:Arial;font-size:10px;fill:#000;text-anchor:middle;font-weight:bold")
		}
	}

	for _, n := range g.Nodes {
		x := int(n.Pos_X + offsetX)
		y := int(n.Pos_Y + offsetY)
		canvas.Circle(x, y, 4, "fill:#333;stroke:#000;stroke-width:1")
	}

	canvas.Text(canvasWidth/2, 30, "Traffic Density Heatmap",
		"font-family:Arial;font-size:16px;fill:#333;text-anchor:middle;font-weight:bold")

	legend := []struct {
		color string
		label string
		y     int
	}{
		{"#00ff00", "Light Traffic", 50},
		{"#ffff00", "Moderate Traffic", 70},
		{"#ff8000", "Heavy Traffic", 90},
		{"#ff0000", "Congested", 110},
	}

	canvas.Rect(10, 35, 140, 90, "fill:white;stroke:#ccc;stroke-width:1;opacity:0.9")
	for _, item := range legend {
		canvas.Circle(20, item.y, 4, fmt.Sprintf("fill:%s;stroke:#000;stroke-width:1", item.color))
		canvas.Text(35, item.y+3, item.label, "font-family:Arial;font-size:9px;fill:#333")
	}

	canvas.End()
	return nil
}

func PlotRoutingPaths(g *coremodels.Grid, vehicles []*coremodels.Vehicle, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	minX, minY, maxX, maxY := math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, n := range g.Nodes {
		minX = math.Min(minX, n.Pos_X)
		minY = math.Min(minY, n.Pos_Y)
		maxX = math.Max(maxX, n.Pos_X)
		maxY = math.Max(maxY, n.Pos_Y)
	}

	padding := 60.0
	contentWidth := maxX - minX
	contentHeight := maxY - minY

	canvasWidth := int(contentWidth + padding*2)
	canvasHeight := int(contentHeight + padding*2)

	canvas := svg.New(f)
	canvas.Start(canvasWidth, canvasHeight)

	canvas.Rect(0, 0, canvasWidth, canvasHeight, "fill:#fafafa")

	offsetX := padding - minX
	offsetY := padding - minY

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

		canvas.Line(x1, y1, x2, y2, "stroke:#ddd;stroke-width:2")
	}

	colors := []string{"#e74c3c", "#3498db", "#2ecc71", "#f39c12", "#9b59b6", "#1abc9c", "#34495e", "#e67e22"}

	for i, vehicle := range vehicles {
		if i >= len(colors) {
			break
		}

		currentX, currentY, err := vehicle.GetCurrentPosition(g)
		if err != nil {
			continue
		}

		targetNode, exists := g.Nodes[vehicle.TargetNodeID]
		if !exists {
			continue
		}

		startX := int(currentX + offsetX)
		startY := int(currentY + offsetY)
		endX := int(targetNode.Pos_X + offsetX)
		endY := int(targetNode.Pos_Y + offsetY)

		color := colors[i%len(colors)]

		canvas.Line(startX, startY, endX, endY,
			fmt.Sprintf("stroke:%s;stroke-width:3;stroke-dasharray:8,4;opacity:0.7", color))

		canvas.Circle(startX, startY, 8, fmt.Sprintf("fill:%s;stroke:#fff;stroke-width:2", color))
		canvas.Circle(endX, endY, 6, fmt.Sprintf("fill:none;stroke:%s;stroke-width:3", color))

		labelX := startX - 15
		labelY := startY - 15
		canvas.Text(labelX, labelY, fmt.Sprintf("V%d", i+1),
			fmt.Sprintf("font-family:Arial;font-size:10px;fill:%s;font-weight:bold", color))
	}

	for _, n := range g.Nodes {
		x := int(n.Pos_X + offsetX)
		y := int(n.Pos_Y + offsetY)
		canvas.Circle(x, y, 3, "fill:#666")
	}

	canvas.Text(canvasWidth/2, 30, "Vehicle Routing Paths",
		"font-family:Arial;font-size:16px;fill:#333;text-anchor:middle;font-weight:bold")

	canvas.End()
	return nil
}

func CreateComparisonView(g *coremodels.Grid, vehicles []*coremodels.Vehicle, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	minX, minY, maxX, maxY := math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, n := range g.Nodes {
		minX = math.Min(minX, n.Pos_X)
		minY = math.Min(minY, n.Pos_Y)
		maxX = math.Max(maxX, n.Pos_X)
		maxY = math.Max(maxY, n.Pos_Y)
	}

	padding := 40.0
	contentWidth := maxX - minX
	contentHeight := maxY - minY

	viewWidth := int(contentWidth + padding*2)
	viewHeight := int(contentHeight + padding*2)
	canvasWidth := viewWidth * 2 + 40
	canvasHeight := viewHeight + 60

	canvas := svg.New(f)
	canvas.Start(canvasWidth, canvasHeight)

	canvas.Rect(0, 0, canvasWidth, canvasHeight, "fill:#f8f9fa")

	offsetX := padding - minX
	offsetY := padding - minY + 40

	canvas.Text(viewWidth/2, 25, "Network Layout",
		"font-family:Arial;font-size:14px;fill:#333;text-anchor:middle;font-weight:bold")
	canvas.Text(viewWidth+20+viewWidth/2, 25, "Live Traffic",
		"font-family:Arial;font-size:14px;fill:#333;text-anchor:middle;font-weight:bold")

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

		canvas.Line(x1, y1, x2, y2, "stroke:#666;stroke-width:2")

		rightX1 := x1 + viewWidth + 20
		rightY1 := y1
		rightX2 := x2 + viewWidth + 20
		rightY2 := y2

		congestionColor := getCongestionColor(seg.CongestionFactor)
		style := fmt.Sprintf("stroke:%s;stroke-width:4;opacity:0.8", congestionColor)
		canvas.Line(rightX1, rightY1, rightX2, rightY2, style)
	}

	for _, n := range g.Nodes {
		x := int(n.Pos_X + offsetX)
		y := int(n.Pos_Y + offsetY)
		canvas.Circle(x, y, 4, "fill:#444")
		canvas.Circle(x+viewWidth+20, y, 4, "fill:#444")
	}

	for _, vehicle := range vehicles {
		x, y, err := vehicle.GetCurrentPosition(g)
		if err != nil {
			continue
		}

		canvasX := int(x + offsetX) + viewWidth + 20
		canvasY := int(y + offsetY)

		color := getVehicleColor(vehicle.Status)
		canvas.Circle(canvasX, canvasY, 6, fmt.Sprintf("fill:%s;stroke:#fff;stroke-width:1", color))
	}

	canvas.End()
	return nil
}

func ValidateVisualization(g *coremodels.Grid, vehicles []*coremodels.Vehicle) []string {
	var warnings []string

	if len(g.Nodes) == 0 {
		warnings = append(warnings, "Grid has no nodes")
	}

	if len(g.Segments) == 0 {
		warnings = append(warnings, "Grid has no segments")
	}

	vehicleCount := len(vehicles)
	if vehicleCount == 0 {
		warnings = append(warnings, "No vehicles to visualize")
	}

	segmentCount := len(g.Segments)
	if vehicleCount > segmentCount*3 {
		warnings = append(warnings, fmt.Sprintf("High vehicle density: %d vehicles on %d segments", vehicleCount, segmentCount))
	}

	for i, vehicle := range vehicles {
		if _, exists := g.Segments[vehicle.CurrentSegmentID]; !exists {
			warnings = append(warnings, fmt.Sprintf("Vehicle %d on invalid segment %d", i, vehicle.CurrentSegmentID))
		}

		if _, exists := g.Nodes[vehicle.TargetNodeID]; !exists {
			warnings = append(warnings, fmt.Sprintf("Vehicle %d has invalid target %d", i, vehicle.TargetNodeID))
		}

		if vehicle.SegmentProgress < 0 || vehicle.SegmentProgress > 1 {
			warnings = append(warnings, fmt.Sprintf("Vehicle %d has invalid progress %.2f", i, vehicle.SegmentProgress))
		}
	}

	return warnings
}

func GetVisualizationMetrics(g *coremodels.Grid, vehicles []*coremodels.Vehicle) map[string]interface{} {
	metrics := make(map[string]interface{})

	metrics["total_nodes"] = len(g.Nodes)
	metrics["total_segments"] = len(g.Segments)
	metrics["total_vehicles"] = len(vehicles)

	statusCounts := make(map[string]int)
	totalSpeed := 0.0
	speedCount := 0

	for _, vehicle := range vehicles {
		statusCounts[vehicle.Status.String()]++
		if vehicle.Status == coremodels.StatusMoving && vehicle.CurrentSpeedKPH > 0 {
			totalSpeed += vehicle.CurrentSpeedKPH
			speedCount++
		}
	}

	metrics["status_counts"] = statusCounts
	if speedCount > 0 {
		metrics["average_speed"] = totalSpeed / float64(speedCount)
	} else {
		metrics["average_speed"] = 0.0
	}

	segmentUtilization := make(map[int64]int)
	for _, vehicle := range vehicles {
		segmentUtilization[vehicle.CurrentSegmentID]++
	}

	utilizationStats := make(map[string]int)
	for _, count := range segmentUtilization {
		if count == 0 {
			utilizationStats["empty"]++
		} else if count <= 2 {
			utilizationStats["light"]++
		} else if count <= 5 {
			utilizationStats["moderate"]++
		} else {
			utilizationStats["heavy"]++
		}
	}
	metrics["segment_utilization"] = utilizationStats

	return metrics
}

func drawVehicle(canvas *svg.SVG, vehicle *coremodels.Vehicle, grid *coremodels.Grid, offsetX, offsetY float64) {
	x, y, err := vehicle.GetCurrentPosition(grid)
	if err != nil {
		return
	}

	canvasX := int(x + offsetX)
	canvasY := int(y + offsetY)

	color := getVehicleColor(vehicle.Status)
	size := 8

	if vehicle.Status == coremodels.StatusWaitingForPermission {
		waitingStyle := "fill:none;stroke:#ffc107;stroke-width:2;stroke-dasharray:3,2"
		canvas.Circle(canvasX, canvasY, size+3, waitingStyle)
	}

	vehicleStyle := fmt.Sprintf("fill:%s;stroke:#343a40;stroke-width:1.5", color)
	canvas.Circle(canvasX, canvasY, size, vehicleStyle)

	if vehicle.Status == coremodels.StatusMoving {
		direction := getVehicleDirection(vehicle, grid)
		if direction != nil {
			arrowX := canvasX + int(direction.X*12)
			arrowY := canvasY + int(direction.Y*12)
			arrowStyle := "stroke:#343a40;stroke-width:2;marker-end:url(#arrowhead)"
			canvas.Line(canvasX, canvasY, arrowX, arrowY, arrowStyle)
		}
	}

	speedText := fmt.Sprintf("%.0f", vehicle.CurrentSpeedKPH)
	speedStyle := "font-family:Arial;font-size:8px;fill:#495057;text-anchor:middle"
	canvas.Text(canvasX, canvasY-12, speedText, speedStyle)

	idText := vehicle.ID.String()[:4]
	idStyle := "font-family:Arial;font-size:7px;fill:#6c757d;text-anchor:middle"
	canvas.Text(canvasX, canvasY+20, idText, idStyle)
}

func getVehicleColor(status coremodels.VehicleStatus) string {
	switch status {
	case coremodels.StatusMoving:
		return "#f00dd9ff"
	case coremodels.StatusWaitingForPermission:
		return "#ffc107"
	case coremodels.StatusReachedDestination:
		return "#ffffffff"
	case coremodels.StatusDeadEnd, coremodels.StatusError:
		return "#dc3545"
	default:
		return "#000000ff"
	}
}

func getCongestionColor(congestionFactor float64) string {
	if congestionFactor <= 1.2 {
		return "#28a745"
	} else if congestionFactor <= 2.0 {
		return "#ffc107"
	} else if congestionFactor <= 3.0 {
		return "#fd7e14"
	} else {
		return "#dc3545"
	}
}

type Direction struct {
	X, Y float64
}

func getVehicleDirection(vehicle *coremodels.Vehicle, grid *coremodels.Grid) *Direction {
	segment, exists := grid.Segments[vehicle.CurrentSegmentID]
	if !exists {
		return nil
	}

	startNode, startExists := grid.Nodes[segment.StartNode]
	endNode, endExists := grid.Nodes[segment.EndNode]

	if !startExists || !endExists {
		return nil
	}

	var dirX, dirY float64
	if vehicle.TravelDirection >= 0 {
		dirX = endNode.Pos_X - startNode.Pos_X
		dirY = endNode.Pos_Y - startNode.Pos_Y
	} else {
		dirX = startNode.Pos_X - endNode.Pos_X
		dirY = startNode.Pos_Y - endNode.Pos_Y
	}

	length := math.Sqrt(dirX*dirX + dirY*dirY)
	if length == 0 {
		return nil
	}

	return &Direction{
		X: dirX / length,
		Y: dirY / length,
	}
}

func CreateAnimatedSequence(grid *coremodels.Grid, vehicles []*coremodels.Vehicle, baseFilename string, frames int) error {
	router := coremodels.NewVehicleRouter()

	for frame := 0; frame < frames; frame++ {
		for _, vehicle := range vehicles {
			if vehicle.Status != coremodels.StatusMoving {
				continue
			}

			vehicle.UpdateProgress(1.0, grid, router)
			if vehicle.HasReachedTarget(grid) {
				vehicle.Status = coremodels.StatusReachedDestination
				continue
			}

			if vehicle.IsAtIntersection() && vehicle.CanMakeMovementRequest() {
				decision, err := router.GetNextSegment(vehicle, grid)
				if err != nil {
					continue
				}

				if decision == nil {
					fmt.Println("no further moves")
					vehicle.Status = coremodels.StatusDeadEnd
					continue
				}

				switch decision.Reason {
				case "optimal", "exploration":
					vehicle.PrepareMovementRequest(decision.ToSegmentID, decision.FromNodeID)
					vehicle.HandleMovementResponse(true, "accepted", 0, grid)
				case "dead_end":
					fmt.Println("no further moves")
					vehicle.Status = coremodels.StatusDeadEnd
				case "reached_destination":
					vehicle.Status = coremodels.StatusReachedDestination
				default:
					fmt.Println("no further moves")
					vehicle.Status = coremodels.StatusDeadEnd
				}
			}
		}

		filename := fmt.Sprintf("%s_%03d.svg", baseFilename, frame)
		err := PlotGridWithVehicles(grid, vehicles, filename)
		if err != nil {
			return err
		}
	}

	return nil
}

func PlotVehicleTrails(grid *coremodels.Grid, vehicles []*coremodels.Vehicle, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	minX, minY, maxX, maxY := math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64
	for _, n := range grid.Nodes {
		minX = math.Min(minX, n.Pos_X)
		minY = math.Min(minY, n.Pos_Y)
		maxX = math.Max(maxX, n.Pos_X)
		maxY = math.Max(maxY, n.Pos_Y)
	}

	padding := 60.0
	contentWidth := maxX - minX
	contentHeight := maxY - minY

	canvasWidth := int(contentWidth + padding*2)
	canvasHeight := int(contentHeight + padding*2)

	canvas := svg.New(f)
	canvas.Start(canvasWidth, canvasHeight)

	canvas.Rect(0, 0, canvasWidth, canvasHeight, "fill:#1e1e1e")

	offsetX := padding - minX
	offsetY := padding - minY

	for _, seg := range grid.Segments {
		n1, ok1 := grid.Nodes[seg.StartNode]
		n2, ok2 := grid.Nodes[seg.EndNode]
		if !ok1 || !ok2 {
			continue
		}

		x1 := int(n1.Pos_X + offsetX)
		y1 := int(n1.Pos_Y + offsetY)
		x2 := int(n2.Pos_X + offsetX)
		y2 := int(n2.Pos_Y + offsetY)

		canvas.Line(x1, y1, x2, y2, "stroke:#333;stroke-width:2;opacity:0.3")
	}

	for _, vehicle := range vehicles {
		for i := 0; i < len(vehicle.RecentPositions)-1; i++ {
			p1 := vehicle.RecentPositions[i]
			p2 := vehicle.RecentPositions[i+1]
			x1 := int(p1.X + offsetX)
			y1 := int(p1.Y + offsetY)
			x2 := int(p2.X + offsetX)
			y2 := int(p2.Y + offsetY)
			ageFactor := float64(i) / float64(len(vehicle.RecentPositions))
			opacity := 0.2 + 0.8*ageFactor
			style := fmt.Sprintf("stroke:#00ff88;stroke-width:2;opacity:%.2f", opacity)
			canvas.Line(x1, y1, x2, y2, style)
		}

		x, y, err := vehicle.GetCurrentPosition(grid)
		if err != nil {
			continue
		}

		canvasX := int(x + offsetX)
		canvasY := int(y + offsetY)

		canvas.Circle(canvasX, canvasY, 8, "fill:#00ff88;opacity:0.9")
		canvas.Circle(canvasX, canvasY, 5, "fill:#ffffff")

		if targetNode, exists := grid.Nodes[vehicle.TargetNodeID]; exists {
			targetX := int(targetNode.Pos_X + offsetX)
			targetY := int(targetNode.Pos_Y + offsetY)

			canvas.Line(canvasX, canvasY, targetX, targetY,
				"stroke:#ff6b6b;stroke-width:1;stroke-dasharray:3,3;opacity:0.6")
		}
	}

	canvas.End()
	return nil
}

func getHeatmapColor(intensity float64) string {
	if intensity <= 0 {
		return "#cccccc"
	}
	if intensity <= 0.25 {
		return "#00ff00"
	}
	if intensity <= 0.5 {
		return "#ffff00"
	}
	if intensity <= 0.75 {
		return "#ff8000"
	}
	return "#ff0000"
}
