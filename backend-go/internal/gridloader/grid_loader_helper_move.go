package gridloader

import (
	"math/rand"

	"math"
)
func (gl *GridLoader) calculateSegmentLength(fromX, fromY, toX, toY int64) float64 {
    dx := float64(toX - fromX)
    dy := float64(toY - fromY)

    baseDistance := math.Sqrt(dx*dx + dy*dy)
    const kmPerGridUnit = 0.5
    const variationRange = 0.2

    variation := 1.0 + (rand.Float64()-0.5)*variationRange
    return baseDistance * kmPerGridUnit * variation
}
