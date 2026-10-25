package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestResolveShuihuoAssetImageSettings(t *testing.T) {
	overriddenModel, overriddenRatio := resolveShuihuoAssetImageSettings(
		batchfactoryv11.BookAsset{ImageModelID: "image-seedream", ImageAspectRatio: "1:1"},
		"image-default",
		"9:16",
	)
	if overriddenModel != "image-seedream" || overriddenRatio != "1:1" {
		t.Fatalf("override resolved to model=%q ratio=%q", overriddenModel, overriddenRatio)
	}

	fallbackModel, fallbackRatio := resolveShuihuoAssetImageSettings(
		batchfactoryv11.BookAsset{},
		"image-default",
		"9:16",
	)
	if fallbackModel != "image-default" || fallbackRatio != "9:16" {
		t.Fatalf("fallback resolved to model=%q ratio=%q", fallbackModel, fallbackRatio)
	}
}

func TestGenerateShuihuoImageArtifactPassesReferencesToLocalExecutor(t *testing.T) {
	ctx := context.Background()
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	executors := localexecutor.NewService(localexecutor.NewMemoryStore(), time.Now)
	pairing, err := executors.CreatePairing(ctx, "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := executors.Pair(ctx, localexecutor.PairInput{Code: pairing.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "comic-ref-test", OS: "windows", Version: "test"})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		claim, claimErr := executors.ClaimJobWait(ctx, paired.Token, 2*time.Second)
		if claimErr != nil {
			done <- claimErr
			return
		}
		var payload map[string]any
		if decodeErr := json.Unmarshal(claim.Job.Payload, &payload); decodeErr != nil {
			done <- decodeErr
			return
		}
		refs, _ := payload["referenceImages"].([]any)
		if len(refs) != 2 || refs[0] != "https://example.com/page.jpg" || refs[1] != "https://example.com/global.jpg" {
			done <- fmt.Errorf("referenceImages=%#v", payload["referenceImages"])
			return
		}
		lease := localexecutor.LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
		for _, state := range []localexecutor.JobState{localexecutor.JobPreparing, localexecutor.JobSubmitting} {
			if progressErr := executors.RecordProgress(ctx, paired.Token, claim.Job.ID, lease, state); progressErr != nil {
				done <- progressErr
				return
			}
		}
		if acceptErr := executors.RecordAcceptance(ctx, paired.Token, claim.Job.ID, lease, localexecutor.AcceptanceInput{AccountID: "account", SubmissionID: "submission"}); acceptErr != nil {
			done <- acceptErr
			return
		}
		for _, state := range []localexecutor.JobState{localexecutor.JobGenerating, localexecutor.JobDownloading, localexecutor.JobUploading} {
			if progressErr := executors.RecordProgress(ctx, paired.Token, claim.Job.ID, lease, state); progressErr != nil {
				done <- progressErr
				return
			}
		}
		png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
		saved, saveErr := files.SaveImage("comic_ref_result", "image/png", bytes.NewReader(png))
		if saveErr != nil {
			done <- saveErr
			return
		}
		digest := sha256.Sum256(png)
		artifact, artifactErr := executors.RecordArtifact(ctx, paired.Token, claim.Job.ID, lease, localexecutor.ArtifactInput{ID: saved.ID, MediaType: saved.MediaType, ByteSize: saved.ByteSize, SHA256: hex.EncodeToString(digest[:]), StorageRef: saved.StorageRef})
		if artifactErr != nil {
			done <- artifactErr
			return
		}
		done <- executors.CompleteJob(ctx, paired.Token, claim.Job.ID, lease, localexecutor.ResultInput{ArtifactID: artifact.ID})
	}()
	target := shuihuoAssetImageTarget{Provider: batchfactoryv11.VideoProviderDoubaoLocal, ModelID: "Seedream 5.0 Pro", AspectRatio: "9:16"}
	_, _, _, err = generateShuihuoImageArtifact(ctx, files, "alice", target, "漫画页提示词", []string{"https://example.com/page.jpg", "https://example.com/global.jpg"}, executors)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestResolveShuihuoAssetImageTargetUsesEnabledSystemDefault(t *testing.T) {
	target, err := resolveShuihuoAssetImageTarget(
		batchfactoryv11.BookAsset{},
		"",
		"9:16",
		"",
		[]map[string]any{{"id": "catalog-default", "kind": "image", "enabled": true, "provider": "personal_api", "modelId": "provider-default"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if target.Provider != "personal_api" || target.ModelID != "provider-default" || target.AspectRatio != "9:16" {
		t.Fatalf("target=%+v", target)
	}
}

func TestResolveShuihuoAssetImageTargetSupportsDoubaoLocalWithoutCatalog(t *testing.T) {
	target, err := resolveShuihuoAssetImageTarget(
		batchfactoryv11.BookAsset{},
		"Seedream 4.5",
		"3:4",
		"doubao_local_executor",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if target.Provider != batchfactoryv11.VideoProviderDoubaoLocal || target.ModelID != "Seedream 4.5" || target.AspectRatio != "3:4" {
		t.Fatalf("target=%+v", target)
	}
}

func TestGenerateShuihuoAssetImagesResolvesOverridesPerSelectedAsset(t *testing.T) {
	ctx := context.Background()
	var requests []struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Size   string `json:"size"`
	}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Model  string `json:"model"`
			Prompt string `json:"prompt"`
			Size   string `json:"size"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode provider request: %v", err)
			return
		}
		requests = append(requests, request)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"iVBORw0KGgoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}]}`))
	}))
	defer provider.Close()

	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "素材图", Books: []batchfactoryv11.CreateBookInput{{Title: "第一章", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	overridden, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "覆盖素材", Prompt: "override prompt", ImageModelID: "image-seedream", ImageAspectRatio: "1:1"})
	if err != nil {
		t.Fatal(err)
	}
	inherited, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "scene", Name: "继承素材", Prompt: "fallback prompt"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}

	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(ctx, "alice", map[string]any{"modelCatalog": []any{
		map[string]any{"id": "image-seedream", "kind": "image", "enabled": true, "modelId": "seedream-4.5", "baseUrl": provider.URL, "credential": "test-secret"},
		map[string]any{"id": "image-default", "kind": "image", "enabled": true, "modelId": "default-image", "baseUrl": provider.URL, "credential": "test-secret"},
	}}); err != nil {
		t.Fatal(err)
	}

	generated, err := generateShuihuoAssetImages(ctx, store, localartifact.NewStore(t.TempDir(), 1<<20), "alice", batch, []string{overridden.ID, inherited.ID}, "image-default", "9:16", "provider", nil, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(generated) != 2 {
		t.Fatalf("generated=%d, want 2", len(generated))
	}
	if len(requests) != 2 {
		t.Fatalf("requests=%d, want 2", len(requests))
	}
	got := map[string]struct{ model, size string }{}
	for _, request := range requests {
		got[request.Prompt] = struct{ model, size string }{model: request.Model, size: request.Size}
	}
	if got["override prompt"] != (struct{ model, size string }{model: "seedream-4.5", size: "1024x1024"}) {
		t.Fatalf("override request=%+v", got["override prompt"])
	}
	if got["fallback prompt"] != (struct{ model, size string }{model: "default-image", size: "1024x1792"}) {
		t.Fatalf("fallback request=%+v", got["fallback prompt"])
	}
}

func TestGenerateShuihuoAssetImagesDispatchesMixedProvidersPerAsset(t *testing.T) {
	ctx := context.Background()
	providerRequests := make(chan map[string]any, 1)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode provider request: %v", err)
			return
		}
		providerRequests <- payload
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="}]}`))
	}))
	defer provider.Close()

	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "混合素材图", Books: []batchfactoryv11.CreateBookInput{{Title: "第一章", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	externalAsset, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "scene", Name: "外部模型素材", Prompt: "external prompt", ImageModelID: "external-catalog", ImageAspectRatio: "16:9"})
	if err != nil {
		t.Fatal(err)
	}
	localAsset, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "本地模型素材", Prompt: "local prompt", ImageModelID: "Seedream 5.0 Pro", ImageAspectRatio: "1:1"})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}

	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(ctx, "alice", map[string]any{"modelCatalog": []any{
		map[string]any{"id": "external-catalog", "kind": "image", "enabled": true, "provider": "personal_api", "modelId": "external-image-v2", "baseUrl": provider.URL, "credential": "test-secret"},
		map[string]any{"id": "local-catalog", "kind": "image", "enabled": true, "provider": "doubao_local_executor", "modelId": "Seedream 5.0 Pro"},
	}}); err != nil {
		t.Fatal(err)
	}

	files := localartifact.NewStore(t.TempDir(), 1<<20)
	executorStore := localexecutor.NewMemoryStore()
	executors := localexecutor.NewService(executorStore, time.Now)
	pairing, err := executors.CreatePairing(ctx, "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := executors.Pair(ctx, localexecutor.PairInput{Code: pairing.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "dry-run", OS: "windows", Version: "test"})
	if err != nil {
		t.Fatal(err)
	}
	localPayloads := make(chan map[string]any, 1)
	executorDone := make(chan error, 1)
	go func() {
		claim, claimErr := executors.ClaimJobWait(ctx, paired.Token, 2*time.Second)
		if claimErr != nil {
			executorDone <- claimErr
			return
		}
		var payload map[string]any
		if decodeErr := json.Unmarshal(claim.Job.Payload, &payload); decodeErr != nil {
			executorDone <- decodeErr
			return
		}
		localPayloads <- payload
		lease := localexecutor.LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
		for _, state := range []localexecutor.JobState{localexecutor.JobPreparing, localexecutor.JobSubmitting} {
			if progressErr := executors.RecordProgress(ctx, paired.Token, claim.Job.ID, lease, state); progressErr != nil {
				executorDone <- progressErr
				return
			}
		}
		if acceptErr := executors.RecordAcceptance(ctx, paired.Token, claim.Job.ID, lease, localexecutor.AcceptanceInput{AccountID: "dry-run-account", SubmissionID: "dry-run-submission"}); acceptErr != nil {
			executorDone <- acceptErr
			return
		}
		for _, state := range []localexecutor.JobState{localexecutor.JobGenerating, localexecutor.JobDownloading, localexecutor.JobUploading} {
			if progressErr := executors.RecordProgress(ctx, paired.Token, claim.Job.ID, lease, state); progressErr != nil {
				executorDone <- progressErr
				return
			}
		}
		png, decodeErr := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
		if decodeErr != nil {
			executorDone <- decodeErr
			return
		}
		saved, saveErr := files.SaveImage("lea_shuihuo_dry_run", "image/png", bytes.NewReader(png))
		if saveErr != nil {
			executorDone <- saveErr
			return
		}
		digest := sha256.Sum256(png)
		artifact, artifactErr := executors.RecordArtifact(ctx, paired.Token, claim.Job.ID, lease, localexecutor.ArtifactInput{ID: saved.ID, MediaType: saved.MediaType, ByteSize: saved.ByteSize, SHA256: hex.EncodeToString(digest[:]), StorageRef: saved.StorageRef})
		if artifactErr != nil {
			executorDone <- artifactErr
			return
		}
		executorDone <- executors.CompleteJob(ctx, paired.Token, claim.Job.ID, lease, localexecutor.ResultInput{ArtifactID: artifact.ID})
	}()

	generated, err := generateShuihuoAssetImages(ctx, store, files, "alice", batch, []string{externalAsset.ID, localAsset.ID}, "external-catalog", "9:16", "personal_api", executors, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-executorDone; err != nil {
		t.Fatal(err)
	}
	if len(generated) != 2 {
		t.Fatalf("generated=%d, want 2", len(generated))
	}
	select {
	case payload := <-providerRequests:
		if payload["model"] != "external-image-v2" || payload["prompt"] != "external prompt" || payload["size"] != "1792x1024" {
			t.Fatalf("external payload=%v", payload)
		}
	default:
		t.Fatal("external provider did not receive its selected asset")
	}
	select {
	case payload := <-localPayloads:
		if payload["model"] != "Seedream 5.0 Pro" || payload["prompt"] != "local prompt" || payload["size"] != "1024x1024" {
			t.Fatalf("local payload=%v", payload)
		}
	default:
		t.Fatal("local executor did not receive its selected asset")
	}
	if len(providerRequests) != 0 || len(localPayloads) != 0 {
		t.Fatal(fmt.Sprintf("unexpected extra dispatches: external=%d local=%d", len(providerRequests), len(localPayloads)))
	}
}
