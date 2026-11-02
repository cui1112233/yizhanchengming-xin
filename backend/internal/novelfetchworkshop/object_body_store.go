package novelfetchworkshop

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type BodyObjectStore struct {
	Bucket string
	Put    func(context.Context, string, string, []byte) error
	Get    func(context.Context, string, string) ([]byte, error)
	Delete func(context.Context, string, string) error
}

type BodyTOSConfig struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

var unsafeBodyKeyPart = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func BodyObjectKey(owner, bookID, versionID string) string {
	ownerHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(owner))))
	bookID = strings.Trim(unsafeBodyKeyPart.ReplaceAllString(strings.TrimSpace(bookID), "_"), "._-")
	versionID = strings.Trim(unsafeBodyKeyPart.ReplaceAllString(strings.TrimSpace(versionID), "_"), "._-")
	return "novel-fetch/" + ownerHash[:24] + "/" + bookID + "/" + versionID + ".txt"
}

func NewBodyTOSStore(config BodyTOSConfig) (*BodyObjectStore, error) {
	config.Endpoint = strings.TrimSpace(config.Endpoint)
	config.Region = strings.TrimSpace(config.Region)
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.AccessKey = strings.TrimSpace(config.AccessKey)
	config.SecretKey = strings.TrimSpace(config.SecretKey)
	if config.Endpoint == "" || config.Region == "" || config.Bucket == "" || config.AccessKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("novel fetch TOS configuration is incomplete")
	}
	client, err := tos.NewClientV2(config.Endpoint, tos.WithRegion(config.Region), tos.WithCredentials(tos.NewStaticCredentials(config.AccessKey, config.SecretKey)), tos.WithConnectionTimeout(30*time.Second), tos.WithRequestTimeout(2*time.Minute), tos.WithSocketTimeout(2*time.Minute, 2*time.Minute), tos.WithMaxRetryCount(3))
	if err != nil {
		return nil, fmt.Errorf("create novel fetch TOS client: %w", err)
	}
	return &BodyObjectStore{
		Bucket: config.Bucket,
		Put: func(ctx context.Context, bucket, key string, body []byte) error {
			_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ContentType: "text/plain; charset=utf-8"}, Content: bytes.NewReader(body)})
			return err
		},
		Get: func(ctx context.Context, bucket, key string) ([]byte, error) {
			result, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
			if err != nil {
				return nil, err
			}
			defer result.Content.Close()
			return io.ReadAll(io.LimitReader(result.Content, 64<<20))
		},
		Delete: func(ctx context.Context, bucket, key string) error {
			_, err := client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{Bucket: bucket, Key: key})
			return err
		},
	}, nil
}

func (store *BodyObjectStore) PutContent(ctx context.Context, key, content string) error {
	if store == nil || store.Put == nil || strings.TrimSpace(store.Bucket) == "" {
		return fmt.Errorf("novel fetch TOS body store is not configured")
	}
	return store.Put(ctx, store.Bucket, key, []byte(content))
}

func (store *BodyObjectStore) GetContent(ctx context.Context, key string) (string, error) {
	if store == nil || store.Get == nil || strings.TrimSpace(store.Bucket) == "" {
		return "", fmt.Errorf("novel fetch TOS body store is not configured")
	}
	body, err := store.Get(ctx, store.Bucket, key)
	return string(body), err
}

func (store *BodyObjectStore) DeleteContent(ctx context.Context, key string) error {
	if store == nil || store.Delete == nil || strings.TrimSpace(store.Bucket) == "" {
		return fmt.Errorf("novel fetch TOS body store is not configured")
	}
	return store.Delete(ctx, store.Bucket, key)
}
