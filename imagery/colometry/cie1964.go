package colometry

import (
	"fmt"
	"math"
)

type Vec3 [3]float64

type Matrix3x3 [3][3]float64

// Multiply multiplies a Vec3 by a Matrix3x3
func (v Vec3) Multiply(m Matrix3x3) Vec3 {
	return Vec3{
		v[0]*m[0][0] + v[1]*m[0][1] + v[2]*m[0][2],
		v[0]*m[1][0] + v[1]*m[1][1] + v[2]*m[1][2],
		v[0]*m[2][0] + v[1]*m[2][1] + v[2]*m[2][2],
	}
}

// MatchingFunction is a struct that holds the values of the CIE 1964 10 degree matching functions
type MatchingFunction struct {
	Wavelength int
	X          float64
	Y          float64
	Z          float64
}

var sRGBMatrix = Matrix3x3{
	{3.2404542, -1.5371385, -0.4985314},
	{-0.9692660, 1.8760108, 0.0415560},
	{0.0556434, -0.2040259, 1.0572252},
}

// NoResponse A number that isn't at the tables to indicate no response
const NoResponse = -77

// Map to hold the MatchingFunction values
var matchingFunctions = map[int]MatchingFunction{
	470: {
		Wavelength: 470,
		X:          0.195618,
		Y:          0.185190,
		Z:          1.31756,
	},
	510: {
		Wavelength: 510,
		X:          0.037465,
		Y:          0.606741,
		Z:          0.112044,
	},
	640: {
		Wavelength: 640,
		X:          0.431567,
		Y:          0.179828,
		Z:          0.0, // TODO: is it 0 or a different no response?
	},
}

// getMatchingFunction retrieves a MatchingFunction by wavelength
func getMatchingFunction(wavelength int) (MatchingFunction, error) {
	if mf, found := matchingFunctions[wavelength]; found {
		return mf, nil
	}
	return MatchingFunction{}, fmt.Errorf("matching function for wavelength %d not found", wavelength)
}

func ToRGB(wavelength int) (r, g, b float64, err error) {
	mf, err := getMatchingFunction(wavelength)
	if err != nil {
		return 0, 0, 0, err
	}
	coord := Vec3{mf.X / mf.Y, 1.0, mf.Z / mf.Y}
	rgb := coord.Multiply(sRGBMatrix)
	rgb[0] = GammaCorrectsRGB(max(0, rgb[0]))
	rgb[1] = GammaCorrectsRGB(max(0, rgb[1]))
	rgb[2] = GammaCorrectsRGB(max(0, rgb[2]))
	return rgb[0], rgb[1], rgb[2], nil
}

// GammaCorrectsRGB From https://stackoverflow.com/a/39446403
func GammaCorrectsRGB(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}

	a := 0.055
	return (1+a)*math.Pow(c, 1/2.4) - a
}
