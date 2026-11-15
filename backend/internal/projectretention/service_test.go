package projectretention

import (
	"testing"
	"time"
)

func TestExpiresAtUsesLocalMidnight(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	service := &Service{Location: loc}
	last := time.Date(2026, 11, 10, 15, 30, 0, 0, loc)
	want := time.Date(2026, 11, 17, 0, 0, 0, 0, loc).UTC()
	if got := service.expiresAt(last, 7); !got.Equal(want) {
		t.Fatalf("expiresAt=%s want=%s", got, want)
	}
}

func TestExpiresAtKeepsEarlyMorningInTheSameDay(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	service := &Service{Location: loc}
	last := time.Date(2026, 11, 10, 0, 5, 0, 0, loc)
	want := time.Date(2026, 11, 13, 0, 0, 0, 0, loc).UTC()
	if got := service.expiresAt(last, 3); !got.Equal(want) {
		t.Fatalf("expiresAt=%s want=%s", got, want)
	}
}

func TestRemainingDaysOnlyWarnsOnLastDay(t *testing.T) {
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		expires time.Time
		want    int
	}{
		{time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC), 1},
		{time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC), 2},
		{time.Date(2026, 11, 16, 0, 0, 0, 0, time.UTC), 0},
	}
	for _, item := range cases {
		if got := remainingDays(item.expires, now); got != item.want {
			t.Fatalf("remainingDays(%s)=%d want=%d", item.expires, got, item.want)
		}
	}
}

func TestValidModuleRejectsUnknownValues(t *testing.T) {
	for _, module := range []Module{ModuleBatchFactory, ModuleComicScreenplay, ModuleCMCanvas, ModuleNovelCreation} {
		if !validModule(module) {
			t.Fatalf("module %s should be valid", module)
		}
	}
	if validModule(Module("unknown")) {
		t.Fatal("unknown module must be rejected")
	}
}
