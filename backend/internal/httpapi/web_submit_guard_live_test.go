package httpapi

import (
	"context"
	"os"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/novelfetchworkshop"
	"testing"
	"time"
)

func TestLiveRedisAccountUploadGuard(t *testing.T) {
	url := os.Getenv("V88_GUARD_TEST_REDIS_URL")
	if url == "" {
		t.Skip("explicit Redis test endpoint required")
	}
	transport, err := mergeworker.NewRedisStore(url, "guard-acceptance")
	if err != nil {
		t.Fatal(err)
	}
	q := novelfetchworkshop.NewRedisQueueWithDo("guard-acceptance", novelfetchworkshop.RedisDoFunc(transport.Do))
	owner := "test-" + time.Now().UTC().Format("20060102150405.000000000")
	release, err := acquireWorkshopSubmit(context.Background(), q, owner, "book-a")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := acquireWorkshopSubmit(context.Background(), q, owner, "book-b"); err == nil {
		t.Fatal("concurrent upload allowed")
	}
	release()
	next, err := acquireWorkshopSubmit(context.Background(), q, owner, "book-b")
	if err != nil {
		t.Fatal(err)
	}
	defer next()
	release() // An old request must not release a newer request's lock.
	if _, err := acquireWorkshopSubmit(context.Background(), q, owner, "book-c"); err == nil {
		t.Fatal("old holder released new lock")
	}
}
