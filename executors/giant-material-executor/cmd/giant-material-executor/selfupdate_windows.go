//go:build windows

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"qiantie/giant-material-executor/internal/agent"
	"qiantie/giant-material-executor/internal/update"
)

func startSelfUpdater(ctx context.Context, stop context.CancelFunc, snapshot func() agent.Snapshot, originMu *sync.RWMutex, origin *string) *update.SelfUpdater {
	root := ""
	if exe, err := os.Executable(); err == nil {
		root = filepath.Dir(exe)
	}
	if root == "" {
		log.Printf("self update disabled: executable path unavailable")
		return nil
	}
	publicKey, err := update.PublicKeyFromBase64(bakedUpdatePublicKey)
	if err != nil {
		log.Printf("self update disabled: signed update key unavailable")
		return nil
	}
	updater := &update.SelfUpdater{
		Root:           root,
		CurrentVersion: version,
		Origin: func() string {
			originMu.RLock()
			defer originMu.RUnlock()
			return *origin
		},
		Idle: func() bool {
			state := snapshot().State
			return state == agent.StateReady || state == agent.StateFailed || state == agent.StateIdle
		},
		PublicKey: publicKey,
		Target:    update.ReleaseTarget{Platform: "windows", Architecture: "amd64"},
		Apply: func(nextVersion string) error {
			helper := filepath.Join(root, "GiantMaterialExecutorUpdater.exe")
			if _, err := os.Stat(helper); err != nil {
				return fmt.Errorf("update helper unavailable: %w", err)
			}
			command := exec.Command(helper, "-platform", "windows", "-parent-pid", fmt.Sprint(os.Getpid()), "-app-root", root, "-stage-dir", filepath.Join(root, ".updates", nextVersion))
			command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
			if err := command.Start(); err != nil {
				return fmt.Errorf("start update helper: %w", err)
			}
			log.Printf("PID-scoped update helper started; stopping executor for %s", nextVersion)
			stop()
			return nil
		},
	}
	go updater.Run(ctx)
	log.Printf("self update enabled root=%s current=%s", root, version)
	return updater
}
