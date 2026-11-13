package httpapi

import (
	"context"
	"qiantie/backend/internal/novelfetchworkshop"
	"testing"
)

func TestWorkshopSubmitGuardSharedAcrossCallers(t *testing.T) {
	values := map[string]string{}
	queue := &novelfetchworkshop.RedisQueue{Prefix: "test", Do: func(_ context.Context, args ...string) (any, error) {
		if args[0] == "EVAL" && len(args) == 6 {
			if _, ok := values[args[3]]; ok {
				return int64(0), nil
			}
			values[args[3]] = args[4]
			return int64(1), nil
		}
		if args[0] == "EVAL" {
			if values[args[3]] == args[4] {
				delete(values, args[3])
			}
			return int64(1), nil
		}
		t.Fatalf("unexpected command %v", args)
		return nil, nil
	}}
	release, err := acquireWorkshopSubmit(context.Background(), queue, "owner", "book")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := acquireWorkshopSubmit(context.Background(), queue, "owner", "book"); err == nil {
		t.Fatal("duplicate upload allowed")
	}
	if _, err := acquireWorkshopSubmit(context.Background(), queue, "owner", "different-book"); err == nil {
		t.Fatal("same account uploaded another book concurrently")
	}
	other, err := acquireWorkshopSubmit(context.Background(), queue, "other", "book")
	if err != nil {
		t.Fatal("owner isolation broken")
	}
	other()
	release()
	again, err := acquireWorkshopSubmit(context.Background(), queue, "owner", "book")
	if err != nil {
		t.Fatal("completed lock not released")
	}
	again()
}
