package httpapi

import (
	"fmt"
	"qiantie/backend/internal/aicut"
)

// This runs after the existing canvas transform, in output time. Inverse sampling
// keeps the canvas fixed in size and makes pixels outside a moved clip transparent.
func buildAICutClipAnimationFilters(clip aicut.Clip, timelineComposition bool) []string {
	value := clip.Animation
	if !aicut.HasClipAnimation(value) {
		return nil
	}
	time := "T"
	if timelineComposition && clip.StartMS > 0 {
		time = fmt.Sprintf("(T-%.6f)", float64(clip.StartMS)/1000)
	}
	duration := float64(clip.DurationMS) / 1000
	in, out := aicut.ClipAnimationDurations(value, clip.DurationMS)
	x, y, scale, opacity := "0", "0", "1", "1"
	for index, phase := range []aicut.ClipAnimationPhase{value.Entrance, value.Exit} {
		amount := "0"
		if index == 0 && in > 0 {
			amount = fmt.Sprintf("(1-clip(%s/%.6f,0,1))", time, in/1000)
		}
		if index == 1 && out > 0 {
			amount = fmt.Sprintf("clip((%s-%.6f+%.6f)/%.6f,0,1)", time, duration, out/1000, out/1000)
		}
		switch phase.Preset {
		case "fade":
			opacity = fmt.Sprintf("(%s)*(1-(%s))", opacity, amount)
		case "slide":
			sign := "-"
			if index == 1 {
				sign = "+"
			}
			x = fmt.Sprintf("(%s)%s(%s)", x, sign, amount)
		case "zoom":
			scale = fmt.Sprintf("(%s)*(1-0.2*(%s))", scale, amount)
		}
	}
	period := float64(max(50, value.Loop.DurationMS)) / 1000
	if value.Loop.Preset == "pulse" {
		scale = fmt.Sprintf("(%s)*(1+0.025*(1-cos(%s/%.6f*2*PI)))", scale, time, period)
	}
	if value.Loop.Preset == "float" {
		y = fmt.Sprintf("0.015*sin(%s/%.6f*2*PI)", time, period)
	}
	sx := fmt.Sprintf("((X-W/2-(%s)*W)/(%s)+W/2)", x, scale)
	sy := fmt.Sprintf("((Y-H/2-(%s)*H)/(%s)+H/2)", y, scale)
	inside := fmt.Sprintf("between(%s,0,W-1)*between(%s,0,H-1)", sx, sy)
	return []string{"format=rgba", fmt.Sprintf("geq=r='r(%s,%s)':g='g(%s,%s)':b='b(%s,%s)':a='if(%s,alpha(%s,%s)*(%s),0)'", sx, sy, sx, sy, sx, sy, inside, sx, sy, opacity)}
}
