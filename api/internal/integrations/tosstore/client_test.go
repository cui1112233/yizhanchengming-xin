package tosstore

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type sdkFake struct {
	putInput *tos.PutObjectV2Input
	signInput *tos.PreSignedURLInput
}

func (f *sdkFake) PutObjectV2(_ context.Context, input *tos.PutObjectV2Input) (*tos.PutObjectV2Output, error) {
	f.putInput = input
	_, _ = io.ReadAll(input.Content)
	return &tos.PutObjectV2Output{}, nil
}

func (f *sdkFake) PreSignedURL(input *tos.PreSignedURLInput) (*tos.PreSignedURLOutput, error) {
	f.signInput = input
	return &tos.PreSignedURLOutput{SignedUrl: "https://signed.example/object"}, nil
}

func TestNewRequiresCompleteTOSConfig(t *testing.T) {
	for _, cfg := range []Config{
		{},
		{Endpoint: "https://tos-cn-beijing.volces.com", Region: "cn-beijing", AccessKey: "ak", SecretKey: "sk"},
		{Endpoint: "https://tos-cn-beijing.volces.com", Region: "cn-beijing", AccessKey: "ak", Bucket: "bucket"},
	} {
		if _, err := New(cfg); err == nil { t.Fatalf("expected incomplete config to fail: %#v", cfg) }
	}
	if _, err := New(Config{Endpoint: "https://tos-cn-beijing.volces.com", Region: "cn-beijing", AccessKey: "ak", SecretKey: "sk", Bucket: "bucket"}); err != nil {
		t.Fatalf("complete config should construct client: %v", err)
	}
}

func TestClientUploadsIntoConfiguredBucket(t *testing.T) {
	fake := &sdkFake{}
	client := newWithSDK(fake, "prod-media")
	object, err := client.Upload(context.Background(), UploadInput{
		Key: "images/2026/a.png",
		ContentType: "image/png",
		ContentLength: 3,
		Body: strings.NewReader("abc"),
	})
	if err != nil { t.Fatal(err) }
	if object.Bucket != "prod-media" || object.Key != "images/2026/a.png" { t.Fatalf("object=%#v", object) }
	if fake.putInput == nil { t.Fatal("expected PutObjectV2 call") }
	if fake.putInput.Bucket != "prod-media" || fake.putInput.Key != "images/2026/a.png" || fake.putInput.ContentType != "image/png" {
		t.Fatalf("input=%#v", fake.putInput)
	}
}

func TestClientCreatesShortLivedSignedGetURL(t *testing.T) {
	fake := &sdkFake{}
	client := newWithSDK(fake, "prod-media")
	url, err := client.SignedGetURL("images/a.png", 600)
	if err != nil { t.Fatal(err) }
	if url != "https://signed.example/object" { t.Fatalf("url=%q", url) }
	if fake.signInput == nil || string(fake.signInput.HTTPMethod) != http.MethodGet || fake.signInput.Bucket != "prod-media" || fake.signInput.Key != "images/a.png" || fake.signInput.Expires != 600 {
		t.Fatalf("sign input=%#v", fake.signInput)
	}
}

func TestClientRejectsUnsafeEmptyObjectKey(t *testing.T) {
	client := newWithSDK(&sdkFake{}, "prod-media")
	if _, err := client.Upload(context.Background(), UploadInput{Body: strings.NewReader("x")}); err == nil { t.Fatal("expected empty key error") }
	if _, err := client.SignedGetURL("", 600); err == nil { t.Fatal("expected empty key error") }
}
