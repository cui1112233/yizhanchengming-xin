package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMacAppBundleFromExecutableRejectsBinaryOutsideBundle(t *testing.T) {
	if _, err := macAppBundleFromExecutable("/tmp/GiantMaterialExecutor"); err == nil {
		t.Fatal("macAppBundleFromExecutable() error = nil, want rejection outside .app")
	}
}

func TestMacAppBundleFromExecutableFindsEnclosingApplication(t *testing.T) {
	executable := "/Applications/GiantMaterialExecutor.app/Contents/MacOS/GiantMaterialExecutor"
	bundle, err := macAppBundleFromExecutable(executable)
	if err != nil {
		t.Fatalf("macAppBundleFromExecutable() error = %v", err)
	}
	if bundle != "/Applications/GiantMaterialExecutor.app" {
		t.Fatalf("bundle = %q, want /Applications/GiantMaterialExecutor.app", bundle)
	}
}

func TestValidateMacStageRejectsMissingMainBinary(t *testing.T) {
	stage := t.TempDir()
	app := filepath.Join(stage, "GiantMaterialExecutor.app")
	if err := os.MkdirAll(filepath.Join(app, "Contents", "MacOS"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := validateMacStage(stage); err == nil {
		t.Fatal("validateMacStage() error = nil, want missing executable rejection")
	}
}

func TestValidateRequestRejectsUnsafeParentPID(t *testing.T) {
	err := validateRequest(updateRequest{Platform: "windows", ParentPID: 0, AppRoot: t.TempDir(), StageDir: t.TempDir()})
	if err == nil {
		t.Fatal("validateRequest() error = nil, want parent PID rejection")
	}
}
