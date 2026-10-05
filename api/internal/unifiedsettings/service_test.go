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

type snapshotSource struct{}
func (snapshotSource) Get121Snapshot(context.Context, int64) (map[string]any, error) {
	return map[string]any{"sources": []map[string]string{{"source": "知乎", "platformId": "4"}}}, nil
}
func (snapshotSource) GetStyleTypeSnapshot(context.Context, int64) (map[string]any, error) {
	return map[string]any{"styles": []string{"剧情"}, "genres": []string{"都市"}, "genders": []string{"女频"}}, nil
}

func TestSaveProductionDoesNotOverwritePublishing(t *testing.T) {
	store := &memoryStore{project: Settings{Publishing: map[string]any{"uploadVideoType": "individual"}}}
	service := NewService(store, StaticDefaults{})

	_, err := service.SaveProduction(context.Background(), 12, map[string]any{"mode": "viral"})
	if err != nil { t.Fatal(err) }
	if store.project.Production["mode"] != "viral" { t.Fatalf("production not saved: %#v", store.project.Production) }
	if store.project.Publishing["uploadVideoType"] != "individual" { t.Fatalf("publishing overwritten: %#v", store.project.Publishing) }
}

func TestSavePublishingPersistsAndReloads(t *testing.T) {
	store := &memoryStore{profile: VersionProfile{ProjectID: 12, Name: "默认版本配置档", Version: "v1"}}
	service := NewService(store, StaticDefaults{})
	current, err := service.SavePublishing(context.Background(), 12, map[string]any{"versionProfile": "女频短剧版"})
	if err != nil { t.Fatal(err) }
	if got := current.Project.Publishing["versionProfile"]; got != "女频短剧版" { t.Fatalf("publishing = %#v", current.Project.Publishing) }
}

func TestSync121PersistsIntoVersionProfile(t *testing.T) {
	store := &memoryStore{profile: VersionProfile{ProjectID: 12, Name: "默认版本配置档", Version: "v1"}}
	service := NewService(store, StaticDefaults{}, snapshotSource{})
	current, err := service.Sync121(context.Background(), 12)
	if err != nil { t.Fatal(err) }
	if len(current.Profile.Settings.Website121) == 0 { t.Fatalf("121 snapshot not persisted: %#v", current.Profile.Settings.Website121) }
	if len(store.profile.Settings.Website121) == 0 { t.Fatalf("store profile missing 121 snapshot: %#v", store.profile.Settings.Website121) }
}

func TestSyncStyleTypesPersistsIntoVersionProfile(t *testing.T) {
	store := &memoryStore{profile: VersionProfile{ProjectID: 12, Name: "默认版本配置档", Version: "v1"}}
	service := NewService(store, StaticDefaults{}, snapshotSource{})
	current, err := service.SyncStyleTypes(context.Background(), 12)
	if err != nil { t.Fatal(err) }
	styles, ok := current.Profile.Settings.StyleTypes["styles"].([]string)
	if !ok || len(styles) != 1 || styles[0] != "剧情" { t.Fatalf("style snapshot = %#v", current.Profile.Settings.StyleTypes) }
	if len(store.profile.Settings.StyleTypes) == 0 { t.Fatalf("store profile missing style snapshot: %#v", store.profile.Settings.StyleTypes) }
}

func TestSaveProfileKeepsPromptReferencesSeparate(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, StaticDefaults{})
	current, err := service.SaveProfile(context.Background(), 12, VersionProfile{
		Name: "女频短剧版",
		Version: "v3",
		Settings: Settings{ProcessingRulePromptRef: "processing-v3", KnowledgePromptRef: "knowledge-v7"},
	})
	if err != nil { t.Fatal(err) }
	if current.Profile.Settings.ProcessingRulePromptRef != "processing-v3" || current.Profile.Settings.KnowledgePromptRef != "knowledge-v7" {
		t.Fatalf("prompt refs = %#v", current.Profile.Settings)
	}
}
