//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func applyUpdate(request updateRequest) error {
	if request.Platform != "windows" {
		return fmt.Errorf("Windows helper cannot apply %s update", request.Platform)
	}
	if err := waitForParent(request.ParentPID); err != nil {
		return err
	}
	if err := copyStage(request.StageDir, request.AppRoot); err != nil {
		return err
	}
	return exec.Command(filepath.Join(request.AppRoot, "GiantMaterialExecutor.exe")).Start()
}

func waitForParent(pid int) error {
	needle := fmt.Sprintf("%d", pid)
	for attempts := 0; attempts < 60; attempts++ {
		output, err := exec.Command("tasklist", "/FI", "PID eq "+needle, "/NH").CombinedOutput()
		if err != nil || !strings.Contains(string(output), needle) {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("executor process %d did not exit", pid)
}

func copyStage(stage, root string) error {
	return filepath.Walk(stage, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(stage, path)
		if err != nil || rel == "." {
			return err
		}
		if filepath.Base(rel) == "GiantMaterialExecutorUpdater.exe" {
			return nil
		}
		destination := filepath.Join(root, rel)
		if info.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		return os.WriteFile(destination, data, info.Mode())
	})
}
