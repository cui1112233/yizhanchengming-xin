package platform

import (
	"fmt"
	"path/filepath"
)

func ExecutorPlatform(goos string) string { return goos }

func WorkerCommand(goos, executableDir string) ([]string, error) {
	switch goos {
	case "darwin":
		return []string{filepath.Join(executableDir, "GiantMaterialOCRWorker")}, nil
	case "windows":
		return []string{"python", filepath.Join(executableDir, "worker", "ocr_worker.py")}, nil
	default:
		return nil, fmt.Errorf("unsupported giant material executor platform: %s", goos)
	}
}
