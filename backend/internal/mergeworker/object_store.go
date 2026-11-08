package mergeworker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

type ObjectStore interface {
	PutMerged(context.Context, string, string) (string, error)
}

// ObjectReader exposes a private object through the authenticated Batch
// Factory API. Public/CDN deployments can keep returning their public URL.
type ObjectReader interface {
	Open(context.Context, string) (io.ReadCloser, error)
}

type PutFileFunc func(context.Context, string, string, string, string) error

type TOSObjectStore struct {
	Bucket        string
	PublicBaseURL string
	PutFile       PutFileFunc
	OpenFile      func(context.Context, string) (io.ReadCloser, error)
	HeadFile      func(context.Context, string) (int64, string, error)
}

var safeTaskID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

func MergeObjectKey(taskID string) (string, error) {
	taskID = strings.TrimSpace(taskID)
	if !safeTaskID.MatchString(taskID) {
		return "", fmt.Errorf("invalid merge task id")
	}
	return "batch-merged/" + taskID + ".mp4", nil
}

func UploadedObjectKey(uploadID string) (string, error) {
	uploadID = strings.TrimSpace(uploadID)
	if !safeTaskID.MatchString(uploadID) {
		return "", fmt.Errorf("invalid upload id")
	}
	return "batch-uploaded/" + uploadID + ".mp4", nil
}

func (s *TOSObjectStore) PutUploadedMP4(ctx context.Context, uploadID, filePath string) (string, error) {
	if s == nil || s.PutFile == nil || strings.TrimSpace(s.Bucket) == "" {
		return "", fmt.Errorf("TOS output store is not configured")
	}
	key, err := UploadedObjectKey(uploadID)
	if err != nil {
		return "", err
	}
	base, err := url.Parse(strings.TrimSpace(s.PublicBaseURL))
	if strings.TrimSpace(s.PublicBaseURL) == "" || err != nil || base.Scheme != "https" || base.Host == "" {
		return "", fmt.Errorf("TOS public base URL must be HTTPS for uploaded video")
	}
	if err := s.PutFile(ctx, strings.TrimSpace(s.Bucket), key, filePath, "video/mp4"); err != nil {
		return "", fmt.Errorf("upload video failed: %w", err)
	}
	if s.HeadFile != nil {
		localSize, localSHA, err := fileSizeAndSHA256(filePath)
		if err != nil {
			return "", fmt.Errorf("verify local uploaded video: %w", err)
		}
		remoteSize, remoteSHA, err := s.HeadFile(ctx, key)
		if err != nil {
			return "", fmt.Errorf("verify uploaded video: %w", err)
		}
		if remoteSize != localSize || !strings.EqualFold(remoteSHA, localSHA) {
			return "", fmt.Errorf("verify uploaded video: size or checksum mismatch")
		}
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + "/" + path.Clean(key)
	base.RawQuery = ""
	base.Fragment = ""
	return base.String(), nil
}

func (s *TOSObjectStore) PutMerged(ctx context.Context, taskID, filePath string) (string, error) {
	if s == nil || s.PutFile == nil || strings.TrimSpace(s.Bucket) == "" {
		return "", fmt.Errorf("TOS output store is not configured")
	}
	key, err := MergeObjectKey(taskID)
	if err != nil {
		return "", err
	}
	if err := s.PutFile(ctx, strings.TrimSpace(s.Bucket), key, filePath, "video/mp4"); err != nil {
		return "", fmt.Errorf("upload merged video failed: %w", err)
	}
	if s.HeadFile != nil {
		localSize, localSHA, err := fileSizeAndSHA256(filePath)
		if err != nil {
			return "", fmt.Errorf("verify local merged video: %w", err)
		}
		remoteSize, remoteSHA, err := s.HeadFile(ctx, key)
		if err != nil {
			return "", fmt.Errorf("verify uploaded merged video: %w", err)
		}
		if remoteSize != localSize || !strings.EqualFold(remoteSHA, localSHA) {
			return "", fmt.Errorf("verify uploaded merged video: size or checksum mismatch")
		}
	} else if s.OpenFile != nil {
		reader, err := s.OpenFile(ctx, key)
		if err != nil {
			return "", fmt.Errorf("verify uploaded merged video: %w", err)
		}
		if err := reader.Close(); err != nil {
			return "", fmt.Errorf("verify uploaded merged video: %w", err)
		}
	}
	if strings.TrimSpace(s.PublicBaseURL) == "" {
		return "tos://" + strings.TrimSpace(s.Bucket) + "/" + key, nil
	}
	base, err := url.Parse(strings.TrimSpace(s.PublicBaseURL))
	if err != nil || (base.Scheme != "https" && base.Scheme != "http") || base.Host == "" {
		return "", fmt.Errorf("invalid TOS public base URL")
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + "/" + path.Clean(key)
	base.RawQuery = ""
	base.Fragment = ""
	return base.String(), nil
}

func fileSizeAndSHA256(filePath string) (int64, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return 0, "", err
	}
	return size, hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *TOSObjectStore) Open(ctx context.Context, taskID string) (io.ReadCloser, error) {
	if s == nil || s.OpenFile == nil {
		return nil, fmt.Errorf("TOS output reader is not configured")
	}
	key, err := MergeObjectKey(taskID)
	if err != nil {
		return nil, err
	}
	return s.OpenFile(ctx, key)
}
