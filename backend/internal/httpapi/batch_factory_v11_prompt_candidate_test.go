package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

const promptCandidateDirectorJSON = `{"characters":[{"name":"林晚","prompt":"黑色长发"}],"scenes":[{"name":"客厅","prompt":"现代客厅"}],"props":[],"storyboard":[{"duration_sec":9,"characters":["林晚"],"props":[],"scene":"客厅","prefix_key":"modern_conflict","shots":[{"start_sec":0,"end_sec":9,"description":"第一候选镜头"}],"video_desc":"第一候选镜头"},{"duration_sec":9,"characters":["林晚"],"props":[],"scene":"客厅","prefix_key":"modern_conflict","shots":[{"start_sec":0,"end_sec":9,"description":"第二候选镜头"}],"video_desc":"第二候选镜头"}],"source_coverage":{"source_complete":true,"has_remaining_source":false}}`

func TestDirectorPromptCandidateRoutesPreserveUntilSelectedVideoAccepts(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "b",
		Books: []batchfactoryv11.CreateBookInput{{
			Title: "k", SourceText: "林晚进入客厅。",
			Videos: []batchfactoryv11.CreateVideoInput{
				{Label: "VIDEO 01", VideoPrompt: "保留第一条"},
				{Label: "VIDEO 02", VideoPrompt: "保留第二条"},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	firstVideo, secondVideo := book.Videos[0], book.Videos[1]
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 2, Store: store,
		Director: &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: promptCandidateDirectorJSON}},
	})
	// The browser workbench uses the public V12 namespace. V11 remains only
	// as an opt-in legacy contract and is not exposed by the review runtime.
	base := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/director-prompt-candidate"

	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, base, map[string]any{"expectedRevision": book.Revision})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var createdBody struct {
		Candidate struct {
			ID                 string            `json:"id"`
			SourceBookRevision int64             `json:"sourceBookRevision"`
			VideoPrompts       map[string]string `json:"videoPrompts"`
		} `json:"candidate"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createdBody); err != nil {
		t.Fatal(err)
	}
	if createdBody.Candidate.SourceBookRevision != book.Revision || createdBody.Candidate.ID == "" {
		t.Fatalf("candidate=%+v", createdBody.Candidate)
	}
	if !strings.Contains(createdBody.Candidate.VideoPrompts[firstVideo.ID], "第一候选镜头") || !strings.Contains(createdBody.Candidate.VideoPrompts[secondVideo.ID], "第二候选镜头") {
		t.Fatalf("candidate prompts=%#v", createdBody.Candidate.VideoPrompts)
	}

	loaded, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Books[0].DirectorRevision != nil || loaded.Books[0].Videos[0].VideoPrompt != "保留第一条" || loaded.Books[0].Videos[1].VideoPrompt != "保留第二条" {
		t.Fatalf("candidate generation mutated accepted state: %+v", loaded.Books[0])
	}
	if _, ok := loaded.Books[0].SettingsState.Patch["directorPromptCandidate"]; !ok {
		t.Fatalf("candidate was not persisted in the book settings: %#v", loaded.Books[0].SettingsState.Patch)
	}

	read := signedJSONRequest(t, api, now, "alice", http.MethodGet, base, nil)
	if read.Code != http.StatusOK || !strings.Contains(read.Body.String(), createdBody.Candidate.ID) {
		t.Fatalf("read status=%d body=%s", read.Code, read.Body.String())
	}

	stale := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+book.ID+"/videos/"+firstVideo.ID+"/director-prompt-candidate/accept", map[string]any{
		"candidateId": createdBody.Candidate.ID, "expectedRevision": firstVideo.Revision - 1,
	})
	if stale.Code != http.StatusConflict {
		t.Fatalf("stale accept status=%d body=%s", stale.Code, stale.Body.String())
	}

	accepted := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+book.ID+"/videos/"+firstVideo.ID+"/director-prompt-candidate/accept", map[string]any{
		"candidateId": createdBody.Candidate.ID, "expectedRevision": firstVideo.Revision,
	})
	if accepted.Code != http.StatusOK {
		t.Fatalf("accept status=%d body=%s", accepted.Code, accepted.Body.String())
	}

	loaded, err = store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	acceptedVideo := candidatePromptReadback(t, loaded.Books[0].Videos[0])
	if acceptedVideo.AIOriginalVideoPrompt != createdBody.Candidate.VideoPrompts[firstVideo.ID] || acceptedVideo.VideoPrompt != createdBody.Candidate.VideoPrompts[firstVideo.ID] {
		t.Fatalf("accepted video=%+v candidate=%+v", acceptedVideo, createdBody.Candidate)
	}
	if untouched := candidatePromptReadback(t, loaded.Books[0].Videos[1]); untouched.AIOriginalVideoPrompt != "" || untouched.VideoPrompt != "保留第二条" {
		t.Fatalf("other video changed: %+v", untouched)
	}

	cancel := signedJSONRequest(t, api, now, "alice", http.MethodDelete, base, map[string]any{"expectedRevision": loaded.Books[0].Revision})
	if cancel.Code != http.StatusOK {
		t.Fatalf("cancel status=%d body=%s", cancel.Code, cancel.Body.String())
	}
	loaded, err = store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := loaded.Books[0].SettingsState.Patch["directorPromptCandidate"]; ok {
		t.Fatalf("candidate survived cancel: %#v", loaded.Books[0].SettingsState.Patch)
	}
	if got := candidatePromptReadback(t, loaded.Books[0].Videos[0]); got.AIOriginalVideoPrompt != createdBody.Candidate.VideoPrompts[firstVideo.ID] || got.VideoPrompt != createdBody.Candidate.VideoPrompts[firstVideo.ID] {
		t.Fatalf("cancel changed accepted values: %+v", got)
	}
}

type candidatePromptVideo struct {
	VideoPrompt           string `json:"videoPrompt"`
	AIOriginalVideoPrompt string `json:"aiOriginalVideoPrompt"`
}

func candidatePromptReadback(t *testing.T, video batchfactoryv11.Video) candidatePromptVideo {
	t.Helper()
	raw, err := json.Marshal(video)
	if err != nil {
		t.Fatal(err)
	}
	var value candidatePromptVideo
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value
}
