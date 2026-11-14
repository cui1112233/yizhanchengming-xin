package screenplay

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

type generationSuccessJSON struct{ captured *string }

func (m generationSuccessJSON) Match(value driver.Value) bool {
	raw, ok := value.(string)
	if !ok {
		return false
	}
	var state StoryboardGenerationState
	if json.Unmarshal([]byte(raw), &state) != nil {
		return false
	}
	a := state.Attempts["stable-sql"]
	if a == nil || a.Episodes[1].Status != "succeeded" || a.Episodes[1].Storyboard == nil || a.Episodes[1].Storyboard.ID != "board-new" || len(a.Result.Storyboards) != 1 {
		return false
	}
	*m.captured = raw
	return true
}
func generationSQLState() string {
	v := StoryboardGenerationState{Revision: 1, Attempts: map[string]*StoryboardGenerationAttempt{"stable-sql": {RequestKey: "stable-sql", Status: "running", Engine: EngineXiaoluoV7, EpisodeNumbers: []int{1}, Episodes: map[int]*StoryboardGenerationEpisode{1: {EpisodeNo: 1, Status: "received", Calls: map[string]StoryboardGenerationCall{"storyboard": {Status: "received", RawResponse: storyboardJSON}}}}}}}
	raw, _ := json.Marshal(v)
	return string(raw)
}

func TestStoryboardGenerationMySQLSuccessCheckpointAndBoardCommitTogether(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	m.ExpectBegin()
	m.ExpectExec("INSERT IGNORE INTO screenplay_storyboard_generations").WithArgs("alice", "p1").WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectQuery("SELECT state_json FROM screenplay_storyboard_generations.*FOR UPDATE").WithArgs("alice", "p1").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(generationSQLState()))
	m.ExpectQuery("SELECT id, revision FROM screenplay_storyboards").WithArgs("p1", "alice", 1, string(EngineXiaoluoV7)).WillReturnRows(sqlmock.NewRows([]string{"id", "revision"}))
	m.ExpectExec("INSERT INTO screenplay_storyboards").WithArgs("board-new", "p1", "alice", 1, string(EngineXiaoluoV7), "done", "source-hash", 1, int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectExec("DELETE FROM screenplay_storyboard_shots").WithArgs("board-new", "p1", "alice").WillReturnResult(sqlmock.NewResult(0, 0))
	m.ExpectExec("INSERT INTO screenplay_storyboard_shots").WithArgs("shot-new", "board-new", "p1", "alice", 1, 1, float64(12), "甲", "家", "无", "实际镜头", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	captured := ""
	m.ExpectExec("UPDATE screenplay_storyboard_generations SET state_json").WithArgs(generationSuccessJSON{captured: &captured}, "alice", "p1").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	board := Storyboard{ID: "board-new", OwnerUsername: "alice", ProjectID: "p1", EpisodeNo: 1, Engine: EngineXiaoluoV7, SourceSHA256: "source-hash", Shots: []StoryboardShot{{ID: "shot-new", Seq: 1, SceneID: 1, Duration: 12, Characters: "甲", Scene: "家", Props: "无", VideoDesc: "实际镜头"}}}
	a, e := NewMySQLStore(db).CommitStoryboardGeneration(context.Background(), "alice", "p1", "stable-sql", 1, board, nil)
	if e != nil || a.Episodes[1].Status != "succeeded" {
		t.Fatal(a, e)
	}
	m.ExpectQuery("SELECT state_json FROM screenplay_storyboard_generations WHERE owner_username").WithArgs("alice", "p1").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(captured))
	fresh, e := NewMySQLStore(db).LoadStoryboardGenerations(context.Background(), "alice", "p1")
	if e != nil || fresh.Attempts["stable-sql"].Result.Storyboards[0].Shots[0].VideoDesc != "实际镜头" {
		t.Fatal(fresh, e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}

func TestStoryboardGenerationMySQLBoardFailureRollsBackWithoutSuccessCheckpoint(t *testing.T) {
	db, m, e := sqlmock.New()
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	m.ExpectBegin()
	m.ExpectExec("INSERT IGNORE INTO screenplay_storyboard_generations").WithArgs("alice", "p1").WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectQuery("SELECT state_json FROM screenplay_storyboard_generations.*FOR UPDATE").WithArgs("alice", "p1").WillReturnRows(sqlmock.NewRows([]string{"state_json"}).AddRow(generationSQLState()))
	m.ExpectQuery("SELECT id, revision FROM screenplay_storyboards").WithArgs("p1", "alice", 1, string(EngineXiaoluoV7)).WillReturnError(fmt.Errorf("database write unavailable"))
	m.ExpectRollback()
	_, e = NewMySQLStore(db).CommitStoryboardGeneration(context.Background(), "alice", "p1", "stable-sql", 1, Storyboard{ID: "b", OwnerUsername: "alice", ProjectID: "p1", EpisodeNo: 1, Engine: EngineXiaoluoV7}, nil)
	if e == nil || e == sql.ErrNoRows {
		t.Fatal(e)
	}
	if e = m.ExpectationsWereMet(); e != nil {
		t.Fatal(e)
	}
}
