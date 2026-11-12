package httpapi

import (
	"fmt"
	"math"
	"qiantie/backend/internal/aicut"
	"strings"
)

func aicutNeedsColorLUT(c aicut.ColorAdjustment) bool {
	if c.Exposure != 0 || c.Temperature != 0 || c.Tint != 0 || c.Shadows != 0 || c.Highlights != 0 || (c.Gamma != 0 && c.Gamma != 1) || aicut.HasColorGradeAdjustments(c) {
		return true
	}
	for _, p := range c.CurvePoints {
		if p.X != p.Y {
			return true
		}
	}
	for i, v := range c.Curve {
		if v != float64(i)/4 {
			return true
		}
	}
	return false
}
func aicutSmoothstep(a, b, x float64) float64 {
	v := max(0, min(1, (x-a)/(b-a)))
	return v * v * (3 - 2*v)
}
func aicutBasicColor(rgb [3]float64, c aicut.ColorAdjustment) [3]float64 {
	y := rgb[0]*.2126 + rgb[1]*.7152 + rgb[2]*.0722
	lift := c.Shadows*.18*(1-aicutSmoothstep(0, .55, y)) + c.Highlights*.15*aicutSmoothstep(.45, 1, y)
	gamma := c.Gamma
	if gamma == 0 {
		gamma = 1
	}
	offsets := [3]float64{c.Temperature * .10, c.Tint * .08, -c.Temperature * .10}
	points := c.CurvePoints
	if len(points) < 2 && len(c.Curve) == 5 {
		for i, v := range c.Curve {
			points = append(points, aicut.ColorCurvePoint{X: float64(i) / 4, Y: v})
		}
	}
	for i, v := range rgb {
		v = math.Pow(max(0, v*math.Pow(2, c.Exposure)+offsets[i]+lift), 1/max(.5, gamma))
		if len(points) >= 2 {
			value := max(0, min(1, v))
			v = points[len(points)-1].Y
			for j := 1; j < len(points); j++ {
				if value <= points[j].X {
					a, b := points[j-1], points[j]
					v = a.Y + (b.Y-a.Y)*(value-a.X)/(b.X-a.X)
					break
				}
			}
		}
		rgb[i] = max(0, min(1, v))
	}
	return rgb
}
func aicutCombinedColorCube(c aicut.ColorAdjustment) string {
	var out strings.Builder
	out.WriteString("LUT_3D_SIZE 33\nDOMAIN_MIN 0 0 0\nDOMAIN_MAX 1 1 1\n")
	for b := 0; b < 33; b++ {
		for g := 0; g < 33; g++ {
			for r := 0; r < 33; r++ {
				rgb := aicutBasicColor([3]float64{float64(r) / 32, float64(g) / 32, float64(b) / 32}, c)
				if aicut.HasColorGradeAdjustments(c) {
					p := aicut.ApplyColorGradeToRGBA([4]uint8{uint8(math.Round(rgb[0] * 255)), uint8(math.Round(rgb[1] * 255)), uint8(math.Round(rgb[2] * 255)), 255}, c)
					rgb = [3]float64{float64(p[0]) / 255, float64(p[1]) / 255, float64(p[2]) / 255}
				}
				fmt.Fprintf(&out, "%.9f %.9f %.9f\n", rgb[0], rgb[1], rgb[2])
			}
		}
	}
	return out.String()
}
func aicutGradientASSTag(index, total int) string {
	stops := [3][3]float64{{255, 214, 107}, {183, 139, 255}, {99, 220, 255}}
	p := float64(index) / float64(max(1, total-1)) * 2
	segment := min(1, int(p))
	fraction := p - float64(segment)
	a, b := stops[segment], stops[segment+1]
	rgb := [3]int{}
	for j := range rgb {
		rgb[j] = int(math.Round(a[j] + (b[j]-a[j])*fraction))
	}
	return fmt.Sprintf("{\\1c&H%02X%02X%02X&}", rgb[2], rgb[1], rgb[0])
}
func aicutGradientASSText(text string) string {
	runes := []rune(text)
	var out strings.Builder
	for i, ch := range runes {
		out.WriteString(aicutGradientASSTag(i, len(runes)))
		out.WriteString(escapeAICutASSText(string(ch)))
	}
	return out.String()
}
