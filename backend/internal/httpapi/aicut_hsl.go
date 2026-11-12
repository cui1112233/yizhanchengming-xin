package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"qiantie/backend/internal/aicut"
)

// Files live only inside the existing render temp directory, cleaned by the
// existing render cleanup. Content hash filenames never contain user input.
func prepareAICutHSLLUTs(dir string, clips []aicutRenderClip) error {
	cache := map[string]string{}
	for index := range clips {
		item := &clips[index]
		item.hslLUTPath = ""
		if item.clip.VisualProcessing == nil || !aicut.HasColorGradeAdjustments(item.clip.VisualProcessing.Color) {
			continue
		}
		color := item.clip.VisualProcessing.Color
		settings := struct {
			HSL       map[string]aicut.HSLBandAdjustment
			RGBCurves map[string][]aicut.ColorCurvePoint
			Wheels    map[string]aicut.ColorWheelAdjustment
		}{color.HSL, color.RGBCurves, color.Wheels}
		data, err := json.Marshal(settings)
		if err != nil {
			return fmt.Errorf("HSL 调色参数无效")
		}
		digest := sha256.Sum256(data)
		signature := hex.EncodeToString(digest[:])
		if existing := cache[signature]; existing != "" {
			item.hslLUTPath = existing
			continue
		}
		target := filepath.Join(dir, "hsl-"+signature+".cube")
		if err := os.WriteFile(target, []byte(aicut.BuildColorGradeCube(color)), 0o600); err != nil {
			return fmt.Errorf("HSL 调色文件准备失败")
		}
		cache[signature] = target
		item.hslLUTPath = target
	}
	return nil
}
