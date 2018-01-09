package math

import "math"

const (
	//弧度因子
	RadianFactor = math.Pi / 180.0
)

func _sin(radian float32) float32 {
	const n = sinTableLen
	i := uint32(radian * (n / math.Pi))
	x := i & n
	index := i & (n - 1)
	if x != 0 {
		return -sinTable[index]
	}

	return sinTable[index]
}

func Cos(radian float32) float32 {
	return _sin(radian + (math.Pi * 0.5))
}

func Sin(radian float32) float32 {
	return _sin(radian)
}

func Tan(radian float32) float32 {
	return float32(math.Tan(float64(radian)))
}

func Atan(radian float32) float32 {
	return float32(math.Atan(float64(radian)))
}

func Atan2(y, x float32) float32 {
	return float32(math.Atan2(float64(y), float64(x)))
}
