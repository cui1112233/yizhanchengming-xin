package aicut

import (
	"fmt"
	"math"
	"strings"
)

// HSLBandAdjustment matches frontend/aicut/hslColor.js. Slider bounds are [-100,100].
type HSLBandAdjustment struct {
	Hue        float64 `json:"hue"`
	Saturation float64 `json:"saturation"`
	Luminance  float64 `json:"luminance"`
}

var hslBandKeys = [...]string{"red", "orange", "yellow", "green", "cyan", "blue", "purple", "magenta"}
var hslBandCenters = [...]float64{0, 30, 60, 120, 180, 240, 270, 300}

func validHSLAdjustments(settings map[string]HSLBandAdjustment) bool {
	for key, value := range settings {
		known := false
		for _, candidate := range hslBandKeys {
			if key == candidate {
				known = true
				break
			}
		}
		if !known {
			return false
		}
		for _, channel := range []float64{value.Hue, value.Saturation, value.Luminance} {
			if math.IsNaN(channel) || math.IsInf(channel, 0) || channel < -100 || channel > 100 {
				return false
			}
		}
	}
	return true
}

func hslClamp(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Max(-100, math.Min(100, value))
}

func HasHSLAdjustments(settings map[string]HSLBandAdjustment) bool {
	for _, key := range hslBandKeys {
		v := settings[key]
		if hslClamp(v.Hue) != 0 || hslClamp(v.Saturation) != 0 || hslClamp(v.Luminance) != 0 {
			return true
		}
	}
	return false
}

func hslShift(channel, amount float64) float64 {
	if amount >= 0 {
		return channel + (1-channel)*amount
	}
	return channel + channel*amount
}

func hslByte(value float64) uint8 { return uint8(math.Max(0, math.Min(255, math.Floor(value+0.5)))) }

// Original hue determines raised cosine membership in adjacent bands. Hue +/-100
// means +/-30 degrees; saturation/lightness move toward endpoints. Gray has no band.
func applyHSLRGB(red, green, blue float64, settings map[string]HSLBandAdjustment) [3]uint8 {
	original := [3]uint8{hslByte(red), hslByte(green), hslByte(blue)}
	r, g, b := red/255, green/255, blue/255
	maximum, minimum := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	delta := maximum - minimum
	if delta == 0 {
		return original
	}
	lightness := (maximum + minimum) / 2
	saturation := delta / (1 - math.Abs(2*lightness-1))
	hue := 0.0
	switch maximum {
	case r:
		hue = 60 * math.Mod((g-b)/delta, 6)
	case g:
		hue = 60 * ((b-r)/delta + 2)
	default:
		hue = 60 * ((r-g)/delta + 4)
	}
	hue = math.Mod(hue+360, 360)
	left := 7
	for i := 0; i < 7; i++ {
		if hue < hslBandCenters[i+1] {
			left = i
			break
		}
	}
	right := (left + 1) % 8
	end := 360.0
	if right != 0 {
		end = hslBandCenters[right]
	}
	weight := (1 - math.Cos(math.Pi*(hue-hslBandCenters[left])/(end-hslBandCenters[left]))) / 2
	a, z := settings[hslBandKeys[left]], settings[hslBandKeys[right]]
	h := hslClamp(a.Hue)*(1-weight) + hslClamp(z.Hue)*weight
	s := hslClamp(a.Saturation)*(1-weight) + hslClamp(z.Saturation)*weight
	l := hslClamp(a.Luminance)*(1-weight) + hslClamp(z.Luminance)*weight
	if h == 0 && s == 0 && l == 0 {
		return original
	}
	nextHue := math.Mod(hue+h*.3+360, 360)
	nextSaturation, nextLightness := hslShift(saturation, s/100), hslShift(lightness, l/100)
	chroma := (1 - math.Abs(2*nextLightness-1)) * nextSaturation
	x := chroma * (1 - math.Abs(math.Mod(nextHue/60, 2)-1))
	m := nextLightness - chroma/2
	rr, gg, bb := 0.0, 0.0, 0.0
	switch int(math.Floor(nextHue / 60)) {
	case 0:
		rr, gg = chroma, x
	case 1:
		rr, gg = x, chroma
	case 2:
		gg, bb = chroma, x
	case 3:
		gg, bb = x, chroma
	case 4:
		rr, bb = x, chroma
	case 5:
		rr, bb = chroma, x
	}
	return [3]uint8{hslByte((rr + m) * 255), hslByte((gg + m) * 255), hslByte((bb + m) * 255)}
}

func ApplyHSLToRGBA(pixel [4]uint8, settings map[string]HSLBandAdjustment) [4]uint8 {
	if pixel[3] == 0 || !HasHSLAdjustments(settings) {
		return pixel
	}
	rgb := applyHSLRGB(float64(pixel[0]), float64(pixel[1]), float64(pixel[2]), settings)
	return [4]uint8{rgb[0], rgb[1], rgb[2], pixel[3]}
}

// BuildHSLCube uses the same quantized 33^3 lattice as the browser atlas. FFmpeg
// trilinear interpolation matches two bilinear blue-slice samples in WebGL.
func BuildHSLCube(settings map[string]HSLBandAdjustment) string {
	var out strings.Builder
	out.Grow(1400000)
	out.WriteString("LUT_3D_SIZE 33\nDOMAIN_MIN 0 0 0\nDOMAIN_MAX 1 1 1\n")
	for b := 0; b < 33; b++ {
		for g := 0; g < 33; g++ {
			for r := 0; r < 33; r++ {
				rgb := applyHSLRGB(float64(r)*255/32, float64(g)*255/32, float64(b)*255/32, settings)
				fmt.Fprintf(&out, "%.9f %.9f %.9f\n", float64(rgb[0])/255, float64(rgb[1])/255, float64(rgb[2])/255)
			}
		}
	}
	return out.String()
}
