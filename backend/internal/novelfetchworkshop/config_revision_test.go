package novelfetchworkshop

import (
	"context"
	"errors"
	"testing"
)

func TestConfigRevisionRejectsLostUpdateAndPreservesNewData(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	a, _ := s.GetConfig(ctx, "alice")
	b, _ := s.GetConfig(ctx, "alice")
	a["first"] = "new"
	if err := s.PutConfig(ctx, "alice", a); err != nil {
		t.Fatal(err)
	}
	b["second"] = "stale"
	if err := s.PutConfig(ctx, "alice", b); !errors.Is(err, ErrConfigConflict) {
		t.Fatalf("stale writer accepted: %v", err)
	}
	saved, _ := s.GetConfig(ctx, "alice")
	if saved["first"] != "new" || saved["second"] != nil {
		t.Fatal("new data overwritten")
	}
	c, _ := s.GetConfig(ctx, "bob")
	if err := s.PutConfig(ctx, "bob", c); err != nil {
		t.Fatal("another account blocked", err)
	}
}
