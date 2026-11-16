package screenplay

import (
	"context"
	"testing"
)

func TestCreativeSubjectConfigPersistenceAndValidation(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	_, _ = s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "alice"})
	c := CreativeConfig{UseSubjectMode: true, VideoGenerationMode: "text_to_video"}
	p, e := s.SaveCreativeConfig(ctx, "alice", "p", c, 0)
	if e != nil {
		t.Fatal(e)
	}
	got, e := s.GetProject(ctx, "alice", "p")
	if e != nil || !got.CreativeConfig.UseSubjectMode || got.CreativeConfig.VideoGenerationMode != "text_to_video" {
		t.Fatal(got, e)
	}
	for _, invalid := range []CreativeConfig{{UseSubjectMode: true, AutoReplaceSora2Info: true, VideoGenerationMode: "text_to_video"}, {AutoReplaceSora2Info: true}, {VideoGenerationMode: "bad"}} {
		if _, e = s.SaveCreativeConfig(ctx, "alice", "p", invalid, p.Revision); e == nil {
			t.Fatalf("accepted invalid %+v", invalid)
		}
	}
	c = CreativeConfig{}
	if _, e = s.SaveCreativeConfig(ctx, "alice", "p", c, p.Revision); e != nil {
		t.Fatal(e)
	}
	got, _ = s.GetProject(ctx, "alice", "p")
	if got.CreativeConfig.UseSubjectMode || got.CreativeConfig.AutoReplaceSora2Info || got.CreativeConfig.VideoGenerationMode != "" {
		t.Fatal("disabled flags not persisted")
	}
}
