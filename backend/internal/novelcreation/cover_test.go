package novelcreation

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubCoverGenerator struct {
	url    string
	err    error
	prompt string
	model  string
}

func (s *stubCoverGenerator) GenerateCoverImage(_ context.Context, _, prompt, model string) (string, error) {
	s.prompt, s.model = prompt, model
	return s.url, s.err
}

func coverFixture(t *testing.T) (*Service, Project) {
	t.Helper()
	service := NewService(NewMemoryStore(), NewMemoryRuntime(), nil)
	project, err := service.CreateProject(context.Background(), "alice", CreateProjectInput{
		Title: "雨幕记者", Genre: "悬疑", World: "旧城连日暴雨", CoreConflict: "真相与自保", ModelID: "text-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	return service, project
}

func TestBuildCoverPromptIncludesProjectFacts(t *testing.T) {
	project := Project{Title: "雨幕记者", Genre: "悬疑", World: "旧城连日暴雨", CoreConflict: "真相与自保"}
	prompt := buildCoverPrompt(project, "")
	for _, want := range []string{"雨幕记者", "悬疑", "旧城连日暴雨", "真相与自保", "网文封面", "不含文字水印"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("missing %q in %q", want, prompt)
		}
	}
	custom := buildCoverPrompt(project, "水墨风")
	if !strings.Contains(custom, "水墨风") || strings.Contains(custom, "电影海报感") {
		t.Fatalf("style hint not applied: %q", custom)
	}
}

func TestGenerateProjectCoverRequiresGenerator(t *testing.T) {
	service, project := coverFixture(t)
	if _, err := service.GenerateProjectCover(context.Background(), "alice", project.ID, "img-1", ""); err == nil {
		t.Fatal("missing generator must fail")
	}
}

func TestGenerateProjectCoverSavesAndReads(t *testing.T) {
	service, project := coverFixture(t)
	generator := &stubCoverGenerator{url: "https://example.test/cover.png"}
	service.ConfigureCoverGenerator(generator)
	cover, err := service.GenerateProjectCover(context.Background(), "alice", project.ID, "img-1", "水墨风")
	if err != nil {
		t.Fatal(err)
	}
	if cover.ImageURL != generator.url || cover.ModelID != "img-1" || generator.model != "img-1" {
		t.Fatalf("cover=%+v generator=%+v", cover, generator)
	}
	if !strings.Contains(cover.Caption, "雨幕记者") || !strings.Contains(generator.prompt, "水墨风") {
		t.Fatalf("caption=%q prompt=%q", cover.Caption, generator.prompt)
	}
	saved, found, err := service.GetProjectCover(context.Background(), "alice", project.ID)
	if err != nil || !found || saved.ImageURL != generator.url || saved.Prompt != cover.Prompt {
		t.Fatalf("read back=%+v found=%v err=%v", saved, found, err)
	}
}

func TestGenerateProjectCoverFailureKeepsOldCover(t *testing.T) {
	service, project := coverFixture(t)
	good := &stubCoverGenerator{url: "https://example.test/old.png"}
	service.ConfigureCoverGenerator(good)
	if _, err := service.GenerateProjectCover(context.Background(), "alice", project.ID, "img-1", ""); err != nil {
		t.Fatal(err)
	}
	service.ConfigureCoverGenerator(&stubCoverGenerator{err: errors.New("上游失败")})
	if _, err := service.GenerateProjectCover(context.Background(), "alice", project.ID, "img-1", ""); err == nil {
		t.Fatal("generator error must surface")
	}
	saved, found, err := service.GetProjectCover(context.Background(), "alice", project.ID)
	if err != nil || !found || saved.ImageURL != "https://example.test/old.png" {
		t.Fatalf("old cover lost: %+v found=%v err=%v", saved, found, err)
	}
}

func TestGetProjectCoverMissingReturnsNotFound(t *testing.T) {
	service, project := coverFixture(t)
	if _, found, err := service.GetProjectCover(context.Background(), "alice", project.ID); err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}
}
