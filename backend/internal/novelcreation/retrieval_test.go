package novelcreation

import (
	"context"
	"strings"
	"testing"
)

func retrievalFixture() []Chapter {
	return []Chapter{
		{ID: "c1", Number: 1, Title: "春雨", Content: "雨水敲着窗，林澜没有开灯。信纸上只有一枚褪色邮戳。"},
		{ID: "c2", Number: 2, Title: "旧信", Content: "邮戳来自城南旧邮局，那里十年前就停业了。"},
		{ID: "c3", Number: 3, Title: "车票", Content: "抽屉里有一张旧车票，目的地是北方小城。"},
		{ID: "c4", Number: 4, Title: "雨停", Content: "天快亮时雨停了，楼下有人发动汽车。"},
	}
}

func TestRetrieveRelevantChaptersPrefersKeywordOverlap(t *testing.T) {
	got := retrieveRelevantChapters(retrievalFixture(), "邮戳和旧邮局有什么线索", 5, 3)
	if len(got) == 0 {
		t.Fatal("no chapters retrieved")
	}
	hit := false
	for _, chapter := range got {
		if chapter.Number == 1 || chapter.Number == 2 {
			hit = true
		}
		if chapter.Excerpt == "" {
			t.Fatalf("empty excerpt: %+v", chapter)
		}
	}
	if !hit {
		t.Fatalf("keyword chapters missing: %+v", got)
	}
	if len(got) > 3 {
		t.Fatalf("limit not applied: %d", len(got))
	}
}

func TestRetrieveRelevantChaptersExcludesTargetAndFuture(t *testing.T) {
	got := retrieveRelevantChapters(retrievalFixture(), "雨", 3, 3)
	for _, chapter := range got {
		if chapter.Number >= 3 {
			t.Fatalf("target or future chapter leaked: %+v", got)
		}
	}
}

func TestRetrieveRelevantChaptersKeepsOnlyRecentWhenNoOverlap(t *testing.T) {
	got := retrieveRelevantChapters(retrievalFixture(), "完全无关的词语", 6, 3)
	if len(got) == 0 {
		t.Fatal("recent chapters should still be available")
	}
	for _, chapter := range got {
		if chapter.Number < 2 {
			t.Fatalf("distant unrelated chapter leaked: %+v", got)
		}
	}
}

func TestRetrievalQueryFallsBackToLatestChapter(t *testing.T) {
	chapters := retrievalFixture()
	if query := retrievalQuery("  ", chapters); !strings.Contains(query, "雨停") {
		t.Fatalf("fallback query=%q", query)
	}
	if query := retrievalQuery("写下一章", chapters); query != "写下一章" {
		t.Fatalf("explicit query=%q", query)
	}
}

func TestRetrievalInjectionIsVersionGated(t *testing.T) {
	retrieved := []RetrievedChapter{{Number: 1, Title: "春雨", Excerpt: "雨水敲着窗", Score: 4.2}}
	for _, tc := range []struct {
		version int
		want    bool
	}{{0, false}, {1, true}} {
		input := fixturePipelineRunContext(NewMemoryRuntime())
		a := NormalizeAgentSettings(AgentSettings{}, 2000)
		a.EnableLengthNorm = false
		input.Agent = &a
		input.RetrievalVersion = tc.version
		input.RetrievedChapters = retrieved
		provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
		if _, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{}); e != nil {
			t.Fatal(e)
		}
		injected := 0
		for _, request := range provider.requests {
			if strings.Contains(request.UserPrompt, "retrieved_chapters") {
				injected++
			}
		}
		if tc.want && injected < 2 {
			t.Fatalf("planner and writer must both receive retrieval: %d", injected)
		}
		if !tc.want && injected != 0 {
			t.Fatalf("version 0 must not inject retrieval: %d", injected)
		}
	}
}
