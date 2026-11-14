package screenplay

import (
	"context"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func TestStoryboardPreferencesMySQLAtomicOwnerRevisionAndFreshRead(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	store := NewMySQLStore(db)
	m.ExpectBegin()
	m.ExpectExec("INSERT IGNORE INTO screenplay_storyboard_preferences").WithArgs("alice").WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectQuery("SELECT state_json FROM screenplay_storyboard_preferences.*FOR UPDATE").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(`{"revision":3,"favorite_ids":[],"recent_ids":[],"prefix_presets":[]}`))
	expected := StoryboardPreferences{Revision: 4, FavoriteIDs: []string{}, RecentIDs: []string{}, PrefixPresets: []StoryboardPrefixPreset{{ID: "p", Name: "名称", Content: "实际正文"}}, SelectedPrefixID: "p"}
	raw, _ := json.Marshal(expected)
	m.ExpectExec("UPDATE screenplay_storyboard_preferences SET state_json").WithArgs(string(raw), "alice").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	saved, e := store.MutateStoryboardPreferences(context.Background(), "alice", func(v *StoryboardPreferences) error {
		v.PrefixPresets = expected.PrefixPresets
		v.SelectedPrefixID = "p"
		return nil
	})
	if e != nil || saved.Revision != 4 {
		t.Fatal(saved, e)
	}
	m.ExpectQuery("SELECT state_json FROM screenplay_storyboard_preferences WHERE owner_username").WithArgs("alice").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(raw))
	fresh, e := NewMySQLStore(db).LoadStoryboardPreferences(context.Background(), "alice")
	if e != nil || fresh.SelectedPrefix() != "实际正文" || fresh.Revision != 4 {
		t.Fatal(fresh, e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
