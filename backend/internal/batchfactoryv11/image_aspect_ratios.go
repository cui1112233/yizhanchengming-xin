package batchfactoryv11

import "strings"

// SupportedImageAspectRatios is the single source of truth for still-image
// aspect ratios across the batch factory, the private comic workspace, the
// script workspace and the agent canvas.
//
// Frontend mirror: frontend/src/shared/imageAspectRatios.js must list exactly
// the same values. The Doubao relay whitelist (executors/doubao-relay/
// src-tauri/src/tasks/validate.rs IMAGE_RATIOS) is the upstream evidence that
// 2:3, 3:4, 4:3 and 3:2 are native options on the live Doubao image page.
var SupportedImageAspectRatios = []string{"9:16", "16:9", "21:9", "1:1", "3:4", "4:3", "2:3", "3:2"}

// IsSupportedImageAspectRatio reports whether ratio belongs to the maintained
// image ratio set. An empty value is treated as "not set" and is not supported
// here; callers apply their own default before validating.
func IsSupportedImageAspectRatio(ratio string) bool {
	ratio = strings.TrimSpace(ratio)
	for _, candidate := range SupportedImageAspectRatios {
		if candidate == ratio {
			return true
		}
	}
	return false
}

// IsPortraitImageAspectRatio reports whether ratio is taller than wide.
func IsPortraitImageAspectRatio(ratio string) bool {
	switch strings.TrimSpace(ratio) {
	case "9:16", "3:4", "2:3":
		return true
	default:
		return false
	}
}

// ImageAspectRatioOrientation returns the prompt-safe composition phrase used
// when compiling comic prompts. Landscape ratios must never inherit the
// portrait-only "竖幅" wording.
func ImageAspectRatioOrientation(ratio string) string {
	switch strings.TrimSpace(ratio) {
	case "9:16", "3:4", "2:3":
		return "竖幅构图"
	case "16:9", "21:9", "4:3", "3:2":
		return "横幅构图"
	case "1:1":
		return "方形构图"
	default:
		return "竖幅构图"
	}
}
