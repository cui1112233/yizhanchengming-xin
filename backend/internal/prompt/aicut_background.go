package prompt

// AICutBackgroundRemoval is kept on the server so the browser never owns or
// drifts from the production image-editing instruction.
func AICutBackgroundRemoval() string {
	return `Remove only the background from the supplied image. Preserve the subject exactly: identity, face, hair, clothing, body proportions, pose, product details, colors, lighting, shadows on the subject, and image dimensions. Keep fine edge detail such as hair and translucent fabric. Do not add, redraw, crop, beautify, or move anything. Return one PNG with a genuinely transparent background (alpha channel), not a checkerboard, white, black, or colored background.`
}
