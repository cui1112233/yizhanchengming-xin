package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/novelcreation"
)

func TestNovelCreationAutomationInputHTTPContract(t *testing.T) {
	for _, kind := range []string{"schedules", "batches", "messages"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			store := novelcreation.NewMemoryStore()
			service := novelcreation.NewService(store, novelcreation.NewMemoryRuntime(), nil)
			project, err := service.CreateProject(ctx, "alice", novelcreation.CreateProjectInput{Title: "作品", Genre: "都市", ModelID: "model"})
			if err != nil {
				t.Fatal(err)
			}
			session, err := service.CreateSession(ctx, "alice", novelcreation.CreateSessionInput{ProjectID: project.ID, Mode: "short_imitation"})
			if err != nil {
				t.Fatal(err)
			}
			material, _ := store.CreateMaterial(ctx, "alice", project.ID, novelcreation.CreateMaterialInput{Content: "资料正文"})
			note, _ := store.CreateNote(ctx, "alice", project.ID, novelcreation.ProjectNoteInput{Content: "备注正文"})
			now := time.Now().UTC()
			mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: service})
			styleResp := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/style-profiles", map[string]any{"title": "文风", "content": "简洁"})
			if styleResp.Code != http.StatusCreated {
				t.Fatalf("style %d: %s", styleResp.Code, styleResp.Body.String())
			}
			var style struct {
				StyleProfile novelcreation.StyleProfile `json:"styleProfile"`
			}
			if err = json.Unmarshal(styleResp.Body.Bytes(), &style); err != nil {
				t.Fatal(err)
			}
			foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/novel-creation/style-profiles/"+style.StyleProfile.ID, nil)
			if foreign.Code == http.StatusOK {
				t.Fatal("foreign account read style")
			}
			body := map[string]any{"sessionId": session.ID, "prompt": "创作", "content": "创作", "intervalMinutes": 60, "totalChapters": 2, "modelId": "model", "idempotencyKey": "http-input", "modeInputs": map[string]string{"referenceText": "参考文本"}, "knowledgeIds": []string{material.ID}, "noteIds": []string{note.ID}, "styleProfileId": style.StyleProfile.ID}
			switch kind {
			case "schedules":
				delete(body, "content")
				delete(body, "totalChapters")
				delete(body, "idempotencyKey")
			case "batches":
				delete(body, "content")
				delete(body, "intervalMinutes")
				delete(body, "idempotencyKey")
			case "messages":
				delete(body, "sessionId")
				delete(body, "prompt")
				delete(body, "totalChapters")
				delete(body, "intervalMinutes")
			}
			path := "/api/novel-creation/projects/" + project.ID + "/" + kind
			if kind == "messages" {
				path = "/api/novel-creation/sessions/" + session.ID + "/messages"
			}
			response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, body)
			expected := http.StatusCreated
			if kind == "messages" {
				expected = http.StatusAccepted
			}
			if response.Code != expected {
				t.Fatalf("%d: %s", response.Code, response.Body.String())
			}
			var run novelcreation.Run
			switch kind {
			case "schedules":
				var payload struct {
					Schedule novelcreation.Schedule `json:"schedule"`
				}
				json.Unmarshal(response.Body.Bytes(), &payload)
				plan := payload.Schedule
				if plan.ModeInputs["referenceText"] != "参考文本" || len(plan.NoteIDs) != 1 || len(plan.KnowledgeIDs) != 1 || plan.StyleProfileID != style.StyleProfile.ID {
					t.Fatalf("schedule dropped input: %+v", plan)
				}
				// Editing the saved task must preserve the same full contract through PUT.
				response = signedJSONRequest(t, mux, now, "alice", http.MethodPut, path+"/"+plan.ID, body)
				if response.Code != http.StatusOK {
					t.Fatalf("update %d: %s", response.Code, response.Body.String())
				}
				run, err = service.RunSchedule(ctx, "alice", project.ID, plan.ID)
			case "batches":
				var payload struct {
					Batch novelcreation.Batch `json:"batch"`
				}
				json.Unmarshal(response.Body.Bytes(), &payload)
				plan := payload.Batch
				if plan.ModeInputs["referenceText"] != "参考文本" || len(plan.NoteIDs) != 1 || len(plan.KnowledgeIDs) != 1 || plan.StyleProfileID != style.StyleProfile.ID {
					t.Fatalf("batch dropped input: %+v", plan)
				}
				var dispatched []novelcreation.DispatchedBatch
				dispatched, err = service.AdvanceBatches(ctx)
				if len(dispatched) != 1 {
					t.Fatalf("dispatch=%v", dispatched)
				}
				run = dispatched[0].Run
			case "messages":
				var payload struct {
					Run novelcreation.Run `json:"run"`
				}
				json.Unmarshal(response.Body.Bytes(), &payload)
				run, err = store.GetRun(ctx, "alice", payload.Run.ID)
			}
			if err != nil {
				t.Fatal(err)
			}
			var snapshot struct {
				ModeInputs map[string]string                 `json:"mode_inputs"`
				Knowledge  []novelcreation.KnowledgeMaterial `json:"knowledge_materials"`
				Notes      []novelcreation.ProjectNote       `json:"notes"`
				Style      string                            `json:"style_profile"`
			}
			if err = json.Unmarshal(run.PromptSnapshot, &snapshot); err != nil {
				t.Fatal(err)
			}
			if snapshot.ModeInputs["referenceText"] != "参考文本" || len(snapshot.Knowledge) != 1 || snapshot.Knowledge[0].Content != "资料正文" || len(snapshot.Notes) != 1 || snapshot.Notes[0].Content != "备注正文" || snapshot.Style != "简洁" {
				t.Fatalf("snapshot=%+v", snapshot)
			}
		})
	}
}
