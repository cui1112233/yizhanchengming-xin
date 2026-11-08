package mergeworker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

// These bounds are intentionally longer than the SDK defaults (30s socket,
// 60s response headers). Video output has repeatedly taken several minutes on
// the ECS-to-TOS path; a progressing transfer must not be treated as failed.
// The local merge lifecycle still has its separate one-hour ceiling.
const (
	tosUploadRequestTimeout = 5 * time.Minute
	tosUploadSocketTimeout  = 5 * time.Minute
	tosUploadConnectTimeout = 30 * time.Second
	tosUploadRetryCount     = 3
)

type TOSConfig struct {
	Endpoint      string
	Region        string
	Bucket        string
	AccessKey     string
	SecretKey     string
	PublicBaseURL string
}

func NewTOSObjectStore(config TOSConfig) (*TOSObjectStore, error) {
	config.Endpoint = strings.TrimSpace(config.Endpoint)
	config.Region = strings.TrimSpace(config.Region)
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.AccessKey = strings.TrimSpace(config.AccessKey)
	config.SecretKey = strings.TrimSpace(config.SecretKey)
	config.PublicBaseURL = strings.TrimSpace(config.PublicBaseURL)
	if config.Endpoint == "" || config.Region == "" || config.Bucket == "" || config.AccessKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("TOS merge output configuration is incomplete")
	}
	client, err := tos.NewClientV2(
		config.Endpoint,
		tos.WithRegion(config.Region),
		tos.WithCredentials(tos.NewStaticCredentials(config.AccessKey, config.SecretKey)),
		tos.WithConnectionTimeout(tosUploadConnectTimeout),
		tos.WithRequestTimeout(tosUploadRequestTimeout),
		tos.WithSocketTimeout(tosUploadSocketTimeout, tosUploadSocketTimeout),
		tos.WithMaxRetryCount(tosUploadRetryCount),
	)
	if err != nil {
		return nil, fmt.Errorf("create TOS merge output client: %w", err)
	}
	return &TOSObjectStore{
		Bucket:        config.Bucket,
		PublicBaseURL: config.PublicBaseURL,
		PutFile: func(ctx context.Context, bucket, key, filePath, _ string) error {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			hash := sha256.New()
			size, err := io.Copy(hash, file)
			closeErr := file.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
			digest := hex.EncodeToString(hash.Sum(nil))
			_, err = client.PutObjectFromFile(ctx, &tos.PutObjectFromFileInput{
				PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ContentLength: size, ContentType: "video/mp4", Meta: map[string]string{"qiantie-sha256": digest}},
				FilePath:            filePath,
			})
			return err
		},
		HeadFile: func(ctx context.Context, key string) (int64, string, error) {
			result, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: config.Bucket, Key: key})
			if err != nil {
				return 0, "", err
			}
			digest, _ := result.Meta.Get("qiantie-sha256")
			return result.ContentLength, digest, nil
		},
		OpenFile: func(ctx context.Context, key string) (io.ReadCloser, error) {
			result, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: config.Bucket, Key: key})
			if err != nil {
				return nil, err
			}
			return result.Content, nil
		},
	}, nil
}
