package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func chapterImportService(st Store) *Service { return NewService(st, NewMemoryRuntime(), nil) }

func TestChapterImportAtomicKeepsEmptyProjectProtection(t *testing.T) {
	st := NewMemoryStore()
	called := false
	e := st.Atomic(context.Background(), "alice", "", func(Store) error { called = true; return nil })
	if !IsCode(e, ErrProjectNotFound) || called {
		t.Fatalf("legacy Atomic accepted missing project: called=%v err=%v", called, e)
	}
}
func chapterImportInput(key string) ImportChaptersInput {
	return ImportChaptersInput{Title: "渡口", IdempotencyKey: key, Chapters: []ImportedChapterInput{{Title: "第二十章 归来", Content: " 雨停了。\n船靠岸。 "}, {Content: "故人等候。"}}}
}

func TestChapterImportFormalVersionsReferencesAndExports(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := chapterImportService(st)
	a, _ := s.SaveStyleProfile(ctx, "alice", StyleProfile{Title: "甲", Content: "短句"})
	b, _ := s.SaveStyleProfile(ctx, "alice", StyleProfile{Title: "乙", Content: "冷静"})
	in := chapterImportInput("one")
	in.KnowledgeBookIDs = []string{b.ID, a.ID}
	fakeCount := 999
	in.Chapters[0].WordCount = &fakeCount
	got, e := s.ImportChapters(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	if got.Book.ModelID != "" || got.Book.TargetChapterWords != 0 || got.Book.Genre != "other" || len(got.Chapters) != 2 || got.Chapters[0].Number != 1 || got.Chapters[1].Title != "第2章" || got.Chapters[0].Content != in.Chapters[0].Content || got.Chapters[0].WordCount == 999 {
		t.Fatalf("%+v", got)
	}
	for i, c := range got.Chapters {
		vs, e := s.ListChapterVersions(ctx, "alice", got.Book.ID, c.ID)
		if e != nil || len(vs) != 1 || vs[0].Version != 1 || vs[0].RunID != "" || vs[0].Content != c.Content {
			t.Fatalf("versions=%+v %v", vs, e)
		}
		var meta struct {
			Source           string
			SourceOrder      int
			KnowledgeBookIDs []string
			Styles           []importedStyleReference
		}
		if e = json.Unmarshal(vs[0].Snapshot, &meta); e != nil {
			t.Fatal(e)
		}
		if meta.Source != chapterImportKind || meta.SourceOrder != i+1 || len(meta.Styles) != 2 || meta.Styles[0].Profile.ID != b.ID || meta.Styles[1].Profile.Content != "短句" {
			t.Fatalf("meta=%+v", meta)
		}
	}
	state, e := s.GetContinuityState(ctx, "alice", got.Book.ID)
	if e != nil || state.DirtyFrom != 1 || state.Revision != 1 || len(state.Entries) > 0 {
		t.Fatalf("%+v %v", state, e)
	}
	if _, e = s.loadContinuity(ctx, "alice", got.Book.ID, got.Chapters); e == nil {
		t.Fatal("imported chapters accepted missing continuity")
	}
	for _, format := range []string{"txt", "md"} {
		doc, e := s.ExportProject(ctx, "alice", got.Book.ID, format)
		if e != nil || !strings.Contains(doc.Content, "船靠岸。") || !strings.Contains(doc.Content, "故人等候。") {
			t.Fatalf("export=%+v %v", doc, e)
		}
	}
	if len(st.runs) != 0 || len(st.executions) != 0 || len(st.materials) != 0 || len(st.foundations) != 0 {
		t.Fatal("import created generation artifacts")
	}
	if _, e = s.CreateProject(ctx, "alice", CreateProjectInput{Title: "旧入口", Genre: "other"}); !IsCode(e, ErrInputRequired) {
		t.Fatalf("legacy model check relaxed: %v", e)
	}
	se, e := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: got.Book.ID, Mode: "long_original"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.SendMessage(ctx, "alice", SendMessageInput{ProjectID: got.Book.ID, SessionID: se.ID, Content: "继续", IdempotencyKey: "no-model"}); !IsCode(e, ErrInputRequired) {
		t.Fatalf("missing model accepted: %v", e)
	}
	if _, e = s.SendMessage(ctx, "alice", SendMessageInput{ProjectID: got.Book.ID, SessionID: se.ID, Content: "继续", ModelID: "chosen", IdempotencyKey: "dirty"}); e == nil {
		t.Fatal("generation accepted dirty imported continuity")
	}
	if len(st.runs) != 0 {
		t.Fatal("unsafe generation created run")
	}
	// Existing sparse chapter numbers append after the maximum, never len+1.
	st.mu.Lock()
	key := memoryKey("alice", got.Book.ID, got.Chapters[1].ID)
	bundle := st.chapterBundles[key]
	bundle.Chapter.Number = 9
	st.chapterBundles[key] = bundle
	st.mu.Unlock()
	next := chapterImportInput("two")
	next.Title = ""
	next.BookID = got.Book.ID
	appended, e := s.ImportChapters(ctx, "alice", next)
	if e != nil || appended.Chapters[0].Number != 10 {
		t.Fatalf("append=%+v %v", appended, e)
	}
	old, _ := s.ListChapterVersions(ctx, "alice", got.Book.ID, got.Chapters[0].ID)
	if len(old) != 1 || old[0].Content != in.Chapters[0].Content {
		t.Fatal("history changed")
	}
}

func TestChapterImportReplayBeforeDeletedReferencesAndArchivedProject(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := chapterImportService(st)
	style, _ := s.SaveStyleProfile(ctx, "alice", StyleProfile{Content: "短句"})
	in := chapterImportInput("replay")
	in.KnowledgeBookIDs = []string{style.ID}
	first, e := s.ImportChapters(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	st.mu.Lock()
	delete(st.styles, memoryKey("alice", style.ID))
	p := st.projects[memoryKey("alice", first.Book.ID)]
	p.Status = "archived"
	st.projects[memoryKey("alice", p.ID)] = p
	st.mu.Unlock()
	replay, e := s.ImportChapters(ctx, "alice", in)
	if e != nil || fingerprint(first) != fingerprint(replay) {
		t.Fatalf("replay=%+v %v", replay, e)
	}
	in.Chapters[0].Content = "变更"
	if _, e = s.ImportChapters(ctx, "alice", in); !IsCode(e, ErrIdempotencyConflict) {
		t.Fatalf("conflict=%v", e)
	}
	fresh := chapterImportInput("fresh")
	fresh.Title = ""
	fresh.BookID = first.Book.ID
	if _, e = s.ImportChapters(ctx, "alice", fresh); e == nil {
		t.Fatal("archived import allowed")
	}
	if _, e = s.ImportChapters(ctx, "bob", fresh); !IsCode(e, ErrProjectNotFound) {
		t.Fatalf("foreign=%v", e)
	}
}

func TestChapterImportRejectsInvalidAndBusyWithoutShell(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	s := chapterImportService(st)
	for _, change := range []func(*ImportChaptersInput){func(i *ImportChaptersInput) { i.Title = "" }, func(i *ImportChaptersInput) { i.Chapters[1].Content = "  " }, func(i *ImportChaptersInput) { i.KnowledgeBookIDs = []string{"foreign"} }} {
		in := chapterImportInput(novelID("key"))
		change(&in)
		if _, e := s.ImportChapters(ctx, "alice", in); e == nil {
			t.Fatal("invalid accepted")
		}
	}
	if len(st.projects) != 0 || len(st.submissions) != 0 || len(st.chapterBundles) != 0 {
		t.Fatal("failed import left artifacts")
	}
	first, e := s.ImportChapters(ctx, "alice", chapterImportInput("ok"))
	if e != nil {
		t.Fatal(e)
	}
	in := chapterImportInput("busy")
	in.BookID = first.Book.ID
	in.Title = ""
	st.PutSlot(ctx, "alice", first.Book.ID, ProjectSlot{RunID: "lease", LeaseUntil: time.Now().Add(time.Hour)})
	if _, e = s.ImportChapters(ctx, "alice", in); !IsCode(e, ErrProjectLocked) {
		t.Fatalf("lease=%v", e)
	}
	st.PutSlot(ctx, "alice", first.Book.ID, ProjectSlot{})
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: first.Book.ID, Mode: "long_original"})
	st.CreateRun(ctx, "alice", Run{ProjectID: first.Book.ID, SessionID: se.ID, IdempotencyKey: "run", Status: RunPaused})
	if _, e = s.ImportChapters(ctx, "alice", in); !IsCode(e, ErrRunInProgress) {
		t.Fatalf("run=%v", e)
	}
}

type failingImportStore struct{ Store }

func (s failingImportStore) AtomicCreate(ctx context.Context, owner string, fn func(Store) error) error {
	return s.Store.AtomicCreate(ctx, owner, func(tx Store) error { return fn(failingImportTx{Store: tx}) })
}

type failingImportTx struct{ Store }

func (s failingImportTx) InsertImportedChapter(ctx context.Context, owner string, c Chapter, v ChapterVersion) error {
	if c.Number == 2 {
		return persistenceError(errors.New("second insert failed"))
	}
	return s.Store.InsertImportedChapter(ctx, owner, c, v)
}
func TestChapterImportMidWriteRollsBackNewBookAndReceipt(t *testing.T) {
	st := NewMemoryStore()
	s := chapterImportService(failingImportStore{Store: st})
	if _, e := s.ImportChapters(context.Background(), "alice", chapterImportInput("rollback")); !IsCode(e, ErrPersistenceFailed) {
		t.Fatalf("%v", e)
	}
	if len(st.projects) != 0 || len(st.chapterBundles) != 0 || len(st.chapterVersions) != 0 || len(st.submissions) != 0 || len(st.contentRecords) != 0 {
		t.Fatal("partial import committed")
	}
}
