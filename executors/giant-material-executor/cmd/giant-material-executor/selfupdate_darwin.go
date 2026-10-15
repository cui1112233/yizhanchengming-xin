//go:build darwin

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"qiantie/giant-material-executor/internal/agent"
	"qiantie/giant-material-executor/internal/update"
)

func startSelfUpdater(ctx context.Context, stop context.CancelFunc, snapshot func() agent.Snapshot, originMu *sync.RWMutex, origin *string) *update.SelfUpdater {
	executable, err := os.Executable()
	if err != nil {
		log.Printf("self update disabled: executable path unavailable")
		return nil
	}
	root := filepath.Dir(executable)
	publicKey, err := update.PublicKeyFromBase64(bakedUpdatePublicKey)
	if err != nil {
		log.Printf("self update disabled: signed update key unavailable")
		return nil
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		log.Printf("self update disabled: cache directory unavailable")
		return nil
	}
	stageRoot := filepath.Join(cacheRoot, "YizhanShengming", "GiantMaterialExecutor", "updates")
	updater := &update.SelfUpdater{
		Root: root, StageRoot: stageRoot, CurrentVersion: version, PublicKey: publicKey,
		Target: update.ReleaseTarget{Platform: "macos", Architecture: "universal"},
		Origin: func() string { originMu.RLock(); defer originMu.RUnlock(); return *origin },
		Idle: func() bool {
			state := snapshot().State
			return state == agent.StateReady || state == agent.StateFailed || state == agent.StateIdle
		},
		Apply: func(nextVersion string) error {
			helper := filepath.Join(root, "GiantMaterialExecutorUpdater")
			if _, err := os.Stat(helper); err != nil {
				return fmt.Errorf("update helper unavailable: %w", err)
			}
			command := exec.Command(helper, "-platform", "macos", "-parent-pid", fmt.Sprint(os.Getpid()), "-app-root", root, "-stage-dir", filepath.Join(stageRoot, nextVersion))
			if err := command.Start(); err != nil {
				return fmt.Errorf("start update helper: %w", err)
			}
			log.Printf("PID-scoped update helper started; stopping executor for %s", nextVersion)
			stop()
			return nil
		},
	}
	go updater.Run(ctx)
	return updater
}
