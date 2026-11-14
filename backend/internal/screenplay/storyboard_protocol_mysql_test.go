package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"qiantie/backend/internal/prompt"
	"reflect"
	"testing"
	"time"
)

func TestStoryboardProtocolMySQLBothSavePathsAndReload(t *testing.T) {
	for _, revision := range []bool{false, true} {
		t.Run(map[bool]string{false: "generation", true: "manual"}[revision], func(t *testing.T) {
			db, m, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			shot := StoryboardShot{ID: "s1", Seq: 1, Duration: 3, ImagePrompt: "图片", VoiceActor: "旁白", Emotion: "平静", Line: "你好", SourceElements: json.RawMessage(`{"人物":["甲"]}`), SourceSceneIndex: 2, SourceSceneLocalOrder: 3, ProtocolFields: json.RawMessage(`{"custom":{"lens":35},"scene_id":"scene-A"}`)}
			board := Storyboard{ID: "b1", ProjectID: "p1", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7, Shots: []StoryboardShot{shot}}
			protocol, _ := marshalStoryboardProtocol(shot)
			m.ExpectBegin()
			m.ExpectQuery("SELECT id, revision FROM screenplay_storyboards").WithArgs("p1", "alice", 1, string(EngineXiaoluoV7)).WillReturnError(sql.ErrNoRows)
			m.ExpectExec("INSERT INTO screenplay_storyboards").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectExec("DELETE FROM screenplay_storyboard_shots").WithArgs("b1", "p1", "alice").WillReturnResult(sqlmock.NewResult(0, 0))
			m.ExpectExec("INSERT INTO screenplay_storyboard_shots").WithArgs("s1", "b1", "p1", "alice", 1, 0, float64(3), "", "", "", "", protocol, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectCommit()
			store := NewMySQLStore(db)
			if revision {
				_, err = store.SaveStoryboardRevision(context.Background(), board, 0)
			} else {
				_, err = store.SaveStoryboard(context.Background(), board)
			}
			if err != nil {
				t.Fatal(err)
			}
			now := time.Now()
			m.ExpectQuery("SELECT id, project_id, owner_username, episode_no, engine, status").WithArgs("p1", "alice", 1, string(EngineXiaoluoV7)).WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "owner_username", "episode_no", "engine", "status", "source_sha256", "shot_count", "revision", "created_at", "updated_at"}).AddRow("b1", "p1", "alice", 1, string(EngineXiaoluoV7), "done", "", 1, 1, now, now))
			m.ExpectQuery("SELECT id, storyboard_id, project_id, owner_username, seq").WithArgs("b1", "p1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "storyboard_id", "project_id", "owner_username", "seq", "scene_id", "duration_sec", "characters_text", "scene_name", "props_text", "video_desc", "protocol_json"}).AddRow("s1", "b1", "p1", "alice", 1, 0, 3, "", "", "", "", protocol))
			saved, err := store.GetStoryboard(context.Background(), "alice", "p1", 1, EngineXiaoluoV7)
			if err != nil {
				t.Fatal(err)
			}
			got := saved.Shots[0]
			got.StoryboardID = ""
			got.ProjectID = ""
			got.OwnerUsername = ""
			if !reflect.DeepEqual(got, shot) {
				t.Fatalf("protocol changed: %#v", got)
			}
			if err = m.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestStoryboardEditorPreservesHiddenProtocolAndExplicitClear(t *testing.T) {
	original := StoryboardShot{ImagePrompt: "图", Line: "台词", Emotion: "开心", VoiceActor: "甲", SourceElements: json.RawMessage(`["甲"]`), SourceSceneIndex: 1, SourceSceneLocalOrder: 2, ProtocolFields: json.RawMessage(`{"lens":35}`)}
	updated := StoryboardShot{VideoDesc: "编辑后的画面"}
	preserveStoryboardProtocol(&updated, original)
	if updated.ImagePrompt != "图" || updated.Line != "台词" || string(updated.ProtocolFields) != `{"lens":35}` || updated.SourceSceneLocalOrder != 2 {
		t.Fatal(updated)
	}
	updated = StoryboardShot{EditedProtocolFields: []string{"line", "image_prompt"}}
	preserveStoryboardProtocol(&updated, original)
	if updated.Line != "" || updated.ImagePrompt != "" || updated.VoiceActor != "甲" {
		t.Fatal(updated)
	}
	if err := validateStoryboardShotInput(StoryboardShot{Duration: 3, ImagePrompt: "图"}); err != nil {
		t.Fatal(err)
	}
}

func TestStoryboardSplitMergeKeepsBothProtocolSources(t *testing.T) {
	store, board := seedStoryboardEditorTest(t)
	board.Shots[0].ImagePrompt = "第一图"
	board.Shots[0].ProtocolFields = json.RawMessage(`{"camera":"wide"}`)
	board.Shots[1].Line = "第二段台词"
	board.Shots[1].ProtocolFields = json.RawMessage(`{"camera":"close"}`)
	if _, err := store.SaveStoryboard(context.Background(), board); err != nil {
		t.Fatal(err)
	}
	split, err := SplitStoryboardShot(context.Background(), store, "alice", "comic-1", 1, EngineXiaoluoV7, "sb-shot-1")
	if err != nil {
		t.Fatal(err)
	}
	if split.Shots[1].ImagePrompt != "第一图" || string(split.Shots[1].ProtocolFields) != `{"camera":"wide"}` {
		t.Fatal(split.Shots)
	}
	merged, err := MergeStoryboardShotWithNext(context.Background(), store, "alice", "comic-1", 1, EngineXiaoluoV7, split.Shots[1].ID)
	if err != nil {
		t.Fatal(err)
	}
	var archive struct {
		Shots []StoryboardShot `json:"merged_shots"`
	}
	if err = json.Unmarshal(merged.Shots[1].ProtocolFields, &archive); err != nil {
		t.Fatal(err)
	}
	if len(archive.Shots) != 2 || archive.Shots[0].ImagePrompt != "第一图" || archive.Shots[1].Line != "第二段台词" || merged.Shots[1].Line != "第二段台词" {
		t.Fatal(archive)
	}
}

func TestStoryboardEditorAllowsUntimedImageProtocolOnly(t *testing.T) {
	for _, test := range []struct {
		shot StoryboardShot
		ok   bool
	}{
		{StoryboardShot{ImagePrompt: "图"}, true},
		{StoryboardShot{ImagePrompt: "图", Duration: -1}, false},
		{StoryboardShot{ImagePrompt: "图", VideoDesc: "视频"}, false},
		{StoryboardShot{VideoDesc: "视频"}, false},
	} {
		err := validateStoryboardShotInput(test.shot)
		if (err == nil) != test.ok {
			t.Fatalf("shot=%+v error=%v", test.shot, err)
		}
	}
}

func TestStoryboardBulkUpdatePreservesProtocol(t *testing.T) {
	ctx := context.Background()
	store, board := seedStoryboardEditorTest(t)
	if _, err := store.CreateProject(ctx, Project{ID: "comic-1", OwnerUsername: "alice"}); err != nil {
		t.Fatal(err)
	}
	board.Shots[0].ImagePrompt = "原图"
	board.Shots[0].SourceElements = json.RawMessage(`"L1-L3"`)
	board.Shots[0].ProtocolFields = json.RawMessage(`{"camera":"wide"}`)
	board, err := store.SaveStoryboard(ctx, board)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := EditStoryboardBatch(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, StoryboardBatchEdit{ExpectedRevision: board.Revision, Operation: "update", Shots: []StoryboardShot{{ID: board.Shots[0].ID, Duration: 3, VideoDesc: "新画面"}}})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Shots[0].ImagePrompt != "原图" || string(updated.Shots[0].SourceElements) != `"L1-L3"` || string(updated.Shots[0].ProtocolFields) != `{"camera":"wide"}` {
		t.Fatal(updated.Shots[0])
	}
}

func TestRecoveredUntimedVideoCanBeEditedOnlyWithCurrentContract(t *testing.T) {
	selected, err := prompt.ResolveCreativePrompt("luoshui_storyboard_137", "storyboard")
	if err != nil {
		t.Fatal(err)
	}
	shots, err := parseStoryboardJSON(`{"storyboard":[{"id":1,"scene":"室内","characters":["甲"],"image_prompt":"甲抬头","source_elements":["L1"],"voice_actor":"甲","emotion":"平静","line":"你好","video_desc":"甲抬头说你好"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	json.Unmarshal(shots[0].ProtocolFields, &fields)
	fields["_cm_contract"] = map[string]any{"prompt_id": selected.ID, "prompt_digest": selected.Digest, "output_protocol": selected.OutputProtocol, "parser_version": selected.ParserVersion}
	shots[0].ProtocolFields, _ = json.Marshal(fields)
	// Video-only edits retain proof from the saved generated row even when the
	// client sends only the visible fields.
	store, board := seedStoryboardEditorTest(t)
	shots[0].ID = board.Shots[0].ID
	board.Shots[0] = shots[0]
	if _, err = store.SaveStoryboard(context.Background(), board); err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateStoryboardShot(context.Background(), store, "alice", "comic-1", 1, EngineXiaoluoV7, shots[0].ID, StoryboardShot{VideoDesc: "修改后的视频词"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Shots[0].Duration != 0 || updated.Shots[0].VideoDesc != "修改后的视频词" || !storyboardProtocolOmitsDuration(updated.Shots[0].ProtocolFields) {
		t.Fatal(updated.Shots[0])
	}
	invalid := shots[0]
	invalid.ImagePrompt = ""
	invalid.ProtocolFields = nil
	if validateStoryboardShotInput(invalid) == nil {
		t.Fatal("unproven untimed video accepted")
	}
	fields["_cm_contract"].(map[string]any)["prompt_digest"] = "outdated"
	invalid.ProtocolFields, _ = json.Marshal(fields)
	if validateStoryboardShotInput(invalid) == nil {
		t.Fatal("outdated catalog accepted")
	}
}
