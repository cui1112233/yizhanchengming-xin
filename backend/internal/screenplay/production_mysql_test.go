package screenplay

import (
	"context"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
	"time"
)

func TestProductionMySQLDurableOwnerAndVersionRead(t *testing.T) {
	db, mock, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	s := NewMySQLStore(db)
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT owner_username,project_id,status,error,units,updated_at,rebuild,run_at FROM screenplay_production_jobs WHERE owner_username=? AND project_id=?`)).WithArgs("a", "p").WillReturnRows(sqlmock.NewRows([]string{"owner", "project", "status", "error", "units", "updated", "rebuild", "run_at"}).AddRow("a", "p", "running", "", 75, now, true, nil))
	j, e := s.GetProductionJob(context.Background(), "a", "p")
	if e != nil || j.Units != 75 || !j.Rebuild {
		t.Fatalf("%+v %v", j, e)
	}
	task := ShotTask{Owner: "a", ProjectID: "p", StoryboardID: "b", Seq: 3, RequestHash: "v2", TaskID: "provider-task"}
	raw, _ := json.Marshal(task)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT payload FROM screenplay_production_shots WHERE owner_username=? AND project_id=? AND storyboard_id=? AND shot_seq=? AND request_hash=?`)).WithArgs("a", "p", "b", 3, "v2").WillReturnRows(sqlmock.NewRows([]string{"payload"}).AddRow(raw))
	got, e := s.GetShotTaskVersion(context.Background(), "a", "p", "b", 3, "v2")
	if e != nil || got.TaskID != "provider-task" {
		t.Fatalf("%+v %v", got, e)
	}
	if e = mock.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
