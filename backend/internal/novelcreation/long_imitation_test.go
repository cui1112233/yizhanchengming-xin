package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"sync"
	"testing"
)

func imitationInput(key string) LongImitationInput {
	return LongImitationInput{IdempotencyKey: key, ModelID: "fake-only", ReferenceText: strings.Repeat("甲乙。", 40), StoryIdea: "全新侦探故事", SelectedDirection: "悬疑", TargetWordCount: 50000, TargetChapters: 20, ReaderPreferences: "节奏明快", SupplementaryInfo: "不照搬人物", ThinkingLevel: "deep"}
}
func TestLongImitationExactRecoveredStatistics(t *testing.T) {
	for _, tt := range []struct {
		text                                string
		sentence, std, paragraph, diversity float64
	}{{"甲乙。甲乙。甲乙。", 2, 0, 9, .333}, {"甲。乙乙乙！", 2, 1, 6, .5}, {"😀。😀。", 2, 0, 6, .25}, {"甲。\n\n乙乙。", 1.5, .5, 3, .667}} {
		v := ImitationStyleStatistics(tt.text, "")
		if v.AvgSentenceLength != tt.sentence || v.SentenceLengthStdDev != tt.std || float64(v.AvgParagraphLength) != tt.paragraph || v.VocabularyDiversity != tt.diversity {
			t.Fatalf("text=%q actual=%+v", tt.text, v)
		}
	}
}
func TestLongImitationCreatesAtomicFrozenRequestAndReplays(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := NewService(st, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		return &queuedPipelineProvider{}, nil
	}))
	in := imitationInput("one")
	out, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	if out.Project.TargetChapterWords != 50000 || out.Session.Mode != "long_imitation" || out.Creation.Input.TargetChapters != 20 {
		t.Fatalf("creation mismatch %+v", out)
	}
	again, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil || again.Project.ID != out.Project.ID || again.Run.ID != out.Run.ID {
		t.Fatal("lost reply replay changed identity", e)
	}
	recovered, e := s.GetLongImitationRequest(ctx, "alice", in.IdempotencyKey)
	if e != nil || recovered.Run.ID != out.Run.ID {
		t.Fatal("lookup failed", e)
	}
	if _, e = s.GetLongImitationRequest(ctx, "bob", in.IdempotencyKey); e == nil {
		t.Fatal("cross-user receipt")
	}
	in.StoryIdea = "changed"
	if _, e = s.StartLongImitation(ctx, "alice", in); !IsCode(e, ErrIdempotencyConflict) {
		t.Fatal("changed frozen request accepted", e)
	}
	creation, e := s.GetLongImitation(ctx, "alice", out.Project.ID)
	if e != nil || creation.Input.StoryIdea != "全新侦探故事" {
		t.Fatal("creation not durable", e)
	}
	var prompt map[string]any
	_ = json.Unmarshal(out.Run.PromptSnapshot, &prompt)
	if prompt["long_imitation"] == nil || prompt["checkpoint"] != true {
		t.Fatal("typed full snapshot absent")
	}
	ex, ok, e := st.GetExecution(ctx, "alice", out.Run.ID)
	if e != nil || !ok || ex.State != "ready" {
		t.Fatal("dispatch missing", e)
	}
}
func TestLongImitationInputBoundaryUsesOriginalNonspaceUTF16(t *testing.T) {
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), &stubPipelineRunner{})
	for _, text := range []string{"", strings.Repeat("甲", 99), strings.Repeat(" \uFEFF\n", 100)} {
		in := imitationInput("bad-" + text)
		in.ReferenceText = text
		if _, e := s.StartLongImitation(context.Background(), "alice", in); e == nil {
			t.Fatal("short reference accepted")
		}
	}
	in := imitationInput("emoji")
	in.ReferenceText = strings.Repeat("😀", 50)
	if _, e := s.StartLongImitation(context.Background(), "alice", in); e != nil {
		t.Fatal("100 UTF16 units rejected", e)
	}
}
func TestLongImitationConcurrentIdempotency(t *testing.T) {
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), &stubPipelineRunner{})
	var wg sync.WaitGroup
	ids := make(chan string, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, e := s.StartLongImitation(context.Background(), "alice", imitationInput("same"))
			if e != nil {
				errs <- e
				return
			}
			ids <- out.Run.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for e := range errs {
		t.Fatal(e)
	}
	first := ""
	for id := range ids {
		if first == "" {
			first = id
		}
		if id != first {
			t.Fatal("duplicate runs")
		}
	}
}
func TestLongImitationNumericAndReferenceCaps(t *testing.T) {
	for _, tt := range []struct {
		words, chapters int
		valid           bool
	}{{1000, 1, true}, {50000, 500, true}, {0, 0, true}, {999, 1, false}, {50001, 1, false}, {5000, 501, false}, {5000, -1, false}} {
		in := imitationInput("boundary")
		in.TargetWordCount = tt.words
		in.TargetChapters = tt.chapters
		normalized, e := normalizeLongImitation(in)
		if (e == nil) != tt.valid {
			t.Fatalf("words=%d chapters=%d err=%v", tt.words, tt.chapters, e)
		}
		if tt.valid && tt.words == 0 && (normalized.TargetWordCount != 5000 || normalized.TargetChapters != 100) {
			t.Fatal("original defaults changed")
		}
	}
	in := imitationInput("oversize")
	in.ReferenceText = strings.Repeat(" ", LongImitationMaxUTF16) + strings.Repeat("甲", 100)
	if _, e := normalizeLongImitation(in); e == nil {
		t.Fatal("raw whitespace bypassed UTF16 whole-input cap")
	}
}
func TestLongImitationFullLengthExternalKeyFitsInternalRunColumn(t *testing.T) {
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), nil)
	in := imitationInput(strings.Repeat("键", 128))
	out, e := s.StartLongImitation(context.Background(), "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	if len(out.Run.IdempotencyKey) > 96 {
		t.Fatalf("internal run key exceeds SQL VARCHAR(96): bytes=%d", len(out.Run.IdempotencyKey))
	}
	want := "long-imitation:" + fingerprint([]string{"alice", longImitationKind, in.IdempotencyKey})
	if out.Run.IdempotencyKey != want {
		t.Fatal("internal binding must cover owner/kind/original key")
	}
	again, e := s.GetLongImitationRequest(context.Background(), "alice", in.IdempotencyKey)
	if e != nil || again.Run.ID != out.Run.ID || again.Creation.Input.IdempotencyKey != in.IdempotencyKey {
		t.Fatal("external receipt key changed", e)
	}
}
