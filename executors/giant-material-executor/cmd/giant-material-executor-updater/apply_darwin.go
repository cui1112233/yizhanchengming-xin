//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

func applyUpdate(request updateRequest) error {
	if request.Platform != "macos" {
		return fmt.Errorf("macOS helper cannot apply %s update", request.Platform)
	}
	if err := waitForParent(request.ParentPID); err != nil {
		return err
	}
	if err := validateMacStage(request.StageDir); err != nil {
		return err
	}
	target, err := macAppBundleFromExecutable(filepath.Join(request.AppRoot, macMainBinary))
	if err != nil {
		return err
	}
	backup := target + ".previous"
	if _, err := os.Stat(backup); err == nil {
		return fmt.Errorf("previous update backup exists: %s", backup)
	}
	if err := os.Rename(target, backup); err != nil {
		return fmt.Errorf("rename current app: %w", err)
	}
	staged := filepath.Join(request.StageDir, appBundleName)
	if err := os.Rename(staged, target); err != nil {
		_ = os.Rename(backup, target)
		return fmt.Errorf("activate staged app: %w", err)
	}
	if err := exec.Command("/usr/bin/open", target).Start(); err != nil {
		_ = os.Rename(target, staged)
		_ = os.Rename(backup, target)
		return fmt.Errorf("restart updated app: %w", err)
	}
	_ = os.RemoveAll(backup)
	return nil
}

func waitForParent(pid int) error {
	for attempts := 0; attempts < 60; attempts++ {
		err := syscall.Kill(pid, 0)
		if err == syscall.ESRCH {
			return nil
		}
		if err != nil && err != syscall.EPERM {
			return err
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("executor process %d did not exit", pid)
}
