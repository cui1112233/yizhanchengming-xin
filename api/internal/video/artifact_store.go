package video

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type ObjectUploader interface {
	PutObjectFromFile(context.Context, string, string, string) error
}

type ArtifactStoreConfig struct {
	Bucket                string
	PublicBaseURL         string
	HTTPClient            *http.Client
	Uploader              ObjectUploader
	AllowInsecureLoopback bool
}

type HTTPArtifactStore struct {
	bucket                string
	publicBaseURL         string
	client                *http.Client
	uploader              ObjectUploader
	allowInsecureLoopback bool
}

func NewArtifactStore(config ArtifactStoreConfig) (*HTTPArtifactStore, error) {
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.PublicBaseURL = strings.TrimRight(strings.TrimSpace(config.PublicBaseURL), "/")
	if config.Bucket == "" || config.PublicBaseURL == "" || config.Uploader == nil {
		return nil, fmt.Errorf("video: artifact store configuration is incomplete")
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 10 * time.Minute}
	}
	return &HTTPArtifactStore{bucket: config.Bucket, publicBaseURL: config.PublicBaseURL, client: config.HTTPClient, uploader: config.Uploader, allowInsecureLoopback: config.AllowInsecureLoopback}, nil
}

func (s *HTTPArtifactStore) Persist(ctx context.Context, sourceURL, objectHint string) (Artifact, error) {
	u, err := url.Parse(strings.TrimSpace(sourceURL))
	if err != nil || u.Host == "" || u.User != nil {
		return Artifact{}, fmt.Errorf("video: provider artifact URL is invalid")
	}
	secure := u.Scheme == "https"
	loopback := u.Scheme == "http" && s.allowInsecureLoopback && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost" || u.Hostname() == "::1")
	if !secure && !loopback {
		return Artifact{}, fmt.Errorf("video: provider artifact URL must use https")
	}
	key, err := validateArtifactObjectKey(objectHint)
	if err != nil {
		return Artifact{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Artifact{}, fmt.Errorf("video: create artifact download request: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return Artifact{}, fmt.Errorf("video: download provider artifact: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Artifact{}, fmt.Errorf("video: provider artifact download returned HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "task14-video-*.bin")
	if err != nil {
		return Artifact{}, fmt.Errorf("video: create artifact temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	const maxArtifactBytes int64 = 2 << 30
	written, copyErr := io.Copy(tmp, io.LimitReader(resp.Body, maxArtifactBytes+1))
	closeErr := tmp.Close()
	if copyErr != nil {
		return Artifact{}, fmt.Errorf("video: download provider artifact: %w", copyErr)
	}
	if closeErr != nil {
		return Artifact{}, fmt.Errorf("video: close artifact temp file: %w", closeErr)
	}
	if written > maxArtifactBytes {
		return Artifact{}, fmt.Errorf("video: provider artifact exceeds 2 GiB phase-1 limit")
	}
	return s.persistFile(ctx, tmpPath, key)
}

// PersistFile uploads an already materialized local artifact to TOS. It is used
// by the merge executor so ffmpeg output does not need to be exposed through a
// temporary public URL before becoming durable.
func (s *HTTPArtifactStore) PersistFile(ctx context.Context, sourcePath, objectHint string) (Artifact, error) {
	key, err := validateArtifactObjectKey(objectHint)
	if err != nil {
		return Artifact{}, err
	}
	cleanPath := filepath.Clean(strings.TrimSpace(sourcePath))
	if cleanPath == "." || cleanPath == "" {
		return Artifact{}, fmt.Errorf("video: artifact source file is invalid")
	}
	info, err := os.Stat(cleanPath)
	if err != nil {
		return Artifact{}, fmt.Errorf("video: stat artifact source file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return Artifact{}, fmt.Errorf("video: artifact source must be a regular file")
	}
	return s.persistFile(ctx, cleanPath, key)
}

func (s *HTTPArtifactStore) persistFile(ctx context.Context, sourcePath, key string) (Artifact, error) {
	if err := s.uploader.PutObjectFromFile(ctx, s.bucket, key, sourcePath); err != nil {
		return Artifact{}, fmt.Errorf("video: upload artifact to TOS: %w", err)
	}
	return Artifact{Bucket: s.bucket, ObjectKey: key, URL: s.publicBaseURL + "/" + key}, nil
}

func validateArtifactObjectKey(objectHint string) (string, error) {
	key := strings.TrimLeft(strings.TrimSpace(objectHint), "/")
	if key == "" || strings.Contains(key, "..") {
		return "", fmt.Errorf("video: artifact object key is invalid")
	}
	return key, nil
}

type TOSUploader struct{ client *tos.ClientV2 }

func NewTOSUploader(endpoint, region, accessKey, secretKey string) (*TOSUploader, error) {
	endpoint, region = strings.TrimSpace(endpoint), strings.TrimSpace(region)
	accessKey, secretKey = strings.TrimSpace(accessKey), strings.TrimSpace(secretKey)
	if endpoint == "" || region == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("video: TOS credentials/configuration are incomplete")
	}
	client, err := tos.NewClientV2(endpoint, tos.WithRegion(region), tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)))
	if err != nil {
		return nil, fmt.Errorf("video: create TOS client: %w", err)
	}
	return &TOSUploader{client: client}, nil
}

func (u *TOSUploader) PutObjectFromFile(ctx context.Context, bucket, key, filename string) error {
	_, err := u.client.PutObjectFromFile(ctx, &tos.PutObjectFromFileInput{PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key}, FilePath: filename})
	return err
}

// GetObject is used by the protected Shuihuo asset proxy. The returned stream
// is never exposed as a bucket URL and must be closed by the caller.
func (u *TOSUploader) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	output, err := u.client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	if err != nil {
		return nil, err
	}
	return output.Content, nil
}
