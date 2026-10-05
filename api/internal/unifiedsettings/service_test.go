package unifiedsettings

import (
	"context"
	"reflect"
	"testing"
)

func TestResolvePrecedenceProjectOverVersionOverSystem(t *testing.T) {
	service := NewService(nil, StaticDefaults{Config: Settings{
		Production: map[string]any{"mode": "system", "copyCount": float64(1)},
		Publishing: map[string]any{"uploadVideoType": "merged"},
	}})

	resolved := service.Resolve(Settings{
		Production: map[string]any{"mode": "system", "copyCount": float64(1)},
		Publishing: map[string]any{"uploadVideoType": "merged"},
	}, Settings{
		Production: map[string]any{"mode": "version", "copyCount": float64(2)},
		Publishing: map[string]any{"materialReuse": true},
	}, Settings{
		Production: map[string]any{"mode": "project"},
	})

	wantProduction := map[string]any{"mode": "project", "copyCount": float64(2)}
	if !reflect.DeepEqual(resolved.Production, wantProduction) {
		t.Fatalf("production = %#v, want %#v", resolved.Production, wantProduction)
	}
	wantPublishing := map[string]any{"uploadVideoType": "merged", "materialReuse": true}
	if !reflect.DeepEqual(resolved.Publishing, wantPublishing) {
		t.Fatalf("publishing = %#v, want %#v", resolved.Publishing, wantPublishing)
	}
}

func TestServiceKeepsProcessingAndKnowledgePromptsSeparate(t *testing.T) {
	resolved := NewService(nil, StaticDefaults{}).Resolve(Settings{}, Settings{
		ProcessingRulePromptRef: "processing-v3",
		KnowledgePromptRef:      "knowledge-v7",
	}, Settings{})

	if resolved.ProcessingRulePromptRef != "processing-v3" {
		t.Fatalf("processing rule prompt = %q", resolved.ProcessingRulePromptRef)
	}
	if resolved.KnowledgePromptRef != "knowledge-v7" {
		t.Fatalf("knowledge prompt = %q", resolved.KnowledgePromptRef)
	}
}

type memoryStore struct {
	profile VersionProfile
	project Settings
}

func (m *memoryStore) GetVersionProfile(context.Context, int64) (VersionProfile, error) { return m.profile, nil }
func (m *memoryStore) SaveVersionProfile(_ context.Context, profile VersionProfile) (VersionProfile, error) { m.profile = profile; return profile, nil }
func (m *memoryStore) GetProjectSettings(context.Context, int64) (Settings, error) { return m.project, nil }
func (m *memoryStore) SaveProjectSettings(_ context.Context, _ int64, settings Settings) (Settings, error) { m.project = settings; return settings, nil }

func TestSaveProductionDoesNotOverwritePublishing(t *testing.T) {
	store := &memoryStore{project: Settings{Publishing: map[string]any{"uploadVideoType": "individual"}}}
	service := NewService(store, StaticDefaults{})

	_, err := service.SaveProduction(context.Background(), 12, map[string]any{"mode": "viral"})
	if err != nil { t.Fatal(err) }
	if store.project.Production["mode"] != "viral" { t.Fatalf("production not saved: %#v", store.project.Production) }
	if store.project.Publishing["uploadVideoType"] != "individual" { t.Fatalf("publishing overwritten: %#v", store.project.Publishing) }
}
