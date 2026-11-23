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

func TestOpeningVariantsRouteReturnsGeneratedVariants(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", TxtText: "她被当众羞辱后沉默离开。", SourceMetadata: map[string]any{"rawSourceStatus": "verified"}, SourceText: "她被当众羞辱后沉默离开。"}}})
	if err != nil {
		t.Fatal(err)
	}
	patch := batchfactoryv11.SettingsPatch{
		"openingEnabled": json.RawMessage("true"),
		"openingCount":   json.RawMessage("4"),
	}
	if _, err := store.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: patch, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	storyboard := []batchfactoryv11.DirectorVideo{
		{DurationSec: 10, FinalPrompt: "原分镜一：新娘低头抚平嫁衣袖口。"},
		{DurationSec: 10, FinalPrompt: "分镜二：陆沉推门而入。"},
	}
	if _, err := store.PersistDirectorRevision(context.Background(), "alice", batch.Books[0], batchfactoryv11.DirectorSnapshot{Mode: "original", MaxVideoDuration: 10, AspectRatio: "9:16"}, "opening-route-digest", "", batchfactoryv11.DirectorResult{Storyboard: storyboard}); err != nil {
		t.Fatal(err)
	}
	reply := strings.Join([]string{
		"===VARIANT 1===",
		"时长：10秒",
		"变体一：茶盏砸落大理石地面碎裂，争吵爆发。",
		"",
		"===VARIANT 2===",
		"时长：10秒",
		"变体二：雨夜推门而入，两人对视沉默。",
		"",
		"===VARIANT 3===",
		"时长：10秒",
		"变体三：庭院灯影摇晃，她攥紧婚书。",
	}, "\n")
	provider := &directorHTTPProvider{output: reply}
	service := &batchfactoryv11.DirectorService{Store: store, Provider: provider}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 2, Store: store, Director: service})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/opening-variants"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"openingMeta": map[string]any{}})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := decodeBody[struct {
		Variants []batchfactoryv11.OpeningVariant `json:"variants"`
	}](t, rec)
	if len(body.Variants) != 3 {
		t.Fatalf("variants = %d, want 3", len(body.Variants))
	}
	if body.Variants[0].Status != "success" || !strings.Contains(body.Variants[0].Prompt, "茶盏砸落") {
		t.Fatalf("variant[0] = %#v", body.Variants[0])
	}
	if body.Variants[1].Status != "success" || !strings.Contains(body.Variants[1].Prompt, "雨夜推门而入") {
		t.Fatalf("variant[1] = %#v", body.Variants[1])
	}
	if body.Variants[2].Status != "success" || !strings.Contains(body.Variants[2].Prompt, "庭院灯影") {
		t.Fatalf("variant[2] = %#v", body.Variants[2])
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls=%d", provider.calls)
	}
}
