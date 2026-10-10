package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type fixedHTTPH3AudioProbe struct{ durationMS int64 }

func seedHTTPH3Assets(t *testing.T, store *batchfactoryv11.MemoryStore, batch batchfactoryv11.Batch, raw []byte) {
	t.Helper()
	var d batchfactoryv11.H3DirectorDocument
	if err := json.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	for _, c := range d.CharacterRoster {
		if _, err := store.CreateBookAsset(context.Background(), "alice", batch.ID, batch.Books[0].ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: c.CanonicalName, Prompt: c.Appearance}); err != nil {
			t.Fatal(err)
		}
	}
}

func (p fixedHTTPH3AudioProbe) DurationMS(context.Context, []byte) (int64, error) {
	return p.durationMS, nil
}

func TestV12H3DirectorRouteWritesValidatedCompleteDocument(t *testing.T) {
	now := time.Unix(1700000000, 0)
	raw, err := os.ReadFile(filepath.Join("..", "batchfactoryv11", "testdata", "h3_v12_complete_director_trace.json"))
	if err != nil {
		t.Fatal(err)
	}
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "raw novel"}}})
	seedHTTPH3Assets(t, store, batch, raw)
	director := &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: string(raw)}}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: store, Director: director})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/h3/director"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"video_source": map[string]any{
			"revision": "video-source-acceptance001-r1",
			"text":     "\n五岁的我刚被认回豪门，爸妈就甩下一百万生活费。\n\n把我和陆晚晚扔在别墅里大眼瞪小眼。\n三个月后，爸妈提前回国，想给我们一个惊喜。\n",
		},
		"preset": map[string]any{"key": "h3-director-normal", "revision": 1, "prompt_body": "complete H3"},
	})
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"h3_director"`) || !strings.Contains(rec.Body.String(), `"character_slot_ids":[]`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12H3TraceRouteShowsLegacyCompatibilityWithoutSynthesizingH3(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。"}}})
	director := &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: directorHTTPJSON}}
	if _, err := director.RunDirector(context.Background(), "alice", batch.ID, batch.Books[0].ID); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/h3/trace"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"legacy":true`) || !strings.Contains(rec.Body.String(), "旧版导演数据") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"director_document"`) {
		t.Fatalf("legacy trace was falsely synthesized: %s", rec.Body.String())
	}
}

func TestV12H3TraceRouteShowsDirectorDocumentBeforeCompilation(t *testing.T) {
	now := time.Unix(1700000000, 0)
	raw, err := os.ReadFile(filepath.Join("..", "batchfactoryv11", "testdata", "h3_v12_complete_director_trace.json"))
	if err != nil {
		t.Fatal(err)
	}
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "raw novel"}}})
	seedHTTPH3Assets(t, store, batch, raw)
	director := &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: string(raw)}}
	if _, err := director.RunH3Director(context.Background(), "alice", batch.ID, batch.Books[0].ID, batchfactoryv11.H3DirectorRunRequest{
		VideoSource: batchfactoryv11.H3VideoSource{Revision: "video-source-acceptance001-r1", Text: "\n五岁的我刚被认回豪门，爸妈就甩下一百万生活费。\n\n把我和陆晚晚扔在别墅里大眼瞪小眼。\n三个月后，爸妈提前回国，想给我们一个惊喜。\n"},
		Preset:      batchfactoryv11.H3DirectorPreset{Key: "h3-director-normal", Revision: 1},
	}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: store, Director: director})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/h3/trace"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"director_document"`) || !strings.Contains(rec.Body.String(), "尚未编译") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12H3CompileRouteDoesNotAcceptLegacyDirectorData(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。"}}})
	director := &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: directorHTTPJSON}}
	revision, err := director.RunDirector(context.Background(), "alice", batch.ID, batch.Books[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: store, Director: director})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/h3/compile"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"director_revision_id": revision.ID})
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12H3CompileRouteRejectsClientSuppliedCanonicalAudioDuration(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。"}}})
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: store})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/h3/compile"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"director_revision_id": "director-1",
		"audio":                map[string]any{"asset_id": "audio-1", "duration_ms": 7420},
	})
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "audio") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestV12H3AudioMeasurementRouteProbesBytesInsteadOfAcceptingDuration(t *testing.T) {
	now := time.Unix(1700000000, 0)
	raw, err := os.ReadFile(filepath.Join("..", "batchfactoryv11", "testdata", "h3_v12_complete_director_trace.json"))
	if err != nil {
		t.Fatal(err)
	}
	store := batchfactoryv11.NewMemoryStore()
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "raw novel"}}})
	seedHTTPH3Assets(t, store, batch, raw)
	director := &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: string(raw)}}
	revision, err := director.RunH3Director(context.Background(), "alice", batch.ID, batch.Books[0].ID, batchfactoryv11.H3DirectorRunRequest{
		VideoSource: batchfactoryv11.H3VideoSource{Revision: "video-source-acceptance001-r1", Text: "五岁的我刚被认回豪门，爸妈就甩下一百万生活费。\n把我和陆晚晚扔在别墅里大眼瞪小眼。\n三个月后，爸妈提前回国，想给我们一个惊喜。"},
		Preset:      batchfactoryv11.H3DirectorPreset{Key: "h3-director-normal", Revision: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = revision
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: store, Director: director, H3AudioProbe: fixedHTTPH3AudioProbe{durationMS: 7420}})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/h3/audio-measurement"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"audio_base64": base64.StdEncoding.EncodeToString([]byte("real-audio")),
	})
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"duration_ms":7420`) || !strings.Contains(rec.Body.String(), `"audio_asset_id":"h3-audio-`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	lines := []map[string]any{}
	for _, card := range revision.Output.H3Director.DirectorCards {
		lines = append(lines, map[string]any{"source_key": card.SourceKey, "source_text": card.SourceText, "audio_base64": base64.StdEncoding.EncodeToString([]byte("line-audio"))})
	}
	rec = signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"director_revision_id": revision.ID, "tts_fingerprint": "voice-settings", "lines": lines})
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"duration_ms":22260`) || !strings.Contains(rec.Body.String(), `"method":"per_line_tts_probe"`) {
		t.Fatalf("line measurements not persisted: status=%d body=%s", rec.Code, rec.Body.String())
	}
}
