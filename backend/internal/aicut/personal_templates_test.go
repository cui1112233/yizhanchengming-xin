package aicut

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func safePersonalLayout() PersonalTemplateLayout {
	return PersonalTemplateLayout{Version: 1, Slots: []PersonalTemplateSlot{{ID: "slot-1", Kind: "video", Label: "素材 1"}}, Tracks: []Track{{ID: "track-1", Kind: TrackVideo, Order: 0, Clips: []Clip{{ID: "clip-1", AssetID: "slot-1", Kind: ClipVideo, DurationMS: 1000, SourceDurationMS: 2000}}}}}
}
func TestPersonalTemplateMySQLScopesEveryOperation(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s := NewMySQLPersonalTemplateStore(db)
	ctx := context.Background()
	l := safePersonalLayout()
	raw, _ := json.Marshal(l)
	m.ExpectExec("INSERT INTO aicut_personal_templates").WithArgs(sqlmock.AnyArg(), "alice", "模板", string(raw), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	created, e := s.Create(ctx, "alice", "模板", l)
	if e != nil {
		t.Fatal(e)
	}
	m.ExpectQuery("SELECT id,title,layout_json,created_at FROM aicut_personal_templates WHERE owner_username=\\?").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "layout_json", "created_at"}).AddRow(created.ID, "模板", raw, time.Now()))
	items, e := s.List(ctx, "alice")
	if e != nil || len(items) != 1 {
		t.Fatalf("list=%v %v", items, e)
	}
	m.ExpectQuery("SELECT id,title,layout_json,created_at FROM aicut_personal_templates WHERE owner_username=\\?").WithArgs("bob").WillReturnRows(sqlmock.NewRows([]string{"id", "title", "layout_json", "created_at"}))
	items, e = s.List(ctx, "bob")
	if e != nil || len(items) != 0 {
		t.Fatalf("foreign list=%v %v", items, e)
	}
	m.ExpectExec("DELETE FROM aicut_personal_templates WHERE owner_username=\\? AND id=\\?").WithArgs("bob", created.ID).WillReturnResult(sqlmock.NewResult(0, 0))
	if e = s.Delete(ctx, "bob", created.ID); !errors.Is(e, ErrPersonalTemplateNotFound) {
		t.Fatalf("foreign delete=%v", e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
func TestPersonalTemplateMemoryDoesNotAliasSavedLayout(t *testing.T) {
	s := NewMemoryPersonalTemplateStore()
	ctx := context.Background()
	l := safePersonalLayout()
	saved, e := s.Create(ctx, "alice", "模板", l)
	if e != nil {
		t.Fatal(e)
	}
	l.Tracks[0].Clips[0].AssetID = "private"
	saved.Layout.Tracks[0].Clips[0].AssetID = "private"
	items, _ := s.List(ctx, "alice")
	if items[0].Layout.Tracks[0].Clips[0].AssetID != "slot-1" {
		t.Fatal("saved layout mutated")
	}
	items[0].Layout.Tracks[0].Clips[0].AssetID = "private"
	items, _ = s.List(ctx, "alice")
	if items[0].Layout.Tracks[0].Clips[0].AssetID != "slot-1" {
		t.Fatal("list aliased storage")
	}
}
