package encoding

import (
	"math"
)

func float32ToInt64(f float32) int64 {
	if f > float32(math.MaxInt64) {
		return math.MaxInt64
	} else if f < float32(math.MinInt64) {
		return math.MinInt64
	}
	bits := math.Float32bits(f)

	return int64(bits)
}

func float64ToInt64(f float64) int64 {
	if f > float64(math.MaxInt64) {
		return math.MaxInt64
	} else if f < float64(math.MinInt64) {
		return math.MinInt64
	}
	bits := math.Float64bits(f)

	return int64(bits)
}

func int64ToFloat32(i int64) float32 {
	bits := uint32(i)

	return math.Float32frombits(bits)
}

func int64ToFloat64(i int64) float64 {
	bits := uint64(i)

	return math.Float64frombits(bits)
}
