package main

/* Taken from http://freespace.virgin.net/hugo.elias/models/m_perlin.htm */

import "math"

func IntNoise(x int) int {
	x = (x << 13) ^ x
	return (1.0 - ((x*(x*x*15731+789221)+1376312589)&0x7fffffff)/1073741824.0)
}

func Noise(x float64) float64 {
	return float64(IntNoise(int(x)))
}

func SmoothedNoise1(x float64) float64 {
	return Noise(x)/2 + Noise(x-1)/4 + Noise(x+1)/4
}

// Cosine
func Interpolate(a, b, x float64) float64 {
	ft := x * 3.1415927
	f := (1 - math.Cos(ft)) * .5

	return a*(1-f) + b*f
}

func InterpolatedNoise_1(x float64) float64 {
	integer_X := math.Floor(x)
	fractional_X := x - integer_X

	v1 := SmoothedNoise1(integer_X)
	v2 := SmoothedNoise1(integer_X + 1)

	return Interpolate(v1, v2, fractional_X)
}

func PerlinNoise_1D(x float64) float64 {
	total := 0.0
	p := 0.25 // persistence
	n := 3.0  // Number_Of_Octaves - 1

	for i := float64(0); i < n; i++ {
		frequency := math.Pow(2, i)
		amplitude := math.Pow(p, i)

		total = total + InterpolatedNoise_1(x*frequency)*amplitude
	}

	return total

}
