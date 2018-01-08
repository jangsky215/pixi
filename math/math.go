package math

import "math"

const (
	//弧度因子
	RadianFactor = math.Pi / 180.0
)

func Cos(radian float32) float32 {
	return float32(math.Cos(float64(radian)))
}

func Sin(radian float32) float32 {
	return float32(math.Sin(float64(radian)))
}

func Tan(radian float32) float32 {
	return float32(math.Tan(float64(radian)))
}

func Atan(radian float32) float32 {
	return float32(math.Atan(float64(radian)))
}
