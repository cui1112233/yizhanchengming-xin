package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"qiantie/backend/internal/novelcreation"
	"testing"
	"time"
)

func TestNovelAutomationCreationIntentHTTPReplayAndConflict(t *testing.T) {
	for _, kind := range []string{"schedules", "batches"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			st := novelcreation.NewMemoryStore()
			s := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), nil)
			p, _ := s.CreateProject(ctx, "alice", novelcreation.CreateProjectInput{Title: "作品", Genre: "都市", ModelID: "model"})
			session, _ := s.CreateSession(ctx, "alice", novelcreation.CreateSessionInput{ProjectID: p.ID, Mode: "short_original"})
			note, _ := st.CreateNote(ctx, "alice", p.ID, novelcreation.ProjectNoteInput{Content: "原资料"})
			now := time.Now().UTC()
			mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: s})
			path := "/api/novel-creation/projects/" + p.ID + "/" + kind
			body := map[string]any{"idempotencyKey": "http-creation", "sessionId": session.ID, "prompt": "原始要求", "noteIds": []string{note.ID}}
			first := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, body)
			if first.Code != 201 {
				t.Fatal(first.Code, first.Body.String())
			}
			var initial map[string]json.RawMessage
			json.Unmarshal(first.Body.Bytes(), &initial)
			key := "schedule"
			if kind == "batches" {
				key = "batch"
			}
			var record struct {
				ID       string `json:"id"`
				Creation struct {
					Key string `json:"idempotencyKey"`
				} `json:"creation"`
			}
			json.Unmarshal(initial[key], &record)
			if record.ID == "" || record.Creation.Key != "http-creation" {
				t.Fatal(first.Body.String())
			}
			st.DeleteNote(ctx, "alice", p.ID, note.ID)
			if kind == "schedules" {
				plan, _ := st.GetSchedule(ctx, "alice", record.ID)
				plan.Prompt = "修改后的配置"
				plan.Enabled = false
				plan.Status = "stopped"
				plan.TotalSavedChapters = 8
				st.UpdateSchedule(ctx, "alice", plan)
			} else {
				plan, _ := st.GetBatch(ctx, "alice", record.ID)
				plan.Status = novelcreation.BatchStopped
				plan.CompletedChapters = 3
				st.UpdateBatch(ctx, "alice", plan)
			}
			if e := s.ArchiveProject(ctx, "alice", p.ID); e != nil {
				t.Fatal(e)
			}
			before, _ := st.GetProject(ctx, "alice", p.ID)
			replay := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, body)
			if replay.Code != 201 {
				t.Fatal(replay.Code, replay.Body.String())
			}
			var got map[string]json.RawMessage
			json.Unmarshal(replay.Body.Bytes(), &got)
			var out struct {
				ID        string `json:"id"`
				Status    string `json:"status"`
				Completed int    `json:"completedChapters"`
				Total     int    `json:"totalSavedChapters"`
			}
			json.Unmarshal(got[key], &out)
			if out.ID != record.ID || out.Status != "paused" || (kind == "batches" && out.Completed != 3) || (kind == "schedules" && out.Total != 8) {
				t.Fatal(replay.Body.String())
			}
			after, _ := st.GetProject(ctx, "alice", p.ID)
			if !after.UpdatedAt.Equal(before.UpdatedAt) {
				t.Fatal("read recovery touched retention clock")
			}
			body["prompt"] = "不同要求"
			conflict := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, body)
			if conflict.Code != 409 {
				t.Fatal(conflict.Code, conflict.Body.String())
			}
			var failure map[string]string
			json.Unmarshal(conflict.Body.Bytes(), &failure)
			if failure["code"] != novelcreation.ErrIdempotencyConflict {
				t.Fatal(failure)
			}
			foreign := signedJSONRequest(t, mux, now, "bob", http.MethodPost, path, body)
			if foreign.Code == 201 {
				t.Fatal("foreign intent returned")
			}
			body["idempotencyKey"] = "new-intent"
			fresh := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, body)
			if fresh.Code == 201 {
				t.Fatal("archived new creation")
			}
		})
	}
}
