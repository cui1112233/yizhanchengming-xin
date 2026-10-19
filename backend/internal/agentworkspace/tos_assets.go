package agentworkspace

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type AssetTOSConfig struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

type TOSAssetObjectStore struct {
	Bucket string
	put    func(context.Context, Asset, io.Reader) error
	open   func(context.Context, Asset) (io.ReadCloser, error)
	delete func(context.Context, Asset) error
}

func NewTOSAssetObjectStore(config AssetTOSConfig) (*TOSAssetObjectStore, error) {
	config.Endpoint = strings.TrimSpace(config.Endpoint)
	config.Region = strings.TrimSpace(config.Region)
	config.Bucket = strings.TrimSpace(config.Bucket)
	config.AccessKey = strings.TrimSpace(config.AccessKey)
	config.SecretKey = strings.TrimSpace(config.SecretKey)
	if config.Endpoint == "" || config.Region == "" || config.Bucket == "" || config.AccessKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("agent asset TOS configuration is incomplete")
	}
	client, err := tos.NewClientV2(config.Endpoint, tos.WithRegion(config.Region), tos.WithCredentials(tos.NewStaticCredentials(config.AccessKey, config.SecretKey)), tos.WithConnectionTimeout(30*time.Second), tos.WithRequestTimeout(5*time.Minute), tos.WithSocketTimeout(5*time.Minute, 5*time.Minute), tos.WithMaxRetryCount(3))
	if err != nil {
		return nil, fmt.Errorf("create agent asset TOS client: %w", err)
	}
	return &TOSAssetObjectStore{
		Bucket: config.Bucket,
		put: func(ctx context.Context, asset Asset, body io.Reader) error {
			_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: config.Bucket, Key: asset.ObjectKey, ContentType: asset.ContentType}, Content: body})
			return err
		},
		open: func(ctx context.Context, asset Asset) (io.ReadCloser, error) {
			result, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: config.Bucket, Key: asset.ObjectKey})
			if err != nil {
				return nil, err
			}
			return result.Content, nil
		},
		delete: func(ctx context.Context, asset Asset) error {
			_, err := client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{Bucket: config.Bucket, Key: asset.ObjectKey})
			return err
		},
	}, nil
}

func (s *TOSAssetObjectStore) Put(ctx context.Context, asset Asset, body io.Reader) error {
	if s == nil || s.put == nil || strings.TrimSpace(s.Bucket) == "" {
		return fmt.Errorf("agent asset object store is not configured")
	}
	return s.put(ctx, asset, body)
}

func (s *TOSAssetObjectStore) Open(ctx context.Context, asset Asset) (io.ReadCloser, error) {
	if s == nil || s.open == nil || strings.TrimSpace(s.Bucket) == "" {
		return nil, fmt.Errorf("agent asset object store is not configured")
	}
	return s.open(ctx, asset)
}

func (s *TOSAssetObjectStore) Delete(ctx context.Context, asset Asset) error {
	if s == nil || s.delete == nil || strings.TrimSpace(s.Bucket) == "" {
		return fmt.Errorf("agent asset object store is not configured")
	}
	return s.delete(ctx, asset)
}
