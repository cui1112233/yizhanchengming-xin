package tosstore

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

const defaultSignedURLSeconds int64 = 900
const maxSignedURLSeconds int64 = 604800

type Config struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
}

type StoredObject struct {
	Bucket string
	Key    string
}

type UploadInput struct {
	Key           string
	ContentType   string
	ContentLength int64
	Body          io.Reader
}

type tosAPI interface {
	PutObjectV2(context.Context, *tos.PutObjectV2Input) (*tos.PutObjectV2Output, error)
	DeleteObjectV2(context.Context, *tos.DeleteObjectV2Input) (*tos.DeleteObjectV2Output, error)
	PreSignedURL(*tos.PreSignedURLInput) (*tos.PreSignedURLOutput, error)
}

type Client struct {
	sdk    tosAPI
	bucket string
}

func New(cfg Config) (*Client, error) {
	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	cfg.Region = strings.TrimSpace(cfg.Region)
	cfg.AccessKey = strings.TrimSpace(cfg.AccessKey)
	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	if cfg.Endpoint == "" || cfg.Region == "" || cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" {
		return nil, errors.New("TOS_ENDPOINT, TOS_REGION, TOS_ACCESS_KEY, TOS_SECRET_KEY and TOS_BUCKET are required")
	}
	sdk, err := tos.NewClientV2(cfg.Endpoint, tos.WithRegion(cfg.Region), tos.WithCredentials(tos.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey)))
	if err != nil { return nil, err }
	return newWithSDK(sdk, cfg.Bucket), nil
}

func newWithSDK(sdk tosAPI, bucket string) *Client {
	return &Client{sdk: sdk, bucket: strings.TrimSpace(bucket)}
}

func (c *Client) Bucket() string {
	if c == nil { return "" }
	return c.bucket
}

func (c *Client) Put(ctx context.Context, key, contentType string, contentLength int64, body io.Reader) (string, string, error) {
	object, err := c.Upload(ctx, UploadInput{Key: key, ContentType: contentType, ContentLength: contentLength, Body: body})
	if err != nil { return "", "", err }
	return object.Bucket, object.Key, nil
}

func (c *Client) Upload(ctx context.Context, input UploadInput) (StoredObject, error) {
	if c == nil || c.sdk == nil || strings.TrimSpace(c.bucket) == "" {
		return StoredObject{}, errors.New("TOS client unavailable")
	}
	key := strings.TrimLeft(strings.TrimSpace(input.Key), "/")
	if key == "" { return StoredObject{}, errors.New("TOS object key is required") }
	if input.Body == nil { return StoredObject{}, errors.New("TOS upload body is required") }
	basic := tos.PutObjectBasicInput{
		Bucket: c.bucket,
		Key: key,
		ContentType: strings.TrimSpace(input.ContentType),
	}
	if input.ContentLength >= 0 { basic.ContentLength = input.ContentLength }
	_, err := c.sdk.PutObjectV2(ctx, &tos.PutObjectV2Input{PutObjectBasicInput: basic, Content: input.Body})
	if err != nil { return StoredObject{}, err }
	return StoredObject{Bucket: c.bucket, Key: key}, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	if c == nil || c.sdk == nil || strings.TrimSpace(c.bucket) == "" {
		return errors.New("TOS client unavailable")
	}
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if key == "" { return errors.New("TOS object key is required") }
	_, err := c.sdk.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{Bucket: c.bucket, Key: key})
	return err
}

func (c *Client) SignedGetURL(key string, expiresSeconds int64) (string, error) {
	if c == nil || c.sdk == nil || strings.TrimSpace(c.bucket) == "" {
		return "", errors.New("TOS client unavailable")
	}
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if key == "" { return "", errors.New("TOS object key is required") }
	if expiresSeconds <= 0 { expiresSeconds = defaultSignedURLSeconds }
	if expiresSeconds > maxSignedURLSeconds { expiresSeconds = maxSignedURLSeconds }
	output, err := c.sdk.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodGet,
		Bucket: c.bucket,
		Key: key,
		Expires: expiresSeconds,
	})
	if err != nil { return "", err }
	if output == nil || strings.TrimSpace(output.SignedUrl) == "" { return "", errors.New("TOS signed URL is empty") }
	return output.SignedUrl, nil
}
