package productionmedia

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/codes"
)

const tosChecksumMetadataKey = "qiantie-sha256"

type TOSConfig struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

type TOSStore struct {
	bucket string
	client *tos.ClientV2
}

func NewTOSStore(config TOSConfig) (*TOSStore, error) {
	config.Endpoint = strings.TrimSpace(config.Endpoint)
	config.Region = strings.TrimSpace(config.Region)
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.AccessKey = strings.TrimSpace(config.AccessKey)
	config.SecretKey = strings.TrimSpace(config.SecretKey)
	parsed, err := url.Parse(config.Endpoint)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || config.Region == "" || config.Bucket == "" || config.AccessKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("production media TOS configuration is incomplete or not HTTPS")
	}
	client, err := tos.NewClientV2(
		config.Endpoint,
		tos.WithRegion(config.Region),
		tos.WithCredentials(tos.NewStaticCredentials(config.AccessKey, config.SecretKey)),
		tos.WithConnectionTimeout(30*time.Second),
		tos.WithRequestTimeout(5*time.Minute),
		tos.WithSocketTimeout(5*time.Minute, 5*time.Minute),
		tos.WithMaxRetryCount(3),
	)
	if err != nil {
		return nil, fmt.Errorf("create production media TOS client: %w", err)
	}
	return &TOSStore{bucket: config.Bucket, client: client}, nil
}

func (s *TOSStore) PutFile(ctx context.Context, key, path, mediaType, sha string, size int64) error {
	if s == nil || s.client == nil || strings.TrimSpace(key) == "" || size < 0 || len(sha) != 64 {
		return fmt.Errorf("invalid production media upload")
	}
	_, err := s.client.PutObjectFromFile(ctx, &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: s.bucket, Key: key, ContentLength: size, ContentType: mediaType,
			Meta: map[string]string{tosChecksumMetadataKey: sha},
		},
		FilePath: path,
	})
	return err
}

func (s *TOSStore) Head(ctx context.Context, key string) (RemoteMetadata, error) {
	if s == nil || s.client == nil {
		return RemoteMetadata{}, fmt.Errorf("production media TOS store is not configured")
	}
	result, err := s.client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: s.bucket, Key: key})
	if err != nil {
		return RemoteMetadata{}, err
	}
	sha, _ := result.Meta.Get(tosChecksumMetadataKey)
	return RemoteMetadata{ByteSize: result.ContentLength, SHA256: sha, MediaType: result.ContentType}, nil
}

func (s *TOSStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("production media TOS store is not configured")
	}
	result, err := s.client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: s.bucket, Key: key})
	if err != nil {
		return nil, err
	}
	return result.Content, nil
}

// DeleteObject 删除一个已上传的素材对象。对象已经不存在时视为成功，这样作品
// 清理可以安全重试而不会因为上一次删除只完成了一半而永久卡住。
func (s *TOSStore) DeleteObject(ctx context.Context, key string) error {
	if s == nil || s.client == nil || strings.TrimSpace(key) == "" {
		return fmt.Errorf("invalid production media object deletion")
	}
	if _, err := s.client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{Bucket: s.bucket, Key: key}); err != nil {
		if tos.Code(err) == codes.NoSuchKey || tos.StatusCode(err) == http.StatusNotFound {
			return nil
		}
		return err
	}
	return nil
}
