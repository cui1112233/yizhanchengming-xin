package update

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SelfUpdater owns local availability checks. It never replaces the running
// executable unless ApplyNow was called by the local executor UI.
type SelfUpdater struct {
	Root           string
	StageRoot      string
	CurrentVersion string
	Origin         func() string
	HTTPClient     *http.Client
	PublicKey      ed25519.PublicKey
	Target         ReleaseTarget
	Idle           func() bool
	Apply          func(nextVersion string) error

	once       sync.Once
	controller *Controller
}

const (
	checkInterval    = 30 * time.Minute
	firstCheckDelay  = time.Minute
	maxDownloadBytes = 128 << 20
)

func (u *SelfUpdater) manifestURL() string {
	return strings.TrimRight(u.Origin(), "/") + "/downloads/giant-material-executor/releases/" + u.Target.Platform + "/latest.json"
}

// Run checks for availability in the background. It never stages, restarts,
// or exits the executor.
func (u *SelfUpdater) Run(ctx context.Context) {
	timer := time.NewTimer(firstCheckDelay)
	defer timer.Stop()
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <-ticker.C:
		}
		if _, err := u.CheckNow(ctx); err != nil {
			log.Printf("self update check skipped: %v", err)
		}
	}
}

// CheckNow only checks for an available release. It does not download or
// restart. An empty string means the executor is already current.
func (u *SelfUpdater) CheckNow(ctx context.Context) (string, error) {
	status, err := u.getController().Check(ctx)
	return status.AvailableVersion, err
}

// ApplyNow stages the available release after an explicit local user action,
// then hands off to a platform-specific replacement helper.
func (u *SelfUpdater) ApplyNow(ctx context.Context) (string, error) {
	if u == nil || u.Apply == nil {
		return "", errors.New("self update is not configured")
	}
	status, err := u.getController().Apply(ctx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(status.AvailableVersion) == "" {
		return "", &CodedError{Code: "UPDATE_NOT_AVAILABLE"}
	}
	if err := u.Apply(status.AvailableVersion); err != nil {
		return "", err
	}
	return status.AvailableVersion, nil
}

func (u *SelfUpdater) Status() Status { return u.getController().Status() }

func (u *SelfUpdater) getController() *Controller {
	u.once.Do(func() {
		u.controller = NewController(ControllerConfig{
			CurrentVersion: u.CurrentVersion,
			Target:         u.Target,
			PublicKey:      u.PublicKey,
			Idle:           u.Idle,
			Fetch: func(ctx context.Context) (ReleaseManifest, error) {
				if u == nil || u.Origin == nil || strings.TrimSpace(u.Origin()) == "" {
					return ReleaseManifest{}, errors.New("self update is not configured")
				}
				return fetchManifest(ctx, u.httpClient(), u.manifestURL())
			},
			Stage: func(ctx context.Context, manifest ReleaseManifest) error {
				if strings.TrimSpace(u.Root) == "" {
					return errors.New("self update root is not configured")
				}
				return u.stage(ctx, u.httpClient(), manifest)
			},
		})
	})
	return u.controller
}

func (u *SelfUpdater) httpClient() *http.Client {
	if u.HTTPClient != nil {
		return u.HTTPClient
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func fetchManifest(ctx context.Context, client *http.Client, url string) (ReleaseManifest, error) {
	var manifest ReleaseManifest
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return manifest, err
	}
	response, err := client.Do(req)
	if err != nil {
		return manifest, errors.New("无法连接更新清单地址：" + err.Error())
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return manifest, fmt.Errorf("更新清单返回 HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, errors.New("更新清单不是合法的 JSON")
	}
	return manifest, nil
}

// stage downloads, verifies and extracts an archive without replacing the app.
func (u *SelfUpdater) stage(ctx context.Context, client *http.Client, manifest ReleaseManifest) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSpace(manifest.URL), nil)
	if err != nil {
		return err
	}
	response, err := client.Do(req)
	if err != nil {
		return errors.New("下载新版本失败：" + err.Error())
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("下载新版本返回 HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxDownloadBytes))
	if err != nil {
		return errors.New("下载新版本失败：" + err.Error())
	}
	sum := sha256.Sum256(data)
	if !strings.EqualFold(hex.EncodeToString(sum[:]), strings.TrimSpace(manifest.SHA256)) {
		return errors.New("新版本文件校验失败（SHA256 不匹配），已放弃更新")
	}
	stageRoot := strings.TrimSpace(u.StageRoot)
	if stageRoot == "" {
		stageRoot = filepath.Join(u.Root, ".updates")
	}
	stageDir := filepath.Join(stageRoot, manifest.Version)
	if err := os.RemoveAll(stageDir); err != nil {
		return err
	}
	if err := unzip(data, stageDir); err != nil {
		_ = os.RemoveAll(stageDir)
		return errors.New("解压新版本失败：" + err.Error())
	}
	return nil
}

func unzip(data []byte, dest string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	dest = filepath.Clean(dest)
	for _, entry := range reader.File {
		name := filepath.FromSlash(strings.ReplaceAll(entry.Name, "\\", "/"))
		target := filepath.Join(dest, name)
		if !strings.HasPrefix(target, dest+string(os.PathSeparator)) {
			return errors.New("update archive contains an unsafe path")
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		source, err := entry.Open()
		if err != nil {
			return err
		}
		content, readErr := io.ReadAll(io.LimitReader(source, maxDownloadBytes))
		_ = source.Close()
		if readErr != nil {
			return readErr
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}
