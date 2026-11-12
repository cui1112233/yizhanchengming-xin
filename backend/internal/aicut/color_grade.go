package aicut

import (
	"fmt"
	"math"
	"strings"
)

type ColorWheelAdjustment struct {
	Hue        float64 `json:"hue"`
	Saturation float64 `json:"saturation"`
	Luminance  float64 `json:"luminance"`
}

var colorGradeChannels = [3]string{"red", "green", "blue"}
var colorGradeWheels = [3]string{"shadows", "midtones", "highlights"}

func finiteGrade(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
func knownGradeKey(key string, known [3]string) bool {
	for _, candidate := range known {
		if key == candidate {
			return true
		}
	}
	return false
}
func validColorGrade(color ColorAdjustment) bool {
	for key, points := range color.RGBCurves {
		if !knownGradeKey(key, colorGradeChannels) || len(points) < 2 || len(points) > 16 {
			return false
		}
		for i, p := range points {
			if !finiteGrade(p.X) || !finiteGrade(p.Y) || p.X < 0 || p.X > 1 || p.Y < 0 || p.Y > 1 || (i > 0 && p.X <= points[i-1].X) {
				return false
			}
		}
		if points[0].X != 0 || points[len(points)-1].X != 1 {
			return false
		}
	}
	for key, w := range color.Wheels {
		if !knownGradeKey(key, colorGradeWheels) || !finiteGrade(w.Hue) || !finiteGrade(w.Saturation) || !finiteGrade(w.Luminance) || w.Hue < -180 || w.Hue > 180 || w.Saturation < 0 || w.Saturation > 100 || w.Luminance < -100 || w.Luminance > 100 {
			return false
		}
	}
	return true
}
func HasColorGradeAdjustments(color ColorAdjustment) bool {
	if HasHSLAdjustments(color.HSL) {
		return true
	}
	for _, key := range colorGradeChannels {
		for _, p := range color.RGBCurves[key] {
			if p.X != p.Y {
				return true
			}
		}
	}
	for _, key := range colorGradeWheels {
		w := color.Wheels[key]
		if w.Saturation != 0 || w.Luminance != 0 {
			return true
		}
	}
	return false
}
func sampleGradeCurve(points []ColorCurvePoint, value float64) float64 {
	if len(points) < 2 {
		return value
	}
	for i := 1; i < len(points); i++ {
		if value <= points[i].X {
			a, b := points[i-1], points[i]
			return a.Y + (b.Y-a.Y)*(value-a.X)/(b.X-a.X)
		}
	}
	return points[len(points)-1].Y
}
func applyColorGradeRGB(red, green, blue float64, color ColorAdjustment) [3]uint8 {
	channels := [3]float64{sampleGradeCurve(color.RGBCurves["red"], red/255), sampleGradeCurve(color.RGBCurves["green"], green/255), sampleGradeCurve(color.RGBCurves["blue"], blue/255)}
	y := channels[0]*.2126 + channels[1]*.7152 + channels[2]*.0722
	weights := [3]float64{(1 - y) * (1 - y), 2 * y * (1 - y), y * y}
	bias := [3]float64{}
	lift := 0.0
	for index, key := range colorGradeWheels {
		w := color.Wheels[key]
		hue := math.Mod(math.Mod(w.Hue, 360)+360, 360)
		x := 1 - math.Abs(math.Mod(hue/60, 2)-1)
		vector := [3]float64{}
		switch {
		case hue < 60:
			vector = [3]float64{1, x, 0}
		case hue < 120:
			vector = [3]float64{x, 1, 0}
		case hue < 180:
			vector = [3]float64{0, 1, x}
		case hue < 240:
			vector = [3]float64{0, x, 1}
		case hue < 300:
			vector = [3]float64{x, 0, 1}
		default:
			vector = [3]float64{1, 0, x}
		}
		vectorY := vector[0]*.2126 + vector[1]*.7152 + vector[2]*.0722
		for i := 0; i < 3; i++ {
			bias[i] += (vector[i] - vectorY) * w.Saturation / 100 * .25 * weights[index]
		}
		lift += w.Luminance / 100 * weights[index]
	}
	for i := 0; i < 3; i++ {
		biased := math.Max(0, math.Min(1, channels[i]+bias[i]))
		channels[i] = math.Max(0, math.Min(1, hslShift(biased, lift))) * 255
	}
	return applyHSLRGB(channels[0], channels[1], channels[2], color.HSL)
}
func ApplyColorGradeToRGBA(pixel [4]uint8, color ColorAdjustment) [4]uint8 {
	if pixel[3] == 0 || !HasColorGradeAdjustments(color) {
		return pixel
	}
	rgb := applyColorGradeRGB(float64(pixel[0]), float64(pixel[1]), float64(pixel[2]), color)
	return [4]uint8{rgb[0], rgb[1], rgb[2], pixel[3]}
}

// Global/all curve is deliberately excluded; it already precedes this lattice.
func BuildColorGradeCube(color ColorAdjustment) string {
	var out strings.Builder
	out.Grow(1400000)
	out.WriteString("LUT_3D_SIZE 33\nDOMAIN_MIN 0 0 0\nDOMAIN_MAX 1 1 1\n")
	for b := 0; b < 33; b++ {
		for g := 0; g < 33; g++ {
			for r := 0; r < 33; r++ {
				rgb := applyColorGradeRGB(float64(r)*255/32, float64(g)*255/32, float64(b)*255/32, color)
				fmt.Fprintf(&out, "%.9f %.9f %.9f\n", float64(rgb[0])/255, float64(rgb[1])/255, float64(rgb[2])/255)
			}
		}
	}
	return out.String()
}
