package localartifact

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	ErrInvalidMP4      = errors.New("invalid mp4 artifact")
	ErrInvalidImage    = errors.New("invalid image artifact")
	ErrTooLarge        = errors.New("artifact exceeds size limit")
	ErrInvalidID       = errors.New("invalid artifact id")
	ErrAlreadyExists   = errors.New("artifact already exists")
	ErrRemoteIntegrity = errors.New("restored artifact failed integrity verification")
	ErrDiskPressure    = errors.New("server disk pressure blocks new media writes")
	ErrRestoreTimeout  = errors.New("artifact restore timed out")
)

// 回源超时默认 2 分钟：本地缓存缺失时从 TOS 重新下载必须有一个上限，否则
// 一次慢回源会长期占住请求 goroutine。大文件场景可通过环境变量调高。
const (
	defaultRestoreTimeout = 2 * time.Minute
	maxRestoreTimeout     = 10 * time.Minute
)

var safeArtifactID = regexp.MustCompile(`^[A-Za-z0-9_-]{3,96}$`)

// diskPressureError 把底层“磁盘已满”IO 错误统一成 ErrDiskPressure。没有这层
// 归一，ENOSPC 会被当成普通瞬时错误反复重试，页面上只会一直显示“执行中”。
func diskPressureError(err error) error {
	if err == nil || !errors.Is(err, syscall.ENOSPC) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrDiskPressure, err)
}

type Artifact struct {
	ID         string
	MediaType  string
	ByteSize   int64
	SHA256     string
	StorageRef string
}

type Store struct {
	root            string
	maxBytes        int64
	mu              sync.RWMutex
	missingRestorer MissingRestorer
	writeGuard      func() error
	restoreTimeout  time.Duration
}

type MissingRestorer interface {
	RestoreStorageRef(context.Context, string) error
}

func NewStore(root string, maxBytes int64) *Store {
	if maxBytes <= 0 {
		maxBytes = 1 << 30
	}
	return &Store{root: filepath.Clean(root), maxBytes: maxBytes, restoreTimeout: defaultRestoreTimeout}
}

// Root returns the private directory that is mounted for local artifacts.
func (s *Store) Root() string {
	if s == nil {
		return ""
	}
	return s.root
}

func (s *Store) SetMissingRestorer(restorer MissingRestorer) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.missingRestorer = restorer
	s.mu.Unlock()
}

func (s *Store) SetWriteGuard(guard func() error) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.writeGuard = guard
	s.mu.Unlock()
}

// SetRestoreTimeout 设置本地缓存缺失时的回源上限。非正值回退默认值，
// 超过上限按上限处理，避免误配置成无限等待。
func (s *Store) SetRestoreTimeout(timeout time.Duration) {
	if s == nil {
		return
	}
	if timeout <= 0 {
		timeout = defaultRestoreTimeout
	}
	if timeout > maxRestoreTimeout {
		timeout = maxRestoreTimeout
	}
	s.mu.Lock()
	s.restoreTimeout = timeout
	s.mu.Unlock()
}

func (s *Store) restoreTimeoutValue() time.Duration {
	if s == nil {
		return defaultRestoreTimeout
	}
	s.mu.RLock()
	timeout := s.restoreTimeout
	s.mu.RUnlock()
	if timeout <= 0 {
		return defaultRestoreTimeout
	}
	return timeout
}

func (s *Store) checkWriteAllowed() error {
	if s == nil {
		return ErrInvalidID
	}
	s.mu.RLock()
	guard := s.writeGuard
	s.mu.RUnlock()
	if guard != nil {
		return guard()
	}
	return nil
}

func (s *Store) SaveMP4(id string, src io.Reader) (artifact Artifact, err error) {
	if s == nil || !safeArtifactID.MatchString(id) || src == nil {
		return Artifact{}, ErrInvalidID
	}
	if err := s.checkWriteAllowed(); err != nil {
		return Artifact{}, err
	}
	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return Artifact{}, diskPressureError(err)
	}
	finalName := id + ".mp4"
	finalPath := filepath.Join(s.root, finalName)
	if _, statErr := os.Stat(finalPath); statErr == nil {
		return Artifact{}, ErrAlreadyExists
	} else if !os.IsNotExist(statErr) {
		return Artifact{}, statErr
	}

	tmp, err := os.CreateTemp(s.root, ".upload-*.tmp")
	if err != nil {
		return Artifact{}, diskPressureError(err)
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	first := make([]byte, 24)
	nFirst, readErr := io.ReadFull(src, first)
	if readErr != nil {
		return Artifact{}, ErrInvalidMP4
	}
	first = first[:nFirst]
	if len(first) < 12 || string(first[4:8]) != "ftyp" {
		return Artifact{}, ErrInvalidMP4
	}
	if int64(len(first)) > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}

	hash := sha256.New()
	writer := io.MultiWriter(tmp, hash)
	if _, err := writer.Write(first); err != nil {
		return Artifact{}, diskPressureError(err)
	}
	remainingLimit := s.maxBytes - int64(len(first)) + 1
	copied, err := io.Copy(writer, io.LimitReader(src, remainingLimit))
	if err != nil {
		return Artifact{}, diskPressureError(err)
	}
	total := int64(len(first)) + copied
	if total > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}
	if err := tmp.Sync(); err != nil {
		return Artifact{}, diskPressureError(err)
	}
	if err := tmp.Close(); err != nil {
		return Artifact{}, err
	}
	if err := os.Rename(tmpName, finalPath); err != nil {
		return Artifact{}, fmt.Errorf("commit artifact: %w", diskPressureError(err))
	}
	committed = true
	return Artifact{
		ID: id, MediaType: "video/mp4", ByteSize: total,
		SHA256: hex.EncodeToString(hash.Sum(nil)), StorageRef: finalName,
	}, nil
}

// SaveImage writes a verified image artifact into the same candidate store as
// local-executor media. The caller supplies the declared media type; its file
// signature must agree so a PNG cannot be persisted as a JPEG (or vice versa).
func (s *Store) SaveImage(id, mediaType string, src io.Reader) (artifact Artifact, err error) {
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	extension, err := imageExtension(mediaType)
	if err != nil {
		return Artifact{}, err
	}
	if s == nil || !safeArtifactID.MatchString(id) || src == nil {
		return Artifact{}, ErrInvalidID
	}
	if err := s.checkWriteAllowed(); err != nil {
		return Artifact{}, err
	}
	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return Artifact{}, diskPressureError(err)
	}
	finalName := id + extension
	finalPath := filepath.Join(s.root, finalName)
	if _, statErr := os.Stat(finalPath); statErr == nil {
		return Artifact{}, ErrAlreadyExists
	} else if !os.IsNotExist(statErr) {
		return Artifact{}, statErr
	}

	tmp, err := os.CreateTemp(s.root, ".image-upload-*.tmp")
	if err != nil {
		return Artifact{}, diskPressureError(err)
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	first := make([]byte, 16)
	nFirst, readErr := io.ReadFull(src, first)
	if readErr != nil {
		return Artifact{}, ErrInvalidImage
	}
	first = first[:nFirst]
	if !matchesImageMediaType(first, mediaType) {
		return Artifact{}, ErrInvalidImage
	}
	if int64(len(first)) > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}

	hash := sha256.New()
	writer := io.MultiWriter(tmp, hash)
	if _, err := writer.Write(first); err != nil {
		return Artifact{}, diskPressureError(err)
	}
	remainingLimit := s.maxBytes - int64(len(first)) + 1
	copied, err := io.Copy(writer, io.LimitReader(src, remainingLimit))
	if err != nil {
		return Artifact{}, diskPressureError(err)
	}
	total := int64(len(first)) + copied
	if total > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}
	if err := tmp.Sync(); err != nil {
		return Artifact{}, diskPressureError(err)
	}
	if err := tmp.Close(); err != nil {
		return Artifact{}, err
	}
	if err := os.Rename(tmpName, finalPath); err != nil {
		return Artifact{}, fmt.Errorf("commit image artifact: %w", diskPressureError(err))
	}
	committed = true
	return Artifact{ID: id, MediaType: mediaType, ByteSize: total, SHA256: hex.EncodeToString(hash.Sum(nil)), StorageRef: finalName}, nil
}

// SaveMedia stores a browser-uploaded audio/video file after the caller has
// validated its type. Unlike SaveImage/SaveMP4 this method intentionally does
// not inspect a codec signature; browser uploads can be WAV, MP3, M4A, WebM or
// other formats that are only meaningful to the media pipeline.
func (s *Store) SaveMedia(id, mediaType, extension string, src io.Reader) (artifact Artifact, err error) {
	if s == nil || !safeArtifactID.MatchString(id) || src == nil {
		return Artifact{}, ErrInvalidID
	}
	if err := s.checkWriteAllowed(); err != nil {
		return Artifact{}, err
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	extension = strings.ToLower(strings.TrimSpace(extension))
	extension = strings.TrimPrefix(extension, ".")
	if !strings.HasPrefix(mediaType, "audio/") && !strings.HasPrefix(mediaType, "video/") {
		return Artifact{}, ErrInvalidMP4
	}
	if extension == "" || len(extension) > 8 || strings.ContainsAny(extension, `/\\.`) {
		return Artifact{}, ErrInvalidID
	}
	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return Artifact{}, diskPressureError(err)
	}
	finalName := id + "." + extension
	finalPath := filepath.Join(s.root, finalName)
	if _, statErr := os.Stat(finalPath); statErr == nil {
		return Artifact{}, ErrAlreadyExists
	} else if !os.IsNotExist(statErr) {
		return Artifact{}, statErr
	}
	tmp, err := os.CreateTemp(s.root, ".media-upload-*.tmp")
	if err != nil {
		return Artifact{}, diskPressureError(err)
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	hash := sha256.New()
	writer := io.MultiWriter(tmp, hash)
	count, err := io.Copy(writer, io.LimitReader(src, s.maxBytes+1))
	if err != nil {
		return Artifact{}, diskPressureError(err)
	}
	if count == 0 {
		return Artifact{}, ErrInvalidMP4
	}
	if count > s.maxBytes {
		return Artifact{}, ErrTooLarge
	}
	if err := tmp.Sync(); err != nil {
		return Artifact{}, diskPressureError(err)
	}
	if err := tmp.Close(); err != nil {
		return Artifact{}, diskPressureError(err)
	}
	if err := os.Rename(tmpName, finalPath); err != nil {
		return Artifact{}, fmt.Errorf("commit media artifact: %w", diskPressureError(err))
	}
	committed = true
	return Artifact{ID: id, MediaType: mediaType, ByteSize: count, SHA256: hex.EncodeToString(hash.Sum(nil)), StorageRef: finalName}, nil
}

func imageExtension(mediaType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case "image/png":
		return ".png", nil
	case "image/jpeg":
		return ".jpg", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", ErrInvalidImage
	}
}

func matchesImageMediaType(first []byte, mediaType string) bool {
	switch mediaType {
	case "image/png":
		return len(first) >= 8 && string(first[:8]) == "\x89PNG\r\n\x1a\n"
	case "image/jpeg":
		return len(first) >= 3 && first[0] == 0xff && first[1] == 0xd8 && first[2] == 0xff
	case "image/webp":
		return len(first) >= 12 && string(first[:4]) == "RIFF" && string(first[8:12]) == "WEBP"
	default:
		return false
	}
}

func (s *Store) Remove(storageRef string) error {
	if s == nil || filepath.Base(storageRef) != storageRef {
		return ErrInvalidID
	}
	err := os.Remove(filepath.Join(s.root, storageRef))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Open 在本地缓存缺失时触发回源恢复。它使用默认回源超时，供后台 worker 等
// 没有请求上下文的调用点使用；HTTP 请求应优先使用 OpenContext 传入请求 ctx。
func (s *Store) Open(storageRef string) (*os.File, error) {
	if s == nil {
		return nil, ErrInvalidID
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.restoreTimeoutValue())
	defer cancel()
	return s.OpenContext(ctx, storageRef)
}

// OpenContext 与 Open 行为一致，但回源受调用方上下文约束：请求断开或取消时
// 立即停止等待，不再让一个已经没人要的下载继续占用 goroutine。
func (s *Store) OpenContext(ctx context.Context, storageRef string) (*os.File, error) {
	if s == nil || filepath.Base(storageRef) != storageRef {
		return nil, ErrInvalidID
	}
	path := filepath.Join(s.root, storageRef)
	file, err := os.Open(path)
	if err == nil || !os.IsNotExist(err) {
		return file, err
	}
	s.mu.RLock()
	restorer := s.missingRestorer
	s.mu.RUnlock()
	if restorer == nil {
		return nil, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()
	restoreErr := restorer.RestoreStorageRef(ctx, storageRef)
	elapsed := time.Since(start).Round(time.Millisecond)
	if restoreErr != nil {
		log.Printf("artifact restore ref=%s elapsed=%s err=%v", storageRef, elapsed, restoreErr)
		switch {
		case errors.Is(restoreErr, context.DeadlineExceeded):
			return nil, fmt.Errorf("%w: %v", ErrRestoreTimeout, restoreErr)
		case errors.Is(restoreErr, context.Canceled):
			return nil, context.Canceled
		case errors.Is(restoreErr, os.ErrNotExist):
			return nil, os.ErrNotExist
		default:
			return nil, restoreErr
		}
	}
	log.Printf("artifact restore ref=%s elapsed=%s result=ok", storageRef, elapsed)
	return os.Open(path)
}

// Restore atomically rehydrates an exact logical reference from durable
// storage. The final path becomes visible only after byte-size and SHA256
// verification succeed.
func (s *Store) Restore(storageRef string, source io.Reader, expectedSize int64, expectedSHA256 string) (err error) {
	if s == nil || filepath.Base(storageRef) != storageRef || source == nil || expectedSize < 0 || expectedSize > s.maxBytes || !regexp.MustCompile(`^[a-fA-F0-9]{64}$`).MatchString(expectedSHA256) {
		return ErrInvalidID
	}
	if err := s.checkWriteAllowed(); err != nil {
		return err
	}
	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return diskPressureError(err)
	}
	tmp, err := os.CreateTemp(s.root, ".restore-*.tmp")
	if err != nil {
		return diskPressureError(err)
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmp, hash), io.LimitReader(source, expectedSize+1))
	if err != nil {
		return diskPressureError(err)
	}
	if written != expectedSize || !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expectedSHA256) {
		return ErrRemoteIntegrity
	}
	if err := tmp.Sync(); err != nil {
		return diskPressureError(err)
	}
	if err := tmp.Close(); err != nil {
		return diskPressureError(err)
	}
	if err := os.Rename(tmpName, filepath.Join(s.root, storageRef)); err != nil {
		return diskPressureError(err)
	}
	committed = true
	return nil
}

// Path returns a verified, root-confined artifact path for an uploader. It
// never accepts traversal or a missing object, so callers cannot upload an
// arbitrary server file by crafting a storage reference.
func (s *Store) Path(storageRef string) (string, error) {
	if s == nil {
		return "", ErrInvalidID
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.restoreTimeoutValue())
	defer cancel()
	return s.PathContext(ctx, storageRef)
}

// PathContext 与 Path 一致，但回源受调用方上下文约束。
func (s *Store) PathContext(ctx context.Context, storageRef string) (string, error) {
	path, err := s.LocalPath(storageRef)
	if err == nil || !os.IsNotExist(err) {
		return path, err
	}
	file, openErr := s.OpenContext(ctx, storageRef)
	if openErr != nil {
		return "", openErr
	}
	if closeErr := file.Close(); closeErr != nil {
		return "", closeErr
	}
	return filepath.Join(s.root, storageRef), nil
}

// LocalPath checks only the server cache. Background discovery uses this
// method so a cleaned TOS-backed object is not downloaded merely because a
// periodic scanner looked at it.
func (s *Store) LocalPath(storageRef string) (string, error) {
	if s == nil || filepath.Base(storageRef) != storageRef {
		return "", ErrInvalidID
	}
	path := filepath.Join(s.root, storageRef)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}
