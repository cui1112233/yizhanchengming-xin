package batchfactoryv11

// EffectiveVideoAspectRatio keeps the video contract separate from reference
// image rendering. Legacy batches only stored aspectRatio, so a historic 1:1
// image setting deliberately falls back to vertical video instead of sending
// an unsupported square-video request to a provider.
func EffectiveVideoAspectRatio(values SettingsPatch) string {
	if aspect := rawString(values, "videoAspectRatio", ""); aspect == "9:16" || aspect == "16:9" {
		return aspect
	}
	if legacy := rawString(values, "aspectRatio", ""); legacy == "16:9" {
		return legacy
	}
	return "9:16"
}

// EffectiveImageAspectRatio is shared by all still-image work, including
// reference images, asset images, and complete comic pages. It intentionally
// permits square images without changing the video contract.
func EffectiveImageAspectRatio(values SettingsPatch) string {
	if aspect := rawString(values, "imageAspectRatio", ""); aspect == "9:16" || aspect == "16:9" || aspect == "1:1" {
		return aspect
	}
	if legacy := rawString(values, "aspectRatio", ""); legacy == "9:16" || legacy == "16:9" || legacy == "1:1" {
		return legacy
	}
	return "9:16"
}

// EffectiveVideoResolution accepts the new explicit setting while preserving
// existing provider-specific legacy resolution values during migration.
func EffectiveVideoResolution(values SettingsPatch) string {
	if resolution := rawString(values, "videoResolution", ""); resolution != "" {
		return resolution
	}
	if resolution := rawString(values, "resolution", ""); resolution != "" {
		return resolution
	}
	return "720p"
}
