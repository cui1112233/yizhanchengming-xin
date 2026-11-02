package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
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

	defaultModel, defaultRatio := resolveShuihuoAssetImageSettings(
		batchfactoryv11.BookAsset{},
		"image-default",
		"9:16",
	)
	if defaultModel != "image-default" || defaultRatio != "16:9" {
		t.Fatalf("default resolved to model=%q ratio=%q", defaultModel, defaultRatio)
	}
}

func TestResolveComicReferenceImageDataCompactsLargePNGForDoubao(t *testing.T) {
	files := localartifact.NewStore(t.TempDir(), 8<<20)
	large := image.NewRGBA(image.Rect(0, 0, 1200, 800))
	noise := rand.New(rand.NewSource(42))
	for y := 0; y < 800; y++ {
		for x := 0; x < 1200; x++ {
			large.SetRGBA(x, y, color.RGBA{
				R: uint8(noise.Intn(256)),
				G: uint8(noise.Intn(256)),
				B: uint8(noise.Intn(256)),
				A: 255,
			})
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, large); err != nil {
		t.Fatal(err)
	}
	if encoded.Len() <= 1<<20 {
		t.Fatalf("fixture is too small to exercise compression: %d bytes", encoded.Len())
	}
	saved, err := files.SaveImage("large-character-reference", "image/png", bytes.NewReader(encoded.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "漫画参考图"})
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := resolveComicReferenceImageData(context.Background(), store, files, nil, "alice", []string{
		"/api/local-executor/v1/batch-comic-references/" + batch.ID + "/" + strings.TrimSuffix(saved.StorageRef, ".png"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 1 || !strings.HasPrefix(resolved[0], "data:image/jpeg;base64,") {
		prefix := ""
		if len(resolved) == 1 {
			prefix = resolved[0][:min(len(resolved[0]), 32)]
		}
		t.Fatalf("large PNG was not compacted for Doubao: count=%d prefix=%q", len(resolved), prefix)
	}
	compacted, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(resolved[0], "data:image/jpeg;base64,"))
	if err != nil {
		t.Fatal(err)
	}
	if len(compacted) >= encoded.Len() || len(compacted) > 1<<20 {
		t.Fatalf("large PNG was not reduced enough for Doubao: before=%d after=%d", encoded.Len(), len(compacted))
	}
}

func TestResolveShuihuoAssetImageSettingsDefaultsEveryAssetKindToWideRatio(t *testing.T) {
	for _, kind := range []string{"character", "scene", "prop"} {
		t.Run(kind, func(t *testing.T) {
			model, ratio := resolveShuihuoAssetImageSettings(
				batchfactoryv11.BookAsset{Kind: kind, Source: "comic"},
				"image-default",
				"9:16",
			)
			if model != "image-default" || ratio != "16:9" {
				t.Fatalf("%s asset resolved to model=%q ratio=%q", kind, model, ratio)
			}
		})
	}
}

func TestCompileShuihuoComicCharacterPromptUsesEditableCMThreeViewPreset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prompt.EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	settings := batchfactoryv11.SettingsPatch{
		"productionMode": json.RawMessage(`"comic"`),
		"aiPromptConfig": json.RawMessage(`{"comic":{"enabled":true,"style":{"presetId":"style","body":"国风漫画画风"}}}`),
	}
	characterPrompt, _, err := compileShuihuoAssetImagePrompt(
		batchfactoryv11.BookAsset{Kind: "character", Source: "comic", Prompt: "青衣少女，银簪"},
		settings,
		prompt.NewCatalog(path),
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{"国风漫画画风", "青衣少女，银簪", "正面近照、正面全身、侧面全身、背面全身", "白色底图"} {
		if !strings.Contains(characterPrompt, phrase) {
			t.Fatalf("character prompt is missing %q: %s", phrase, characterPrompt)
		}
	}
	scenePrompt, _, err := compileShuihuoAssetImagePrompt(
		batchfactoryv11.BookAsset{Kind: "scene", Source: "comic", Prompt: "古宅庭院"},
		settings,
		prompt.NewCatalog(path),
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(scenePrompt, "正面近照") || strings.Contains(scenePrompt, "白色底图") {
		t.Fatalf("scene prompt unexpectedly contains character sheet instructions: %s", scenePrompt)
	}
}

func TestCompileShuihuoComicAssetPromptScopesGlobalStyleReferenceToCharacters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prompt.EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	settings := batchfactoryv11.SettingsPatch{
		"productionMode": json.RawMessage(`"comic"`),
		"aiPromptConfig": json.RawMessage(`{"comic":{"enabled":true,"style":{"presetId":"style","body":"都市彩漫画风"},"globalReferenceImage":{"assetId":"style:style","url":"/api/local-executor/v1/batch-comic-style-previews/style","source":"style-library","enabled":true,"applyToAssets":true,"applyToPages":true}}}`),
	}
	for _, test := range []struct {
		kind           string
		wantReferences int
	}{
		{kind: "character", wantReferences: 1},
		{kind: "scene", wantReferences: 0},
		{kind: "prop", wantReferences: 0},
	} {
		t.Run(test.kind, func(t *testing.T) {
			_, references, err := compileShuihuoAssetImagePrompt(
				batchfactoryv11.BookAsset{Kind: test.kind, Source: "comic", Prompt: "资产描述"},
				settings,
				prompt.NewCatalog(path),
				true,
			)
			if err != nil {
				t.Fatal(err)
			}
			if len(references) != test.wantReferences {
				t.Fatalf("%s references=%#v, want %d", test.kind, references, test.wantReferences)
			}
		})
	}
}

func TestCompileShuihuoComicAssetPromptAppliesStyleToLegacyAssetSource(t *testing.T) {
	settings := batchfactoryv11.SettingsPatch{
		"productionMode": json.RawMessage(`"comic"`),
		"aiPromptConfig": json.RawMessage(`{"comic":{"enabled":true,"style":{"presetId":"style","body":"Q版漫画画风"}}}`),
	}
	promptText, _, err := compileShuihuoAssetImagePrompt(
		batchfactoryv11.BookAsset{Kind: "scene", Source: "manual", Prompt: "现代办公室"},
		settings,
		nil,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(promptText, "Q版漫画画风") {
		t.Fatalf("legacy asset source lost comic style: %s", promptText)
	}
}

// 非漫画模式下，人物资产默认也套三视图，开关关闭时退化为单张设定图。
func TestCompileShuihuoAssetPromptThreeViewToggle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prompt.EnsureComicPresets(path); err != nil {
		t.Fatal(err)
	}
	catalog := prompt.NewCatalog(path)
	asset := batchfactoryv11.BookAsset{Kind: "character", Source: "manual", Prompt: "青衣少女，银簪"}

	withThreeView, _, err := compileShuihuoAssetImagePrompt(asset, batchfactoryv11.SettingsPatch{}, catalog, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{"青衣少女，银簪", "正面近照、正面全身、侧面全身、背面全身", "白色底图"} {
		if !strings.Contains(withThreeView, phrase) {
			t.Fatalf("three-view prompt is missing %q: %s", phrase, withThreeView)
		}
	}

	withoutThreeView, _, err := compileShuihuoAssetImagePrompt(asset, batchfactoryv11.SettingsPatch{}, catalog, false)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(withoutThreeView, "正面近照") || strings.Contains(withoutThreeView, "白色底图") {
		t.Fatalf("three-view disabled but template still applied: %s", withoutThreeView)
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
		lease := localexecutor.LeaseCredential{Token: claim.LeaseToken, Generation: claim.LeaseGeneration}
		refs, _ := payload["referenceImages"].([]any)
		if len(refs) != 2 || !strings.HasPrefix(fmt.Sprint(refs[0]), "data:image/jpeg;base64,") || refs[1] != "https://example.com/page.jpg" {
			_ = executors.FailJob(ctx, paired.Token, claim.Job.ID, lease, localexecutor.FailureInput{Code: "assertion", Message: "reference image was not resolved"})
			done <- fmt.Errorf("referenceImages=%#v", payload["referenceImages"])
			return
		}
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
	_, _, _, err = generateShuihuoImageArtifact(ctx, nil, files, nil, "alice", target, "漫画页提示词", []string{"/api/local-executor/v1/batch-comic-style-previews/batch-comic-style-16", "https://example.com/page.jpg"}, executors)
	if err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestGenerateShuihuoImageArtifactPassesReferencesToPersonalAPI(t *testing.T) {
	ctx := context.Background()
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/images/edits" {
			t.Errorf("path=%s, want /v1/images/edits", r.URL.Path)
		}
		if err := r.ParseMultipartForm(4 << 20); err != nil {
			t.Errorf("parse multipart: %v", err)
		}
		if r.FormValue("model") != "provider-image" || r.FormValue("prompt") != "漫画页提示词" {
			t.Errorf("form=%v", r.MultipartForm.Value)
		}
		if len(r.MultipartForm.File["image[]"]) != 1 {
			t.Errorf("reference files=%d", len(r.MultipartForm.File["image[]"]))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="}]}`))
	}))
	defer provider.Close()

	target := shuihuoAssetImageTarget{
		Provider:     "personal_api",
		ModelID:      "provider-image",
		AspectRatio:  "9:16",
		CatalogModel: map[string]any{"endpoint": provider.URL + "/v1/images/generations", "credential": "test-secret"},
	}
	_, _, source, err := generateShuihuoImageArtifact(ctx, nil, files, nil, "alice", target, "漫画页提示词", []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(png)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if source != "provider" {
		t.Fatalf("source=%q, want provider", source)
	}
}

func TestGenerateShuihuoImageArtifactDownloadsAuthorizedProviderURL(t *testing.T) {
	ctx := context.Background()
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	const credential = "test-secret"
	var providerURL string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/images/generations":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"data":[{"url":%q}]}`, providerURL+"/private/result.png")
		case "/private/result.png":
			if r.Header.Get("Authorization") != "Bearer "+credential {
				http.Error(w, "missing provider authorization", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(png)
		default:
			http.NotFound(w, r)
		}
	}))
	defer provider.Close()
	providerURL = provider.URL

	target := shuihuoAssetImageTarget{
		Provider:     "personal_api",
		ModelID:      "provider-image",
		AspectRatio:  "9:16",
		CatalogModel: map[string]any{"endpoint": provider.URL + "/v1/images/generations", "credential": credential},
	}
	_, _, source, err := generateShuihuoImageArtifact(ctx, nil, files, nil, "alice", target, "漫画页提示词", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if source != "provider" {
		t.Fatalf("source=%q, want provider", source)
	}
}

func TestShuihuoImageProviderErrorDetailExtractsActionableMessage(t *testing.T) {
	got := shuihuoImageProviderErrorDetail([]byte(`{"error":{"message":"image edits are not supported for this model","code":"unsupported_operation"}}`))
	if got != ": image edits are not supported for this model" {
		t.Fatalf("detail=%q", got)
	}
}

func TestShuihuoGptImageDefaultsToGenerationEndpointUnlessEditsAreDeclared(t *testing.T) {
	if shuihuoImageModelSupportsEdits(map[string]any{}, "custom-off-gpt-image-2-5") {
		t.Fatal("gpt-image model should use the generation endpoint by default")
	}
	if !shuihuoImageModelSupportsEdits(map[string]any{}, "provider-image") {
		t.Fatal("ordinary image model should retain edit support")
	}
	if !shuihuoImageModelSupportsEdits(map[string]any{"supportsImageEdits": true}, "custom-off-gpt-image-2-5") {
		t.Fatal("explicit edit capability should override the compatibility default")
	}
}

func TestShuihuoGPTImageRequestCompatibility(t *testing.T) {
	if got := shuihuoImageSizeForModel("[off]gpt-image-2.5", "9:16"); got != "1024x1536" {
		t.Fatalf("gpt-image portrait size = %q, want 1024x1536", got)
	}
	if got := shuihuoImageSizeForModel("[off]gpt-image-2.5", "16:9"); got != "1536x1024" {
		t.Fatalf("gpt-image landscape size = %q, want 1536x1024", got)
	}
	if got := shuihuoImageSizeForModel("seedream-4.5", "9:16"); got != "1024x1792" {
		t.Fatalf("seedream portrait size = %q, want legacy 1024x1792", got)
	}
}

func TestResolveComicReferenceImageDataReadsManualAndBookAssetImages(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画参考图", Books: []batchfactoryv11.CreateBookInput{{Title: "第一章", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	asset, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Source: "comic", Name: "姜予宁", Prompt: "人物描述"})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	manual, err := files.SaveImage("manual-reference-a", "image/png", bytes.NewReader(png))
	if err != nil {
		t.Fatal(err)
	}
	assetFile, err := files.SaveImage("character-reference-a", "image/png", bytes.NewReader(png))
	if err != nil {
		t.Fatal(err)
	}
	assetImage, err := store.CreateBookAssetImage(ctx, "alice", batch.ID, book.ID, asset.ID, batchfactoryv11.CreateBookAssetImageInput{StorageRef: assetFile.StorageRef, MediaType: assetFile.MediaType, Source: "upload"})
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := resolveComicReferenceImageData(ctx, store, files, nil, "alice", []string{
		"/api/local-executor/v1/batch-comic-references/" + batch.ID + "/" + strings.TrimSuffix(manual.StorageRef, ".png"),
		"/api/local-executor/v1/batch-book-asset-images/" + batch.ID + "/" + book.ID + "/" + asset.ID + "/" + assetImage.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 2 {
		t.Fatalf("resolved=%#v", resolved)
	}
	for index, value := range resolved {
		if !strings.HasPrefix(value, "data:image/png;base64,") {
			t.Fatalf("resolved[%d]=%q", index, value)
		}
	}
}

func TestResolveComicReferenceImageDataReadsAbsoluteLocalReferenceURLs(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画参考图"})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	png, _ := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	saved, err := files.SaveImage("absolute-local-reference", "image/png", bytes.NewReader(png))
	if err != nil {
		t.Fatal(err)
	}
	references := []string{
		"http://localhost:4001/api/local-executor/v1/batch-comic-references/" + batch.ID + "/" + strings.TrimSuffix(saved.StorageRef, ".png"),
		"http://localhost:4001/api/presets/comic-style/batch-comic-style-01/preview",
	}
	resolved, err := resolveComicReferenceImageData(ctx, store, files, nil, "alice", references)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 2 {
		t.Fatalf("resolved=%#v", resolved)
	}
	for index, value := range resolved {
		if !strings.HasPrefix(value, "data:image/") || !strings.Contains(value, ";base64,") {
			t.Fatalf("resolved[%d]=%q", index, value)
		}
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
	if target.Provider != "personal_api" || target.ModelID != "provider-default" || target.AspectRatio != "16:9" {
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
	if target.Provider != batchfactoryv11.VideoProviderDoubaoLocal || target.ModelID != "Seedream 4.5" || target.AspectRatio != "16:9" {
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
	if got["fallback prompt"] != (struct{ model, size string }{model: "default-image", size: "1792x1024"}) {
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

func TestShuihuoImageResponseContentRejectedRecognizesProviderSoftRefusal(t *testing.T) {
	softRefusal := []byte(`{"error":{"message":"目前未能成功生成图片。请重新发送一次生成请求，我会按要求制作 1536×1024 横版画面。(request id: 2026110117491859091069118142872)"}}`)
	if !shuihuoImageResponseContentRejected(400, softRefusal) {
		t.Fatal("provider soft refusal must go through the degraded-prompt retry path")
	}
	if shuihuoImageResponseContentRejected(400, []byte(`{"error":{"message":"balance pre-deduction insufficient"}}`)) {
		t.Fatal("account balance errors must stay permanent failures")
	}
}
