package novelcreation

import (
	"context"
	"errors"
	"os"
	"qiantie/backend/internal/documenttext"
	"strings"
	"testing"
	"time"
)

type gatedDocumentParser struct{ entered, release chan struct{} }

func (p gatedDocumentParser) Parse(ctx context.Context, _ string, _ []byte) (documenttext.Result, error) {
	close(p.entered)
	select {
	case <-ctx.Done():
		return documenttext.Result{}, ctx.Err()
	case <-p.release:
		return documenttext.Result{Text: "旧尝试正文"}, nil
	}
}
func TestParseAttemptCannotResurrectDeletedDocument(t *testing.T) {
	ctx := context.Background()
	s, _, p := materialService(t)
	parser := gatedDocumentParser{make(chan struct{}), make(chan struct{})}
	s.ConfigureDocuments(&fakeDocumentObjects{objects: map[string][]byte{}}, parser)
	j, e := s.UploadDocument(ctx, "alice", p.ID, "a.txt", strings.NewReader("text"))
	if e != nil {
		t.Fatal(e)
	}
	done := make(chan error, 1)
	go func() { _, e := s.ParseDocument(ctx, "alice", p.ID, j.ID); done <- e }()
	<-parser.entered
	if e = s.DeleteDocument(ctx, "alice", p.ID, j.ID); e != nil {
		t.Fatal(e)
	}
	close(parser.release)
	if e = <-done; e == nil {
		t.Fatal("old parse resurrected deleted source")
	}
	if _, e = s.GetDocument(ctx, "alice", p.ID, j.ID); e == nil {
		t.Fatal("deleted became ready")
	}
}

type failDocumentSave struct {
	Store
	fail *bool
}

func (s failDocumentSave) Atomic(ctx context.Context, o, p string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, o, p, func(tx Store) error { return fn(failDocumentSave{tx, s.fail}) })
}
func (s failDocumentSave) PutContentRecord(ctx context.Context, r ContentRecord) error {
	if r.Status == "queued" && *s.fail {
		*s.fail = false
		return errors.New("database unavailable after upload")
	}
	return s.Store.PutContentRecord(ctx, r)
}
func TestUploadIntentRetainedOnDatabaseFailureAndRecover(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	objects := &fakeDocumentObjects{objects: map[string][]byte{}}
	s.ConfigureDocuments(objects, nil)
	fail := true
	s.store = failDocumentSave{st, &fail}
	j, e := s.UploadDocument(ctx, "alice", p.ID, "a.txt", strings.NewReader("资料正文"))
	if e == nil {
		t.Fatal("missing failure")
	}
	s.store = st
	persisted, e := s.GetDocument(ctx, "alice", p.ID, j.ID)
	if e != nil || persisted.Status != "uploading" || persisted.Object.Key == "" || objects.objects[persisted.Object.Key] == nil {
		t.Fatalf("lost orphan tracking %+v %v", persisted, e)
	}
	if _, e = s.RetryDocument(ctx, "alice", p.ID, j.ID); e != nil {
		t.Fatal(e)
	}
	ready, e := s.ParseDocument(ctx, "alice", p.ID, j.ID)
	if e != nil || ready.Status != "ready" {
		t.Fatal(e)
	}
	if e = s.PurgeDocumentObjects(ctx, "alice", p.ID); e != nil || len(objects.objects) != 0 {
		t.Fatal("object purge failed")
	}
}
func TestChinesePDFChaptersAndSelectionSnapshot(t *testing.T) {
	ctx := context.Background()
	s, _, p := materialService(t)
	s.ConfigureDocuments(&fakeDocumentObjects{objects: map[string][]byte{}}, nil)
	b, e := os.ReadFile("../documenttext/testdata/unicode.pdf")
	if e != nil {
		t.Fatal(e)
	}
	j, e := s.UploadDocument(ctx, "alice", p.ID, "中文.pdf", strings.NewReader(string(b)))
	if e != nil {
		t.Fatal(e)
	}
	j, e = s.ParseDocument(ctx, "alice", p.ID, j.ID)
	if e != nil || len(j.Chapters) != 2 || !strings.Contains(j.Chapters[0].Content, "你好世界") {
		t.Fatalf("chapters=%+v %v", j.Chapters, e)
	}
	selected, e := s.SelectDocument(ctx, "alice", p.ID, j.ID, DocumentSelection{j.Version, []string{j.Chapters[1].ID}})
	if e != nil {
		t.Fatal(e)
	}
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_imitation"})
	r, e := s.SendMessage(ctx, "alice", SendMessageInput{SessionID: se.ID, ModelID: "fake", Content: "仿写", IdempotencyKey: "selection", KnowledgeIDs: []string{selected.ID}})
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(string(r.PromptSnapshot), "selectedChapterIds") || strings.Contains(string(r.PromptSnapshot), "你好世界") {
		t.Fatal("whole document leaked into selected snapshot")
	}
	s.CancelRun(ctx, "alice", r.ID)
	s.DeleteDocument(ctx, "alice", p.ID, j.ID)
	if _, e = s.SendMessage(ctx, "alice", SendMessageInput{SessionID: se.ID, ModelID: "fake", Content: "new", IdempotencyKey: "deleted", KnowledgeIDs: []string{selected.ID}}); e == nil {
		t.Fatal("deleted reference reused")
	}
}
func TestExpiredParseAttemptIsDispatchedAfterRestart(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	s.ConfigureDocuments(&fakeDocumentObjects{objects: map[string][]byte{}}, nil)
	j, e := s.UploadDocument(ctx, "alice", p.ID, "a.txt", strings.NewReader("正文"))
	if e != nil {
		t.Fatal(e)
	}
	j.Status = "parsing"
	j.Attempt = 1
	st.Atomic(ctx, "alice", p.ID, func(tx Store) error { return storeDocument(ctx, tx, "alice", j, time.Now().Add(-time.Second)) })
	rows, e := st.ListParseDispatch(ctx, 8)
	if e != nil || len(rows) != 1 {
		t.Fatal(e)
	}
	ready, e := s.ParseDocument(ctx, "alice", p.ID, j.ID)
	if e != nil || ready.Attempt != 2 || ready.Status != "ready" {
		t.Fatalf("%+v %v", ready, e)
	}
}
