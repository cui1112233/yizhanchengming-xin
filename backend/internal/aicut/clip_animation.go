package aicut

// ClipAnimation is optional, independent of filters, transitions and keyframes.
type ClipAnimation struct {
	Entrance ClipAnimationPhase `json:"entrance"`
	Exit     ClipAnimationPhase `json:"exit"`
	Loop     ClipAnimationPhase `json:"loop"`
}

type ClipAnimationPhase struct {
	Preset     string `json:"preset"`
	DurationMS int64  `json:"durationMs"`
}

func validClipAnimation(value *ClipAnimation) bool {
	if value == nil {
		return true
	}
	for index, phase := range []ClipAnimationPhase{value.Entrance, value.Exit, value.Loop} {
		if phase.DurationMS < 0 || phase.DurationMS > 10000 {
			return false
		}
		if phase.Preset == "" || phase.Preset == "none" {
			continue
		}
		if phase.DurationMS < 50 || phase.DurationMS > 10000 {
			return false
		}
		if index == 2 {
			if phase.Preset != "pulse" && phase.Preset != "float" {
				return false
			}
		} else if phase.Preset != "fade" && phase.Preset != "slide" && phase.Preset != "zoom" {
			return false
		}
	}
	return true
}

func HasClipAnimation(value *ClipAnimation) bool {
	if value == nil {
		return false
	}
	for _, phase := range []ClipAnimationPhase{value.Entrance, value.Exit, value.Loop} {
		if phase.Preset != "" && phase.Preset != "none" {
			return true
		}
	}
	return false
}

// ClipAnimationDurations proportionally fits entrance and exit into short clips.
func ClipAnimationDurations(value *ClipAnimation, durationMS int64) (float64, float64) {
	if value == nil {
		return 0, 0
	}
	in, out := float64(0), float64(0)
	if value.Entrance.Preset != "" && value.Entrance.Preset != "none" {
		in = float64(value.Entrance.DurationMS)
	}
	if value.Exit.Preset != "" && value.Exit.Preset != "none" {
		out = float64(value.Exit.DurationMS)
	}
	if total := in + out; total > float64(durationMS) && total > 0 {
		ratio := max(0, float64(durationMS)) / total
		in *= ratio
		out *= ratio
	}
	return in, out
}
