package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"qiantie/backend/internal/prompt"
	"testing"
	"time"
)

func storyboard139Example(imageMode string) string {
	row := map[string]any{"id": 1, "source_elements": []string{"1-1-001", "1-1-002"}, "characters": "裴应臣,兄弟A,兄弟B,林鸢", "scene": "教室走廊", "voice_actor": "旁白,兄弟A", "emotion": "平静,带着调侃笑意", "line": "他身后的两个兄弟，有些惊讶地看向我们|裴哥，原来你跟转校生认识啊", "video_desc": `【角色标签】@[裴应臣],这个人是裴应臣
【场景人物】裴应臣、兄弟A、兄弟B、林鸢
【画面描述】走廊上，两人好奇地打量林鸢。
【配音脚本】[兄弟A]：裴哥，原来你跟转校生认识啊。`}
	if imageMode != "omitted" {
		if imageMode == "empty" {
			row["image_prompt"] = ""
		} else {
			row["image_prompt"] = "真正提供的静帧图片词"
		}
	}
	raw, _ := json.Marshal(map[string]any{"storyboard": []any{row}})
	return string(raw)
}
func TestStoryboard139FinalExampleGeneratesSavesReloadsWithoutInventedImage(t *testing.T) {
	for _, mode := range []string{"omitted", "empty", "provided"} {
		t.Run(mode, func(t *testing.T) {
			s, store, g := newStoryboardService(t, PermissionBypass)
			g.responses = []string{storyboard139Example(mode)}
			req := generationTestRequest("139-optional-" + mode)
			req.PromptID = "139"
			out, err := s.Generate(context.Background(), req)
			if err != nil {
				t.Fatalf("139 final example with %s image rejected: %v", mode, err)
			}
			if len(out.Storyboards) != 1 || g.callCount() != 1 {
				t.Fatal(out, g.callCount())
			}
			saved, err := store.GetStoryboard(context.Background(), "alice", "p1", 1, StoryboardEngine("prompt_139"))
			if err != nil {
				t.Fatal(err)
			}
			shot := saved.Shots[0]
			want := ""
			if mode == "provided" {
				want = "真正提供的静帧图片词"
			}
			if shot.ImagePrompt != want || shot.VideoDesc == "" || shot.VoiceActor != "旁白,兄弟A" || shot.Line == "" || string(shot.SourceElements) != `["1-1-001","1-1-002"]` || shot.Duration != 15 {
				t.Fatalf("fields changed: %+v", shot)
			}
			var fields map[string]any
			json.Unmarshal(shot.ProtocolFields, &fields)
			if mode == "omitted" {
				if _, exists := fields["image_prompt"]; exists {
					t.Fatal("source omitted image was synthesized")
				}
			}
			contract := fields["_cm_contract"].(map[string]any)
			if contract["parser_version"] != "cm_storyboard_139_optional_image_v2" {
				t.Fatal(contract)
			}
			replay, err := s.Generate(context.Background(), req)
			if err != nil || len(replay.Storyboards) != 1 || g.callCount() != 1 {
				t.Fatal("replay resubmitted", fmt.Sprint(err), g.callCount())
			}
		})
	}
}

func TestStoryboard139MySQLSaveReloadKeepsAbsentImageAndSourceFields(t *testing.T) {
	raw := storyboard139Example("omitted")
	shots, err := parseStoryboardJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	selected, _ := prompt.ResolveCreativePrompt("139", "storyboard")
	if err = validateSelectedStoryboardProtocol(selected, raw, shots); err != nil {
		t.Fatal(err)
	}
	shot := shots[0]
	shot.ID = "shot139"
	shot.Duration = 15
	var fields map[string]any
	json.Unmarshal(shot.ProtocolFields, &fields)
	fields["_cm_contract"] = map[string]any{"parser_version": selected.ParserVersion, "prompt_digest": selected.Digest}
	shot.ProtocolFields, _ = json.Marshal(fields)
	board := Storyboard{ID: "board139", ProjectID: "p1", OwnerUsername: "alice", EpisodeNo: 1, Engine: StoryboardEngine("prompt_139"), Shots: []StoryboardShot{shot}}
	protocol, _ := marshalStoryboardProtocol(shot)
	db, m, _ := sqlmock.New()
	defer db.Close()
	m.ExpectBegin()
	m.ExpectQuery("SELECT id, revision FROM screenplay_storyboards").WillReturnError(sql.ErrNoRows)
	m.ExpectExec("INSERT INTO screenplay_storyboards").WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectExec("DELETE FROM screenplay_storyboard_shots").WillReturnResult(sqlmock.NewResult(0, 0))
	m.ExpectExec("INSERT INTO screenplay_storyboard_shots").WithArgs("shot139", "board139", "p1", "alice", shot.Seq, 0, float64(15), shot.Characters, shot.Scene, "", shot.VideoDesc, protocol, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	m.ExpectCommit()
	store := NewMySQLStore(db)
	if _, err = store.SaveStoryboard(context.Background(), board); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	m.ExpectQuery("SELECT id, project_id, owner_username, episode_no, engine, status").WithArgs("p1", "alice", 1, "prompt_139").WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "owner_username", "episode_no", "engine", "status", "source_sha256", "shot_count", "revision", "created_at", "updated_at"}).AddRow("board139", "p1", "alice", 1, "prompt_139", "done", "", 1, 1, now, now))
	m.ExpectQuery("SELECT id, storyboard_id, project_id, owner_username, seq").WithArgs("board139", "p1", "alice").WillReturnRows(sqlmock.NewRows([]string{"id", "storyboard_id", "project_id", "owner_username", "seq", "scene_id", "duration_sec", "characters_text", "scene_name", "props_text", "video_desc", "protocol_json"}).AddRow("shot139", "board139", "p1", "alice", 1, 0, 15, shot.Characters, shot.Scene, "", shot.VideoDesc, protocol))
	loaded, err := store.GetStoryboard(context.Background(), "alice", "p1", 1, "prompt_139")
	if err != nil {
		t.Fatal(err)
	}
	got := loaded.Shots[0]
	if got.ImagePrompt != "" || got.VideoDesc != shot.VideoDesc || got.VoiceActor != shot.VoiceActor || got.Line != shot.Line || string(got.SourceElements) != string(shot.SourceElements) {
		t.Fatal(got)
	}
	var original map[string]any
	json.Unmarshal(got.ProtocolFields, &original)
	if _, exists := original["image_prompt"]; exists {
		t.Fatal("SQL synthesized source image")
	}
	if err = m.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestStoryboard139InvalidOrCancelledResultKeepsOldDraft(t *testing.T) {
	for _, mode := range []string{"invalid-json", "null-image", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			s, store, g := newStoryboardService(t, PermissionBypass)
			old, err := store.SaveStoryboard(context.Background(), Storyboard{ProjectID: "p1", OwnerUsername: "alice", EpisodeNo: 1, Engine: "prompt_139", Shots: []StoryboardShot{{Seq: 1, Duration: 15, VideoDesc: "旧稿不可丢", ImagePrompt: "旧图片词", Line: "旧台词"}}})
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if mode == "cancel" {
				run, cancel := context.WithCancel(ctx)
				defer cancel()
				ctx = run
				s.Generator = storyboardGenerationTestFunc(func(context.Context, TextRequest) (TextResponse, error) {
					cancel()
					return TextResponse{Content: storyboard139Example("omitted")}, nil
				})
			} else if mode == "null-image" {
				var payload map[string]any
				json.Unmarshal([]byte(storyboard139Example("omitted")), &payload)
				payload["storyboard"].([]any)[0].(map[string]any)["image_prompt"] = nil
				raw, _ := json.Marshal(payload)
				g.responses = []string{string(raw)}
			} else {
				g.responses = []string{"not-json"}
			}
			req := generationTestRequest("139-protected-" + mode)
			req.PromptID = "139"
			if _, err = s.Generate(ctx, req); err == nil {
				t.Fatal("bad/cancelled result accepted")
			}
			got, err := store.GetStoryboard(context.Background(), "alice", "p1", 1, "prompt_139")
			if err != nil || got.Revision != old.Revision || got.Shots[0].VideoDesc != "旧稿不可丢" || got.Shots[0].ImagePrompt != "旧图片词" {
				t.Fatal(got, err)
			}
		})
	}
}
func TestStoryboard139OptionalImageDoesNotRelax137RequiredImage(t *testing.T) {
	selected, _ := prompt.ResolveCreativePrompt("137", "storyboard")
	report := "\n【分镜自检】\n场景：1-1 家\n剧本元素总数：1 -> 生成分镜数：1\n台词总数：1句 -> 遗漏：0\n连贯性检查：单组无相邻衔接\n映射完整性：角色1个/场景1个/物品1个"
	for _, mode := range []string{"omitted", "empty", "provided"} {
		raw := storyboard139Example(mode) + report
		shots, err := parseStoryboardJSON(raw)
		if err == nil {
			err = validateSelectedStoryboardProtocol(selected, raw, shots)
		}
		if mode == "provided" && err != nil {
			t.Fatal(err)
		}
		if mode != "provided" && err == nil {
			t.Fatal("137 required image loosened", mode)
		}
	}
}
func TestStoryboard139NumericAliasCannotBypassFrozenUnsupportedContract(t *testing.T) {
	ctx := context.Background()
	s, store, g := newStoryboardService(t, PermissionBypass)
	g.responses = []string{storyboard139Example("omitted")}
	project, err := store.GetProject(ctx, "alice", "p1")
	if err != nil {
		t.Fatal(err)
	}
	frozen, _ := prompt.ResolveCreativePrompt("139", "storyboard")
	frozen.Supported = false
	frozen.UnsupportedReason = "旧版未审核"
	frozen.ParserVersion = "storyboard_image_voice_v1"
	config := EffectiveCreativeConfig(project)
	config.StoryboardPromptID = frozen.ID
	config.PromptSnapshots = map[string]CreativePromptSnapshot{"storyboard": frozen}
	if _, err = store.SaveCreativeConfig(ctx, "alice", "p1", config, project.Revision); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"139", frozen.ID} {
		req := generationTestRequest("frozen-alias-" + id)
		req.PromptID = id
		if _, err := s.Generate(ctx, req); err == nil || g.callCount() != 0 {
			t.Fatalf("alias %q bypassed frozen unsupported contract: err=%v calls=%d", id, err, g.callCount())
		}
	}
}
