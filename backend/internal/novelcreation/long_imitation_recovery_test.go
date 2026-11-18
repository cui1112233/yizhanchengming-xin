package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"unicode/utf16"
)

type imitationCommitFault struct {
	Store
	fail *bool
}

func (s imitationCommitFault) AtomicCreate(ctx context.Context, owner string, fn func(Store) error) error {
	return s.Store.AtomicCreate(ctx, owner, func(tx Store) error { return fn(imitationCommitFault{tx, s.fail}) })
}
func (s imitationCommitFault) CompleteSubmission(ctx context.Context, owner, kind, key string, r submissionReceipt) error {
	if *s.fail {
		*s.fail = false
		return errors.New("fake receipt save fault")
	}
	return s.Store.CompleteSubmission(ctx, owner, kind, key, r)
}
func TestLongImitationCreationRollbackLeavesNoBookOrQueue(t *testing.T) {
	st := NewMemoryStore()
	fail := true
	s := NewService(imitationCommitFault{st, &fail}, NewMemoryRuntime(), nil)
	in := imitationInput("rollback")
	if _, e := s.StartLongImitation(context.Background(), "alice", in); e == nil {
		t.Fatal("fault not injected")
	}
	projects, _ := st.ListProjects(context.Background(), "alice")
	dispatch, _ := st.ListDispatch(context.Background(), 20)
	if len(projects) != 0 || len(dispatch) != 0 {
		t.Fatal("partial creation survived")
	}
	if _, e := s.GetLongImitationRequest(context.Background(), "alice", in.IdempotencyKey); e == nil {
		t.Fatal("failed creation has successful receipt")
	}
	if _, e := s.StartLongImitation(context.Background(), "alice", in); e != nil {
		t.Fatal("same key did not recover unsent transaction", e)
	}
}
func TestLongImitationFreezesOwnerMemosAndStylesBeforeChunks(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := NewService(st, NewMemoryRuntime(), nil)
	p, e := st.CreateProject(ctx, "alice", CreateProjectInput{Title: "资料", Genre: "other", ModelID: "fake-only"})
	if e != nil {
		t.Fatal(e)
	}
	note, e := st.CreateNote(ctx, "alice", p.ID, ProjectNoteInput{Title: "备忘", Content: "仅写全新人物"})
	if e != nil {
		t.Fatal(e)
	}
	style, e := st.SaveStyleProfile(ctx, "alice", StyleProfile{Title: "短句", Content: "文风事实固定内容"})
	if e != nil {
		t.Fatal(e)
	}
	in := imitationInput("refs")
	in.MemoReferences = []ImitationMemoReference{{p.ID, note.ID}}
	in.KnowledgeBookIDs = []string{style.ID}
	out, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	if out.Creation.Input.ReferenceText != in.ReferenceText || !strings.Contains(out.Creation.EffectiveReferenceText, "--- 备忘录: 备忘 ---\n\n仅写全新人物") || out.Creation.Statistics.AvgParagraphLength != 120 {
		t.Fatal("reference/memo/stat separation lost")
	}
	_, _ = st.UpdateNote(ctx, "alice", p.ID, note.ID, ProjectNoteInput{Title: "后改", Content: "新内容"})
	_, _ = st.SaveStyleProfile(ctx, "alice", StyleProfile{ID: style.ID, Title: "后改", Content: "变化"})
	again, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil || again.Creation.Memos[0].Content != "仅写全新人物" || again.Creation.Styles[0].Profile.Content != "文风事实固定内容" {
		t.Fatal("replay re-resolved external references", e)
	}
	bad := in
	bad.IdempotencyKey = "foreign"
	if _, e = s.StartLongImitation(ctx, "bob", bad); e == nil {
		t.Fatal("cross-account memo/style allowed")
	}
	bad = in
	bad.IdempotencyKey = "short"
	bad.ReferenceText = strings.Repeat("甲", 99)
	if _, e = s.StartLongImitation(ctx, "alice", bad); e == nil {
		t.Fatal("memo counted toward original reference threshold")
	}
}
func TestLongImitationUTF16ChunksKeepAllTextAndMarkSurrogatePolicy(t *testing.T) {
	source := strings.Repeat("甲", 49999) + "😀" + strings.Repeat("乙", 50000)
	chunks := imitationChunks(source)
	joined := ""
	for _, chunk := range chunks {
		var text string
		if e := json.Unmarshal(chunk, &text); e != nil {
			t.Fatal(e)
		}
		if len(utf16.Encode([]rune(text))) > 50000 || strings.ContainsRune(text, '\ufffd') {
			t.Fatal("chunk cut surrogate or exceeded limit")
		}
		joined += text
	}
	if joined != source || len(chunks) != 3 {
		t.Fatal("reference silently truncated")
	}
	v := ImitationStyleStatistics("甲😀。甲😁。甲😀。", "")
	if len(v.TopPatterns) != 1 {
		t.Fatal("UTF16 prefix count not preserved")
	}
	var pattern string
	_ = json.Unmarshal(v.TopPatterns[0], &pattern)
	if pattern != "甲�...(3次)" {
		t.Fatal("isolated prefix display policy wrong", pattern)
	}
}
func TestLongImitationParseIsFreeAndCreatesNoBook(t *testing.T) {
	st := NewMemoryStore()
	s := NewService(st, nil, nil)
	chapters, e := s.ParseLongImitationReference(context.Background(), "alice", "sample.txt", []byte("第1章 开始\n甲😀\n第2章 继续\n乙"))
	if e != nil || len(chapters) != 2 || !strings.HasPrefix(chapters[0].Content, "第1章") {
		t.Fatal("chapter parsing contract", e)
	}
	if _, e = s.ParseLongImitationReference(context.Background(), "alice", "bad.docx", []byte("broken")); e == nil {
		t.Fatal("broken DOCX accepted")
	}
	if _, e = s.ParseLongImitationReference(context.Background(), "alice", "empty.txt", nil); e == nil {
		t.Fatal("empty reference accepted")
	}
	projects, _ := st.ListProjects(context.Background(), "alice")
	if len(projects) != 0 {
		t.Fatal("parse created hidden project")
	}
}
func TestLongImitationSummaryOmitsReferenceAndFrozenBodies(t *testing.T) {
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), nil)
	out, e := s.StartLongImitation(context.Background(), "alice", imitationInput("summary"))
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(LongImitationSummary(out.Creation))
	if strings.Contains(string(raw), out.Creation.Input.ReferenceText) || strings.Contains(string(raw), "referenceChunks") || strings.Contains(string(raw), "effectiveReferenceText") {
		t.Fatal("summary leaked full reference")
	}
	if !strings.Contains(string(raw), "50000_utf16_surrogate_safe") {
		t.Fatal("chunk policy not disclosed")
	}
}
func TestLongImitationLaterManualAndBatchUseStoredConfigAndStopAtChapterLimit(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := NewService(st, NewMemoryRuntime(), nil)
	out, e := s.StartLongImitation(ctx, "alice", imitationInput("cap"))
	if e != nil {
		t.Fatal(e)
	}
	selected, e := s.prepareGenerationInput(ctx, "alice", out.Project.ID, out.Session.ID, AutomationInput{})
	if e != nil || selected.ModeInputs["referenceText"] != out.Creation.EffectiveReferenceText {
		t.Fatal("later automatic/manual run lost frozen reference", e)
	}
	batch, e := s.CreateBatch(ctx, "alice", out.Project.ID, BatchInput{SessionID: out.Session.ID, ModelID: "fake-only", TotalChapters: 2, Prompt: "继续"})
	if e != nil || batch.ModeInputs["referenceText"] != out.Creation.EffectiveReferenceText {
		t.Fatal("batch did not inherit reference", e)
	}
	st.mu.Lock()
	st.chapterBundles[memoryKey("alice", out.Project.ID, "last")] = ChapterBundle{ProjectID: out.Project.ID, Chapter: Chapter{ID: "last", ProjectID: out.Project.ID, Number: 20}}
	st.mu.Unlock()
	_, e = s.SendMessage(ctx, "alice", SendMessageInput{ProjectID: out.Project.ID, SessionID: out.Session.ID, IdempotencyKey: "after-cap", ModelID: "fake-only", Content: "继续"})
	if !IsCode(e, ErrInvalidTransition) {
		t.Fatal("manual chapter ceiling not enforced", e)
	}
	runs, _ := st.ListRuns(ctx, "alice", out.Project.ID)
	if len(runs) != 1 {
		t.Fatal("cap created another paid run")
	}
}
