package httpapi

import (
	"strings"
	"testing"
	"time"
)

func TestCollectionNameAIGCPrefixAndLength(t *testing.T) {
	now := time.Date(2026, 11, 25, 16, 1, 0, 0, time.UTC)
	name, err := n8ComicCollectionName(" 邱琬晶 ", " 不宠 ", now)
	if err != nil || name != "AIGC-邱琬晶1126-不宠" {
		t.Fatalf("name=%q err=%v", name, err)
	}

	prefix := "AIGC-邱琬晶1126-"
	name, err = n8ComicCollectionName("邱琬晶", strings.Repeat("文", 100), now)
	want := prefix + strings.Repeat("文", 64-len([]rune(prefix)))
	if err != nil || name != want {
		t.Fatalf("long title name=%q err=%v", name, err)
	}

	if _, err := n8ComicCollectionName(strings.Repeat("人", 54), "不宠", now); err == nil {
		t.Fatal("expected an error when the prefix leaves no room for the title")
	}
}
