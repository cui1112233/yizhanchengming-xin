package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"testing"
	"time"
)

func TestScreenplayNumberedFilesValidateSourceRevisionAndOwner(t *testing.T) {
	now := time.Now()
	store := screenplay.NewMemoryStore()
	_, err := store.CreateProject(context.Background(), screenplay.Project{ID: "files", OwnerUsername: "alice", Title: "文件验收", SourceText: "第一行\r\n\r\n第三行"})
	if err != nil {
		t.Fatal(err)
	}
	authors := &screenplay.AuthoringService{Store: store, Repository: screenplay.NewMemoryAuthoringRepository()}
	mux := http.NewServeMux()
	registerScreenplayFileRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, authors)
	base := "/api/comic-screenplays/projects/files"
	response := signedJSONRequest(t, mux, now, "alice", "GET", base+"/source-numbered", nil)
	var numbered struct {
		Hash  string                 `json:"source_sha256"`
		Lines []screenplaySourceLine `json:"lines"`
	}
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &numbered) != nil || len(numbered.Lines) != 3 || numbered.Lines[1].Text != "" {
		t.Fatalf("numbered %d %s", response.Code, response.Body.String())
	}
	input := map[string]any{"expected_revision": 0, "source_sha256": numbered.Hash, "storyboard": screenplay.Storyboard{Engine: screenplay.EngineXiaoluoV7, EpisodeNo: 1, Shots: []screenplay.StoryboardShot{{ID: "foreign", OwnerUsername: "bob", Duration: 3, VideoDesc: "〔原文 L1–L3〕推门"}}}}
	imported := signedJSONRequest(t, mux, now, "alice", "POST", base+"/storyboards/import", input)
	if imported.Code != 200 {
		t.Fatalf("import %d %s", imported.Code, imported.Body.String())
	}
	board, e := store.GetStoryboard(context.Background(), "alice", "files", 1, screenplay.EngineXiaoluoV7)
	if e != nil || board.Shots[0].OwnerUsername != "alice" || board.Shots[0].ID == "foreign" {
		t.Fatalf("ownership not normalized %+v %v", board, e)
	}
	stale := signedJSONRequest(t, mux, now, "alice", "POST", base+"/storyboards/import", input)
	if stale.Code != 409 {
		t.Fatalf("stale CAS %d", stale.Code)
	}
	input["expected_revision"] = board.Revision
	input["source_sha256"] = "old-source"
	stale = signedJSONRequest(t, mux, now, "alice", "POST", base+"/storyboards/import", input)
	if stale.Code != 409 {
		t.Fatalf("source drift %d", stale.Code)
	}
	forbidden := signedJSONRequest(t, mux, now, "bob", "GET", base+"/source-numbered", nil)
	if forbidden.Code < 400 {
		t.Fatalf("private source leaked %d", forbidden.Code)
	}
}
func TestScreenplayImportRechecksExactOwnerAfterGrantRevocation(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	memory := screenplay.NewMemoryStore()
	repo := screenplay.NewMemoryAuthoringRepository()
	store := &screenplayFileOwnerRaceStore{MemoryStore: memory}
	memory.CreateProject(ctx, screenplay.Project{ID: "same", OwnerUsername: "alice", SourceText: "原文"})
	authors := &screenplay.AuthoringService{Store: store, Repository: repo}
	state, _ := authors.Load(ctx, "alice", "same")
	state.Members = []screenplay.MemberGrant{{Username: "bob", Read: true, Edit: true}}
	state, _ = authors.Save(ctx, "alice", "same", state, 0)
	store.beforeLock = func() {
		state.Members = nil
		repo.Save(ctx, state, state.Revision)
		memory.CreateProject(ctx, screenplay.Project{ID: "same", OwnerUsername: "bob", SourceText: "自己的原文"})
	}
	mux := http.NewServeMux()
	registerScreenplayFileRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, authors)
	hash, _ := numberedScreenplaySource("原文")
	response := signedJSONRequest(t, mux, now, "bob", "POST", "/api/comic-screenplays/projects/same/storyboards/import", map[string]any{"expected_revision": 0, "source_sha256": hash, "storyboard": screenplay.Storyboard{Engine: screenplay.EngineXiaoluoV7, EpisodeNo: 1, Shots: []screenplay.StoryboardShot{{Duration: 3, VideoDesc: "不应写入已撤权的项目"}}}})
	if response.Code != 403 {
		t.Fatalf("owner changed but write proceeded: %d %s", response.Code, response.Body.String())
	}
	if _, e := memory.GetStoryboard(ctx, "alice", "same", 1, screenplay.EngineXiaoluoV7); !errors.Is(e, screenplay.ErrNotFound) {
		t.Fatal("foreign project was modified", e)
	}
}

type screenplayFileOwnerRaceStore struct {
	*screenplay.MemoryStore
	beforeLock func()
}

func (s *screenplayFileOwnerRaceStore) AcquireLock(ctx context.Context, name, token string, ttl time.Duration) error {
	if s.beforeLock != nil {
		fn := s.beforeLock
		s.beforeLock = nil
		fn()
	}
	return s.MemoryStore.AcquireLock(ctx, name, token, ttl)
}

func TestScreenplayFilesRecoveredProtocolRoundTripAndUntimedImage(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	store := screenplay.NewMemoryStore()
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "protocol-files", OwnerUsername: "alice", SourceText: "原文"}); err != nil {
		t.Fatal(err)
	}
	authors := &screenplay.AuthoringService{Store: store, Repository: screenplay.NewMemoryAuthoringRepository()}
	mux := http.NewServeMux()
	registerScreenplayFileRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, authors)
	hash, _ := numberedScreenplaySource("原文")
	for mode := 0; mode < 3; mode++ {
		imageOnly := mode == 1
		episode := 1
		duration := float64(3)
		video := "视频词"
		if imageOnly {
			episode = 2
			duration = 0
			video = ""
		}
		shot := screenplay.StoryboardShot{Duration: duration, Characters: "甲", Scene: "家", Props: "信", VideoDesc: video, ImagePrompt: "图片词", VoiceActor: "甲", Emotion: "平静", Line: "这是台词", SourceElements: json.RawMessage(`{"lines":[1,2]}`), SourceSceneIndex: 2, SourceSceneLocalOrder: 1, ProtocolFields: json.RawMessage(`{"camera":"wide","scene_id":"source-A"}`)}
		if mode == 2 {
			episode = 3
			duration = 0
			shot.Duration = 0
			selected, err := prompt.ResolveCreativePrompt("luoshui_storyboard_137", "storyboard")
			if err != nil {
				t.Fatal(err)
			}
			shot.ProtocolFields, _ = json.Marshal(map[string]any{"video_desc": shot.VideoDesc, "_cm_contract": map[string]any{"prompt_id": selected.ID, "prompt_digest": selected.Digest, "output_protocol": selected.OutputProtocol, "parser_version": selected.ParserVersion}})
		}
		input := map[string]any{"expected_revision": 0, "source_sha256": hash, "storyboard": screenplay.Storyboard{Engine: screenplay.StoryboardEngine("prompt_137"), EpisodeNo: episode, Shots: []screenplay.StoryboardShot{shot}}}
		result := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/protocol-files/storyboards/import", input)
		if result.Code != 200 {
			t.Fatalf("import imageOnly=%v %d %s", imageOnly, result.Code, result.Body.String())
		}
		var exported struct {
			Board screenplay.Storyboard `json:"storyboard"`
		}
		if err := json.Unmarshal(result.Body.Bytes(), &exported); err != nil {
			t.Fatal(err)
		}
		got := exported.Board.Shots[0]
		if got.ImagePrompt != shot.ImagePrompt || got.VoiceActor != shot.VoiceActor || got.Emotion != shot.Emotion || got.Line != shot.Line || got.SourceSceneIndex != 2 || got.SourceSceneLocalOrder != 1 || got.Duration != duration || got.VideoDesc != video || string(got.SourceElements) != string(shot.SourceElements) || string(got.ProtocolFields) != string(shot.ProtocolFields) {
			t.Fatalf("fields lost: %+v", got)
		}
		// The downloaded file contains this complete board. Import the JSON-decoded
		// export again to prove no field disappears at the file boundary.
		input["expected_revision"] = exported.Board.Revision
		input["storyboard"] = exported.Board
		again := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/protocol-files/storyboards/import", input)
		if again.Code != 200 {
			t.Fatalf("reimport %d %s", again.Code, again.Body.String())
		}
		reloaded, err := store.GetStoryboard(ctx, "alice", "protocol-files", episode, screenplay.StoryboardEngine("prompt_137"))
		if err != nil {
			t.Fatal(err)
		}
		if reloaded.Shots[0].ImagePrompt != "图片词" || reloaded.Shots[0].VoiceActor != "甲" || reloaded.Shots[0].Line != "这是台词" || string(reloaded.Shots[0].ProtocolFields) != string(shot.ProtocolFields) {
			t.Fatal(reloaded.Shots[0])
		}
	}
}
