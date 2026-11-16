package novelcreation

import (
	"context"
	"strings"
	"testing"
)

func TestSavedExportEmptyAndTOSOwnership(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	for _, f := range []string{"txt", "md"} {
		if _, e := s.ExportProject(ctx, "alice", p.ID, f); e == nil || !strings.Contains(e.Error(), "暂无可导出正文") {
			t.Fatalf("empty %s: %v", f, e)
		}
	}
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
	r, _ := st.CreateRun(ctx, "alice", Run{ProjectID: p.ID, SessionID: se.ID, IdempotencyKey: "saved", Status: RunQueued})
	b := serviceTestBundle(RunContext{Run: r, Project: p, Session: se, ChapterNumber: 1}, RunSucceeded)
	if e := st.SaveChapterBundle(ctx, "alice", b); e != nil {
		t.Fatal(e)
	}
	draft, _ := st.CreateRun(ctx, "alice", Run{ProjectID: p.ID, SessionID: se.ID, IdempotencyKey: "draft", Status: RunQueued})
	if _, e := s.GetShortResult(ctx, "alice", draft.ID); e == nil {
		t.Fatal("unsaved short leaked")
	}
	short, e := s.GetShortResult(ctx, "alice", r.ID)
	if e != nil || short.Content != b.Chapter.Content {
		t.Fatalf("%+v %v", short, e)
	}
	if _, e = s.PersistExport(ctx, "alice", p.ID, "txt"); e == nil {
		t.Fatal("TOS unavailable accepted")
	}
	objects := &fakeDocumentObjects{objects: map[string][]byte{}}
	s.ConfigureDocuments(objects, nil)
	receipt, e := s.PersistExport(ctx, "alice", p.ID, "txt")
	if e != nil || receipt.Status != "ready" || len(receipt.Sources) != 1 {
		t.Fatalf("%+v %v", receipt, e)
	}
	export, e := s.DownloadExport(ctx, "alice", p.ID, receipt.ID)
	if e != nil || !strings.Contains(export.Content, b.Chapter.Content) {
		t.Fatal(e)
	}
	if _, e = s.DownloadExport(ctx, "bob", p.ID, receipt.ID); e == nil {
		t.Fatal("foreign export read")
	}
	if e = s.PurgeDocumentObjects(ctx, "alice", p.ID); e != nil || len(objects.objects) != 0 {
		t.Fatal("export objects not purged")
	}
}
