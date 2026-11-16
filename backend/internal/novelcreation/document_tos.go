package novelcreation

import (
	"context"
	"fmt"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"io"
	"strings"
	"time"
)

type DocumentTOSConfig struct{ Endpoint, Region, Bucket, AccessKey, SecretKey string }
type TOSDocumentStore struct {
	client *tos.ClientV2
	bucket string
}

func NewTOSDocumentStore(c DocumentTOSConfig) (*TOSDocumentStore, error) {
	if strings.TrimSpace(c.Endpoint) == "" || strings.TrimSpace(c.Region) == "" || strings.TrimSpace(c.Bucket) == "" || strings.TrimSpace(c.AccessKey) == "" || strings.TrimSpace(c.SecretKey) == "" {
		return nil, fmt.Errorf("小说资料TOS配置不完整")
	}
	client, e := tos.NewClientV2(c.Endpoint, tos.WithRegion(c.Region), tos.WithCredentials(tos.NewStaticCredentials(c.AccessKey, c.SecretKey)), tos.WithConnectionTimeout(15*time.Second), tos.WithRequestTimeout(time.Minute), tos.WithMaxRetryCount(2))
	if e != nil {
		return nil, e
	}
	return &TOSDocumentStore{client: client, bucket: c.Bucket}, nil
}
func (s *TOSDocumentStore) Put(ctx context.Context, o DocumentObject, r io.Reader) error {
	_, e := s.client.PutObjectV2(ctx, &tos.PutObjectV2Input{PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: s.bucket, Key: o.Key, ContentType: o.ContentType}, Content: r})
	return e
}
func (s *TOSDocumentStore) Open(ctx context.Context, o DocumentObject) (io.ReadCloser, error) {
	r, e := s.client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: s.bucket, Key: o.Key})
	if e != nil {
		return nil, e
	}
	return r.Content, nil
}
func (s *TOSDocumentStore) Delete(ctx context.Context, o DocumentObject) error {
	_, e := s.client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{Bucket: s.bucket, Key: o.Key})
	return e
}
