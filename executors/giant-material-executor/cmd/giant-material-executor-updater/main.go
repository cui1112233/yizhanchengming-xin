package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	appBundleName = "GiantMaterialExecutor.app"
	macMainBinary = "GiantMaterialExecutor"
)

type updateRequest struct {
	Platform  string
	ParentPID int
	AppRoot   string
	StageDir  string
}

func validateRequest(request updateRequest) error {
	if request.ParentPID <= 0 {
		return errors.New("parent PID is required")
	}
	if request.Platform != "windows" && request.Platform != "macos" {
		return errors.New("unsupported update platform")
	}
	if strings.TrimSpace(request.AppRoot) == "" || strings.TrimSpace(request.StageDir) == "" {
		return errors.New("app root and stage directory are required")
	}
	return nil
}

// macAppBundleFromExecutable derives the one application bundle an updater is
// allowed to replace. It rejects a loose binary so an arbitrary file cannot be
// selected as an update target.
func macAppBundleFromExecutable(executable string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(executable))
	needle := string(os.PathSeparator) + appBundleName + string(os.PathSeparator) + "Contents" + string(os.PathSeparator) + "MacOS" + string(os.PathSeparator)
	index := strings.Index(clean, needle)
	if index <= 0 || filepath.Base(clean) != macMainBinary {
		return "", errors.New("executable is not inside GiantMaterialExecutor.app")
	}
	return clean[:index+len(string(os.PathSeparator)+appBundleName)], nil
}

// validateMacStage checks the minimum install shape before any existing app is
// renamed. Archive extraction itself remains the responsibility of the main
// executor update staging step.
func validateMacStage(stage string) error {
	app := filepath.Join(filepath.Clean(stage), appBundleName)
	info, err := os.Stat(app)
	if err != nil || !info.IsDir() {
		return errors.New("staged GiantMaterialExecutor.app is missing")
	}
	binary := filepath.Join(app, "Contents", "MacOS", macMainBinary)
	info, err = os.Stat(binary)
	if err != nil || info.IsDir() || info.Mode()&0o111 == 0 {
		return errors.New("staged GiantMaterialExecutor.app main binary is missing or not executable")
	}
	return nil
}

func main() {
	platform := flag.String("platform", "", "target platform")
	parentPID := flag.Int("parent-pid", 0, "executor process id")
	appRoot := flag.String("app-root", "", "executor install root")
	stageDir := flag.String("stage-dir", "", "verified staged update directory")
	flag.Parse()
	request := updateRequest{Platform: strings.TrimSpace(*platform), ParentPID: *parentPID, AppRoot: *appRoot, StageDir: *stageDir}
	if err := validateRequest(request); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := applyUpdate(request); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
