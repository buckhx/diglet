package tile_system

import (
	"fmt"
	"math"
)

const (
	MinLat       float64 = -85.05112878
	MaxLat       float64 = 85.05112878
	MinLon       float64 = -180
	MaxLon       float64 = 180
	EarthRadiusM uint    = 6378137
)

// if val is outside of min-max range, clip it to min or max
func clip(val, min, max float64) (float64, error) {
	if min > max {
		return 0, fmt.Errorf("clip min %s > max %s", min, max)
	}
	return math.Min(math.Max(val, min), max), nil
}
