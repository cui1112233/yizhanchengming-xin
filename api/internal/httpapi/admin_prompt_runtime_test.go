package httpapi

import (
	"context"
	"errors"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"net/http"
	"net/http/httptest"
)

type adminRuntimePromptStore struct {
	prompt generation.Prompt
}

func (s *adminRuntimePromptStore) GetBookForProject(context.Context, int64, int64) (intake.Book, error) {
	return intake.Book{}, generation.ErrNotFound
}
func (s *adminRuntimePromptStore) ListBooksForProject(context.Context, int64) ([]intake.Book, error) {
	return nil, nil
}
func (s *adminRuntimePromptStore) ResolvePrompt(_ context.Context, key string) (generation.Prompt, error) {
	if key != s.prompt.Key || !s.prompt.Enabled {
		return generation.Prompt{}, generation.ErrNotFound
	}
	return s.prompt, nil
}
func (s *adminRuntimePromptStore) ListPrompts(context.Context) ([]generation.Prompt, error) {
	return []generation.Prompt{s.prompt}, nil
}
func (s *adminRuntimePromptStore) CreateBookRun(context.Context, generation.BookRun) (generation.BookRun, error) {
	return generation.BookRun{}, errors.New("unused")
}
func (s *adminRuntimePromptStore) UpdateBookRun(context.Context, generation.BookRun) (generation.BookRun, error) {
	return generation.BookRun{}, errors.New("unused")
}
func (s *adminRuntimePromptStore) LatestBookRun(context.Context, int64, int64) (generation.BookRun, error) {
	return generation.BookRun{}, generation.ErrNotFound
}
func (s *adminRuntimePromptStore) ListBookRunsByProject(context.Context, int64) ([]generation.BookRun, error) {
	return nil, nil
}
func (s *adminRuntimePromptStore) CreateStageRun(context.Context, generation.StageRun) (generation.StageRun, error) {
	return generation.StageRun{}, errors.New("unused")
}
func (s *adminRuntimePromptStore) UpdateStageRun(context.Context, generation.StageRun) (generation.StageRun, error) {
	return generation.StageRun{}, errors.New("unused")
}
func (s *adminRuntimePromptStore) ListStageRuns(context.Context, int64) ([]generation.StageRun, error) {
	return nil, nil
}
func (s *adminRuntimePromptStore) LatestStageRun(context.Context, int64, generation.Stage) (generation.StageRun, error) {
	return generation.StageRun{}, generation.ErrNotFound
}
func (s *adminRuntimePromptStore) AudioMeasurementByAsset(context.Context, int64, int64, string) (generation.AudioMeasurement, error) {
	return generation.AudioMeasurement{}, generation.ErrNotFound
}
func (s *adminRuntimePromptStore) LatestAudioMeasurement(context.Context, int64, int64) (generation.AudioMeasurement, error) {
	return generation.AudioMeasurement{}, generation.ErrNotFound
}
func (s *adminRuntimePromptStore) CreateAudioMeasurement(context.Context, generation.AudioMeasurement) (generation.AudioMeasurement, error) {
	return generation.AudioMeasurement{}, errors.New("unused")
}

type runtimeChangingAdminPromptService struct {
	runtime *adminRuntimePromptStore
	runtimePromptVersions
}

type runtimePromptVersions struct {
	versions []generation.AdminPrompt
}

func (s *runtimeChangingAdminPromptService) ListAdminPrompts(context.Context, string) ([]generation.AdminPrompt, error) {
	return append([]generation.AdminPrompt(nil), s.versions...), nil
}
func (s *runtimeChangingAdminPromptService) GetAdminPrompt(_ context.Context, key string, version int) (generation.AdminPrompt, error) {
	for _, prompt := range s.versions {
		if prompt.Key == key && prompt.Version == version {
			return prompt, nil
		}
	}
	return generation.AdminPrompt{}, generation.ErrNotFound
}
func (s *runtimeChangingAdminPromptService) CreateAdminPromptDraft(context.Context, int64, string, string) error {
	return errors.New("unused")
}
func (s *runtimeChangingAdminPromptService) UpdateAdminPromptDraft(context.Context, int64, string, int, string) error {
	return errors.New("unused")
}
func (s *runtimeChangingAdminPromptService) PublishAdminPrompt(_ context.Context, _ int64, key string, version int) error {
	var content string
	for index := range s.versions {
		if s.versions[index].Key == key {
			s.versions[index].Lifecycle = "archived"
			s.versions[index].Enabled = false
		}
		if s.versions[index].Key == key && s.versions[index].Version == version {
			content = s.versions[index].Content
			s.versions[index].Lifecycle = "published"
			s.versions[index].Enabled = true
		}
	}
	if content == "" {
		return generation.ErrNotFound
	}
	s.runtime.prompt = generation.Prompt{Key: key, Version: version, Enabled: true, Content: content}
	return nil
}
func (s *runtimeChangingAdminPromptService) RestoreAdminPrompt(_ context.Context, _ int64, key string, source int) error {
	var content string
	max := 0
	for _, prompt := range s.versions {
		if prompt.Key == key && prompt.Version > max {
			max = prompt.Version
		}
		if prompt.Key == key && prompt.Version == source {
			content = prompt.Content
		}
	}
	if content == "" {
		return generation.ErrNotFound
	}
	newVersion := max + 1
	s.versions = append(s.versions, generation.AdminPrompt{ID: 90, Key: key, Version: newVersion, Content: content, Lifecycle: "published", Enabled: true})
	s.runtime.prompt = generation.Prompt{Key: key, Version: newVersion, Enabled: true, Content: content}
	return nil
}

func TestAdminPromptPublishAndRestoreChangePromptResolverVersion(t *testing.T) {
	runtime := &adminRuntimePromptStore{prompt: generation.Prompt{Key: generation.PromptScript, Version: 3, Enabled: true, Content: "published-v3"}}
	service := &runtimeChangingAdminPromptService{runtime: runtime, runtimePromptVersions: runtimePromptVersions{versions: []generation.AdminPrompt{
		{ID: 11, Key: generation.PromptScript, Version: 3, Content: "published-v3", Lifecycle: "published", Enabled: true},
		{ID: 12, Key: generation.PromptScript, Version: 4, Content: "draft-v4", Lifecycle: "draft", Enabled: false},
	}}}
	h := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7, Capabilities: authn.EffectiveCapabilities("owner", nil)}}, AdminPrompts: service})
	resolver := generation.NewPromptResolver(runtime)

	before, err := resolver.Resolve(context.Background(), generation.PromptScript)
	if err != nil || before.Version != 3 {
		t.Fatalf("before=%#v err=%v", before, err)
	}
	publish := adminPromptHandlerRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/versions/4/publish", "", authn.User{})
	publish.Body = http.NoBody
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, publish)
	if rec.Code != http.StatusOK {
		t.Fatalf("publish status=%d body=%s", rec.Code, rec.Body.String())
	}
	afterPublish, err := resolver.Resolve(context.Background(), generation.PromptScript)
	if err != nil || afterPublish.Version != 4 || afterPublish.Content != "draft-v4" {
		t.Fatalf("after publish=%#v err=%v", afterPublish, err)
	}

	restore := adminPromptHandlerRequest(http.MethodPost, "/api/v1/admin/prompts/script.default/versions/3/restore", "", authn.User{})
	restore.Body = http.NoBody
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, restore)
	if rec.Code != http.StatusCreated {
		t.Fatalf("restore status=%d body=%s", rec.Code, rec.Body.String())
	}
	afterRestore, err := resolver.Resolve(context.Background(), generation.PromptScript)
	if err != nil || afterRestore.Version != 5 || afterRestore.Content != "published-v3" {
		t.Fatalf("after restore=%#v err=%v", afterRestore, err)
	}
}
