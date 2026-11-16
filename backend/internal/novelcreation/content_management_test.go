package novelcreation

import (
	"context"
	"encoding/json"
	"testing"
)

func TestManualChapterVersionsAndEarlierDirty(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	se, _ := s.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: p.ID, Mode: "long_original"})
	for n := 1; n <= 2; n++ {
		r, _ := st.CreateRun(ctx, "alice", Run{ID: novelID("run"), ProjectID: p.ID, SessionID: se.ID, IdempotencyKey: novelID("key"), Status: RunQueued})
		b := serviceTestBundle(RunContext{Run: r, Project: p, Session: se, ChapterNumber: n}, RunSucceeded)
		b.Chapter.ID = novelID("chapter")
		b.Chapter.Number = n
		b.ChapterVersion.ChapterID = b.Chapter.ID
		b.ChapterVersion.Version = 1
		b.Chapter.Version = 1
		if e := st.SaveChapterBundle(ctx, "alice", b); e != nil {
			t.Fatal(e)
		}
	}
	ch, _ := st.ListChapters(ctx, "alice", p.ID)
	edited, e := s.EditChapter(ctx, "alice", p.ID, ch[0].ID, ChapterEdit{ExpectedVersion: 1, Title: "改题", Content: "新的第一章"})
	if e != nil || edited.Version != 2 {
		t.Fatalf("%+v %v", edited, e)
	}
	if _, e = s.EditChapter(ctx, "alice", p.ID, ch[0].ID, ChapterEdit{ExpectedVersion: 1, Content: "旧窗口"}); e == nil {
		t.Fatal("stale edit")
	}
	versions, e := s.ListChapterVersions(ctx, "alice", p.ID, ch[0].ID)
	if e != nil || len(versions) != 2 {
		t.Fatalf("%+v %v", versions, e)
	}
	current, _ := st.ListChapters(ctx, "alice", p.ID)
	if _, e = s.loadContinuity(ctx, "alice", p.ID, current); e == nil {
		t.Fatal("unchanged last chapter accepted stale earlier memory")
	}
	restored, e := s.RestoreChapter(ctx, "alice", p.ID, ch[0].ID, 2, 1)
	if e != nil || restored.Version != 3 || restored.Content != ch[0].Content {
		t.Fatalf("%+v %v", restored, e)
	}
	state, e := s.GetContinuityState(ctx, "alice", p.ID)
	if e != nil || state.DirtyFrom != 1 {
		t.Fatalf("%+v %v", state, e)
	}
	raw, _ := json.Marshal(state)
	if len(raw) == 0 {
		t.Fatal("no state")
	}
}
