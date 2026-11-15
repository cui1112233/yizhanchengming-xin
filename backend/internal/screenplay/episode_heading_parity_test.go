package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestEpisodeHeadingParityPreservesPreambleAndEmptyEpisodes(t *testing.T) {
	text := "故事前言\r\n第1集：相遇\r\n对白：第2集会再见。\r\n第2集会再见。\r\n第十二章 章节文字\r\n第九回 回目\r\n## 第 二 集\r\n### 第十二集 转折\r\n完整尾句。"
	rows := splitEpisodes(text)
	if len(rows) != 3 {
		t.Fatalf("episodes=%+v, want 3", rows)
	}
	for i, no := range []int{1, 2, 12} {
		if rows[i].No != no {
			t.Errorf("episode %d=%d, want %d", i, rows[i].No, no)
		}
	}
	if !strings.HasPrefix(rows[0].Content, "故事前言\r\n第1集：相遇") || !strings.Contains(rows[0].Content, "第2集会再见。") || !strings.Contains(rows[0].Content, "第十二章 章节文字") {
		t.Fatalf("preamble/body lost: %q", rows[0].Content)
	}
	if rows[1].Content != "## 第 二 集" {
		t.Fatalf("empty episode lost: %q", rows[1].Content)
	}
	if rows[2].Content != "### 第十二集 转折\r\n完整尾句。" {
		t.Fatalf("tail=%q", rows[2].Content)
	}
}
func TestEpisodeHeadingParityChineseNumbers(t *testing.T) {
	rows := splitEpisodes("第十一集\na\n第两百零三集：标题\nb\n第 二〇四 集\nc")
	if len(rows) != 3 {
		t.Fatalf("rows=%+v", rows)
	}
	for i, n := range []int{11, 203, 204} {
		if rows[i].No != n {
			t.Errorf("no=%d want %d", rows[i].No, n)
		}
	}
}
func TestEpisodeHeadingParityDuplicateRejectedBeforeGeneration(t *testing.T) {
	for _, selected := range []int{0, 1} {
		t.Run(string(rune('0'+selected)), func(t *testing.T) {
			service, store, generator := newStoryboardService(t, PermissionBypass)
			_, err := store.SaveScreenplay(context.Background(), Document{ID: "doc", ProjectID: "p1", OwnerUsername: "alice", Source: "b", Content: "第1集\n甲\n# 第一集\n乙"}, nil, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			_, err = service.Generate(context.Background(), StoryboardRequest{Owner: "alice", ProjectID: "p1", Engine: EngineXiaoyunqueV3, EpisodeNo: selected, IdempotencyKey: "duplicate-episode"})
			if !errors.Is(err, ErrInvalidArgument) || !strings.Contains(err.Error(), "重复") {
				t.Fatalf("err=%v, want duplicate invalid argument", err)
			}
			if generator.callCount() != 0 {
				t.Fatalf("provider calls=%d", generator.callCount())
			}
			rows, _ := store.ListStoryboards(context.Background(), "alice", "p1", EngineXiaoyunqueV3)
			if len(rows) != 0 {
				t.Fatalf("saved=%+v", rows)
			}
		})
	}
}
func TestEpisodeHeadingParityEmptyBodyRejectedBeforeAnyGeneration(t *testing.T) {
	cases := []struct {
		name, content string
		selected      int
		emptyNo       string
	}{
		{"plain-single", "第二集", 2, "第2集"},
		{"markdown-single", "# 第2集：仅标题", 2, "第2集"},
		{"preamble-not-body", "这是作品前言。\n第二集：仅标题", 2, "第2集"},
		{"batch-preflight", "第1集\n完整剧本正文。\n第二集", 0, "第2集"},
		{"markdown-batch-preflight", "# 第1集\n完整剧本正文。\n## 第2集：仅标题", 0, "第2集"},
		{"preamble-empty-first", "前言。\n第一集\n第二集\n完整剧本正文。", 0, "第1集"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, store, generator := newStoryboardService(t, PermissionBypass)
			ctx := context.Background()
			if _, err := store.SaveScreenplay(ctx, Document{ID: "empty-doc", ProjectID: "p1", OwnerUsername: "alice", Source: "b", Content: tc.content}, nil, nil, nil); err != nil {
				t.Fatal(err)
			}
			_, err := service.Generate(ctx, StoryboardRequest{Owner: "alice", ProjectID: "p1", Engine: EngineXiaoyunqueV3, EpisodeNo: tc.selected, IdempotencyKey: "empty-body"})
			if !errors.Is(err, ErrEmptyContent) || !strings.Contains(err.Error(), tc.emptyNo) {
				t.Errorf("err=%v, want empty content with %s", err, tc.emptyNo)
			}
			if generator.callCount() != 0 {
				t.Errorf("provider calls=%d, want zero", generator.callCount())
			}
			boards, _ := store.ListStoryboards(ctx, "alice", "p1", EngineXiaoyunqueV3)
			if len(boards) != 0 {
				t.Errorf("saved=%+v, want none", boards)
			}
			doc, err := store.ActiveDocument(ctx, "alice", "p1")
			if err != nil || doc.Content != tc.content {
				t.Errorf("saved document changed: %+v/%v", doc, err)
			}
		})
	}
}
