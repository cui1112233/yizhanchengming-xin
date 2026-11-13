package screenplay

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

// MySQL JSON columns reject []byte bound as a binary-character-set string.
type jsonTextBinding struct{}

func (jsonTextBinding) Match(value driver.Value) bool {
	text, ok := value.(string)
	return ok && json.Valid([]byte(text))
}
func TestScreenplayNullableJSONUsesText(t *testing.T) {
	if value := nullableBytes([]byte(`{"人物":"林晚"}`)); !(jsonTextBinding{}).Match(value) {
		t.Fatalf("JSON parameter must be text, got %T", value)
	}
	if nullableBytes(nil) != nil {
		t.Fatal("nullable JSON must preserve SQL NULL")
	}
}
func TestScreenplayMySQLJSONParametersAreText(t *testing.T) {
	t.Run("media", func(t *testing.T) {
		db, m, e := sqlmock.New()
		if e != nil {
			t.Fatal(e)
		}
		defer db.Close()
		store := NewMySQLStore(db)
		m.ExpectBegin()
		m.ExpectExec("INSERT IGNORE INTO screenplay_media_workbench").WithArgs("owner", "project").WillReturnResult(sqlmock.NewResult(1, 1))
		m.ExpectQuery("SELECT state_json FROM screenplay_media_workbench").WithArgs("owner", "project").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(`{}`))
		m.ExpectExec("UPDATE screenplay_media_workbench SET state_json").WithArgs(jsonTextBinding{}, "owner", "project").WillReturnResult(sqlmock.NewResult(0, 1))
		m.ExpectCommit()
		_, e = store.MutateMedia(context.Background(), "owner", "project", func(state *MediaState) error { state.Voices["narrator"] = "中文音色"; return nil })
		if e != nil {
			t.Fatal(e)
		}
		if e = m.ExpectationsWereMet(); e != nil {
			t.Fatal(e)
		}
	})
	t.Run("shot", func(t *testing.T) {
		db, m, e := sqlmock.New()
		if e != nil {
			t.Fatal(e)
		}
		defer db.Close()
		store := NewMySQLStore(db)
		now := time.Now()
		m.ExpectExec("INSERT INTO screenplay_production_shots").WithArgs("owner", "project", "board", 1, "hash", jsonTextBinding{}, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
		e = store.SaveShotTask(context.Background(), ShotTask{Owner: "owner", ProjectID: "project", StoryboardID: "board", Seq: 1, RequestHash: "hash", UpdatedAt: now})
		if e != nil {
			t.Fatal(e)
		}
		if e = m.ExpectationsWereMet(); e != nil {
			t.Fatal(e)
		}
	})
}
