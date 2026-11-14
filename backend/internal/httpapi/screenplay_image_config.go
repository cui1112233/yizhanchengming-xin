package httpapi

import (
	"fmt"
	"qiantie/backend/internal/screenplay"
	"strconv"
	"strings"
)

// Image dimensions accompany Resolution because standard image channels serialize size only.
func screenplayImageDimensions(p screenplay.Project, aspect string) (string, string, error) {
	c := screenplay.EffectiveCreativeConfig(p)
	if aspect == "" {
		aspect = c.AspectRatio
	}
	if !screenplay.ValidCreativeAspectRatio(aspect) {
		return "", "", fmt.Errorf("%w: 不支持的图片比例%s", screenplay.ErrInvalidArgument, aspect)
	}
	longest := 2048
	switch c.ImageResolution {
	case "1K":
		longest = 1024
	case "4K":
		longest = 4096
	case "2K", "":
	default:
		return "", "", fmt.Errorf("%w: 不支持的图片分辨率%s", screenplay.ErrInvalidArgument, c.ImageResolution)
	}
	parts := strings.Split(aspect, ":")
	a, _ := strconv.Atoi(parts[0])
	b, _ := strconv.Atoi(parts[1])
	div := a
	if b > a {
		div = b
	}
	scale := longest / div
	return fmt.Sprintf("%dx%d", a*scale, b*scale), c.ImageResolution, nil
}

func screenplayCorePrefix(p screenplay.Project, text string) string {
	core := strings.TrimSpace(screenplay.EffectiveCreativeConfig(p).ImageCoreInstruction)
	if core == "" {
		return text
	}
	return core + "\n\n" + text
}
