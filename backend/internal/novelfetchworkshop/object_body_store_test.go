package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestBodyObjectStoreRoundTrip(t *testing.T) {
	objects := map[string][]byte{}
	store := &BodyObjectStore{
		Bucket: "novel-bodies",
		Put: func(_ context.Context, bucket, key string, body []byte) error {
			if bucket != "novel-bodies" {
				t.Fatalf("bucket=%q", bucket)
			}
			objects[key] = append([]byte(nil), body...)
			return nil
		},
		Get: func(_ context.Context, bucket, key string) ([]byte, error) {
			return append([]byte(nil), objects[key]...), nil
		},
		Delete: func(_ context.Context, bucket, key string) error {
			delete(objects, key)
			return nil
		},
	}
	key := BodyObjectKey("alice@example.com", "12345678901", "original_raw")
	if err := store.PutContent(context.Background(), key, "完整正文"); err != nil {
		t.Fatal(err)
	}
	content, err := store.GetContent(context.Background(), key)
	if err != nil || content != "完整正文" {
		t.Fatalf("content=%q err=%v", content, err)
	}
	if err := store.DeleteContent(context.Background(), key); err != nil {
		t.Fatal(err)
	}
}

func TestBodyObjectKeyDoesNotExposeOwner(t *testing.T) {
	key := BodyObjectKey("alice@example.com", "12345678901", "ai1")
	if key == "" || key == "novel-fetch/alice@example.com/12345678901/ai1.txt" {
		t.Fatalf("unsafe object key=%q", key)
	}
}
