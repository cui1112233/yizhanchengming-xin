package screenplay

import (
	"context"
	"errors"
	"testing"
)

func TestMediaStateIsolationAndRollback(t *testing.T) {
	s := NewMediaMemoryStore()
	ctx := context.Background()
	_, err := s.MutateMedia(ctx, "a", "p", func(v *MediaState) error { v.Items = append(v.Items, MediaItem{ID: "x"}); return nil })
	if err != nil {
		t.Fatal(err)
	}
	_, _ = s.MutateMedia(ctx, "a", "p", func(v *MediaState) error { v.Items = nil; return errors.New("fail") })
	v, _ := s.LoadMedia(ctx, "a", "p")
	if len(v.Items) != 1 {
		t.Fatal(v)
	}
	v.Items[0].ID = "leak"
	v, _ = s.LoadMedia(ctx, "a", "p")
	if v.Items[0].ID != "x" {
		t.Fatal(v)
	}
	v, _ = s.LoadMedia(ctx, "b", "p")
	if len(v.Items) != 0 {
		t.Fatal("cross actor leak")
	}
}
