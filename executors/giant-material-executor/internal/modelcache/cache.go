package modelcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	bundleName   = "runtime.bundle"
	manifestName = "manifest.json"
)

type DownloadProgress struct {
	Completed int64 `json:"completed"`
	Total     int64 `json:"total"`
	Percent   int   `json:"percent"`
}

type ModelInstall struct {
	Dir      string
	Version  string
	Manifest Manifest
}

type Cache struct {
	Root   string
	Client *http.Client
}

func (c Cache) Ensure(ctx context.Context, manifest Manifest, progress func(DownloadProgress)) (ModelInstall, error) {
	if err := manifest.Validate(); err != nil {
		return ModelInstall{}, err
	}
	if strings.TrimSpace(c.Root) == "" {
		return ModelInstall{}, fmt.Errorf("MODEL_MANIFEST_INVALID: cache root is required")
	}
	if err := os.MkdirAll(c.Root, 0o700); err != nil {
		return ModelInstall{}, fmt.Errorf("MODEL_DOWNLOAD_FAILED: create cache root: %w", err)
	}

	activeDir := filepath.Join(c.Root, manifest.Version)
	if activeInstall, ok := c.validActiveInstall(activeDir, manifest); ok {
		emitProgress(progress, DownloadProgress{Completed: manifest.Size, Total: manifest.Size, Percent: 100})
		return activeInstall, nil
	}

	partPath := filepath.Join(c.Root, manifest.Version+".part")
	if err := c.download(ctx, manifest, partPath, progress); err != nil {
		return ModelInstall{}, err
	}

	hash, err := hashFile(partPath)
	if err != nil {
		return ModelInstall{}, fmt.Errorf("MODEL_DOWNLOAD_FAILED: hash package: %w", err)
	}
	if !strings.EqualFold(hash, manifest.SHA256) {
		return ModelInstall{}, fmt.Errorf("MODEL_HASH_MISMATCH: expected %s, got %s", strings.ToLower(manifest.SHA256), hash)
	}

	install, err := c.activate(partPath, activeDir, manifest)
	if err != nil {
		return ModelInstall{}, err
	}
	emitProgress(progress, DownloadProgress{Completed: manifest.Size, Total: manifest.Size, Percent: 100})
	return install, nil
}

func (c Cache) download(ctx context.Context, manifest Manifest, partPath string, progress func(DownloadProgress)) error {
	start := int64(0)
	if info, err := os.Stat(partPath); err == nil {
		start = info.Size()
		if start > manifest.Size {
			if err := os.Truncate(partPath, 0); err != nil {
				return fmt.Errorf("MODEL_DOWNLOAD_FAILED: reset oversized partial: %w", err)
			}
			start = 0
		}
	}

	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := c.request(ctx, client, manifest.PackageURL, start)
	if err != nil {
		return fmt.Errorf("MODEL_DOWNLOAD_FAILED: %w", err)
	}
	appendMode := start > 0 && resp.StatusCode == http.StatusPartialContent
	if start > 0 && !appendMode {
		resp.Body.Close()
		start = 0
		resp, err = c.request(ctx, client, manifest.PackageURL, 0)
		if err != nil {
			return fmt.Errorf("MODEL_DOWNLOAD_FAILED: %w", err)
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("MODEL_DOWNLOAD_FAILED: server returned HTTP %d", resp.StatusCode)
	}

	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	file, err := os.OpenFile(partPath, flags, 0o600)
	if err != nil {
		return fmt.Errorf("MODEL_DOWNLOAD_FAILED: open partial: %w", err)
	}
	defer file.Close()

	completed := start
	buffer := make([]byte, 32*1024)
	for {
		read, readErr := resp.Body.Read(buffer)
		if read > 0 {
			completed += int64(read)
			if completed > manifest.Size {
				return fmt.Errorf("MODEL_DOWNLOAD_FAILED: package exceeds manifest size")
			}
			if _, err := file.Write(buffer[:read]); err != nil {
				return fmt.Errorf("MODEL_DOWNLOAD_FAILED: write partial: %w", err)
			}
			emitProgress(progress, DownloadProgress{Completed: completed, Total: manifest.Size, Percent: percent(completed, manifest.Size)})
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return fmt.Errorf("MODEL_DOWNLOAD_FAILED: read package: %w", readErr)
		}
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("MODEL_DOWNLOAD_FAILED: sync partial: %w", err)
	}
	if completed != manifest.Size {
		return fmt.Errorf("MODEL_DOWNLOAD_FAILED: expected %d bytes, got %d", manifest.Size, completed)
	}
	return nil
}

func (c Cache) request(ctx context.Context, client *http.Client, packageURL string, start int64) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, packageURL, nil)
	if err != nil {
		return nil, err
	}
	if start > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", start))
	}
	return client.Do(req)
}

func (c Cache) validActiveInstall(activeDir string, expected Manifest) (ModelInstall, bool) {
	data, err := os.ReadFile(filepath.Join(activeDir, manifestName))
	if err != nil {
		return ModelInstall{}, false
	}
	var actual Manifest
	if err := json.Unmarshal(data, &actual); err != nil {
		return ModelInstall{}, false
	}
	if actual.Version != expected.Version || actual.Size != expected.Size || !strings.EqualFold(actual.SHA256, expected.SHA256) || actual.RuntimeVersion != expected.RuntimeVersion {
		return ModelInstall{}, false
	}
	info, err := os.Stat(filepath.Join(activeDir, bundleName))
	if err != nil || info.Size() != expected.Size {
		return ModelInstall{}, false
	}
	actualHash, err := hashFile(filepath.Join(activeDir, bundleName))
	if err != nil || !strings.EqualFold(actualHash, expected.SHA256) {
		return ModelInstall{}, false
	}
	return ModelInstall{Dir: activeDir, Version: actual.Version, Manifest: actual}, true
}

func (c Cache) activate(partPath, activeDir string, manifest Manifest) (ModelInstall, error) {
	stagingDir, err := os.MkdirTemp(c.Root, "."+manifest.Version+".staging-")
	if err != nil {
		return ModelInstall{}, fmt.Errorf("MODEL_DOWNLOAD_FAILED: create staging directory: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(stagingDir)
		}
	}()
	if err := os.Rename(partPath, filepath.Join(stagingDir, bundleName)); err != nil {
		return ModelInstall{}, fmt.Errorf("MODEL_DOWNLOAD_FAILED: stage package: %w", err)
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return ModelInstall{}, fmt.Errorf("MODEL_MANIFEST_INVALID: encode manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(stagingDir, manifestName), append(manifestData, '\n'), 0o600); err != nil {
		return ModelInstall{}, fmt.Errorf("MODEL_DOWNLOAD_FAILED: write manifest: %w", err)
	}

	backupDir := ""
	if _, err := os.Stat(activeDir); err == nil {
		backupDir = activeDir + ".old-" + time.Now().UTC().Format("20060102T150405.000000000Z")
		if err := os.Rename(activeDir, backupDir); err != nil {
			return ModelInstall{}, fmt.Errorf("MODEL_DOWNLOAD_FAILED: stage old active version: %w", err)
		}
	}
	if err := os.Rename(stagingDir, activeDir); err != nil {
		if backupDir != "" {
			_ = os.Rename(backupDir, activeDir)
		}
		return ModelInstall{}, fmt.Errorf("MODEL_DOWNLOAD_FAILED: activate package: %w", err)
	}
	cleanup = false
	if backupDir != "" {
		_ = os.RemoveAll(backupDir)
	}
	return ModelInstall{Dir: activeDir, Version: manifest.Version, Manifest: manifest}, nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func percent(completed, total int64) int {
	if total <= 0 {
		return 0
	}
	value := int((completed * 100) / total)
	if value > 99 {
		return 99
	}
	return value
}

func emitProgress(progress func(DownloadProgress), update DownloadProgress) {
	if progress != nil {
		progress(update)
	}
}
