package unifiedsettings

import (
	"context"
	"errors"
)

type Store interface {
	GetVersionProfile(context.Context, int64) (VersionProfile, error)
	SaveVersionProfile(context.Context, VersionProfile) (VersionProfile, error)
	GetProjectSettings(context.Context, int64) (Settings, error)
	SaveProjectSettings(context.Context, int64, Settings) (Settings, error)
}

type SnapshotSource interface {
	Get121Snapshot(context.Context, int64) (map[string]any, error)
	GetStyleTypeSnapshot(context.Context, int64) (map[string]any, error)
}

type Defaults interface { DefaultSettings() Settings }

type StaticDefaults struct { Config Settings }
func (d StaticDefaults) DefaultSettings() Settings { return d.Config }

type Service struct {
	store Store
	defaults Defaults
	source SnapshotSource
}

func NewService(store Store, defaults Defaults, source ...SnapshotSource) *Service {
	s := &Service{store: store, defaults: defaults}
	if len(source) > 0 { s.source = source[0] }
	return s
}

func mergeMap(base, overlay map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range base { out[k] = v }
	for k, v := range overlay { out[k] = v }
	return out
}

func mergeSettings(base, overlay Settings) Settings {
	out := base
	out.Production = mergeMap(base.Production, overlay.Production)
	out.Publishing = mergeMap(base.Publishing, overlay.Publishing)
	out.Website121 = mergeMap(base.Website121, overlay.Website121)
	out.StyleTypes = mergeMap(base.StyleTypes, overlay.StyleTypes)
	if overlay.ProcessingRulePromptRef != "" { out.ProcessingRulePromptRef = overlay.ProcessingRulePromptRef }
	if overlay.KnowledgePromptRef != "" { out.KnowledgePromptRef = overlay.KnowledgePromptRef }
	return out
}

func (s *Service) Resolve(system, version, project Settings) Settings {
	return mergeSettings(mergeSettings(system, version), project)
}

func (s *Service) GetCurrent(ctx context.Context, projectID int64) (Current, error) {
	if s.store == nil { return Current{}, errors.New("unified settings store unavailable") }
	project, err := s.store.GetProjectSettings(ctx, projectID)
	if err != nil { return Current{}, err }
	profile, err := s.store.GetVersionProfile(ctx, projectID)
	if err != nil { return Current{}, err }
	system := Settings{}
	if s.defaults != nil { system = s.defaults.DefaultSettings() }
	return Current{
		ProjectID: projectID,
		Effective: s.Resolve(system, profile.Settings, project),
		Project: project,
		Profile: profile,
		Priority: []string{"project", "version_profile", "system_default"},
	}, nil
}

func (s *Service) SaveProduction(ctx context.Context, projectID int64, production map[string]any) (Current, error) {
	project, err := s.store.GetProjectSettings(ctx, projectID)
	if err != nil { return Current{}, err }
	project.Production = production
	if _, err := s.store.SaveProjectSettings(ctx, projectID, project); err != nil { return Current{}, err }
	return s.GetCurrent(ctx, projectID)
}

func (s *Service) SavePublishing(ctx context.Context, projectID int64, publishing map[string]any) (Current, error) {
	project, err := s.store.GetProjectSettings(ctx, projectID)
	if err != nil { return Current{}, err }
	project.Publishing = publishing
	if _, err := s.store.SaveProjectSettings(ctx, projectID, project); err != nil { return Current{}, err }
	return s.GetCurrent(ctx, projectID)
}

func (s *Service) SaveProfile(ctx context.Context, projectID int64, profile VersionProfile) (Current, error) {
	profile.ProjectID = projectID
	if profile.Name == "" { profile.Name = "默认版本配置档" }
	if profile.Version == "" { profile.Version = "v1" }
	if _, err := s.store.SaveVersionProfile(ctx, profile); err != nil { return Current{}, err }
	return s.GetCurrent(ctx, projectID)
}

func (s *Service) Sync121(ctx context.Context, projectID int64) (Current, error) {
	if s.source == nil { return Current{}, errors.New("121 snapshot source unavailable") }
	profile, err := s.store.GetVersionProfile(ctx, projectID)
	if err != nil { return Current{}, err }
	snapshot, err := s.source.Get121Snapshot(ctx, projectID)
	if err != nil { return Current{}, err }
	profile.ProjectID = projectID
	if profile.Name == "" { profile.Name = "默认版本配置档" }
	if profile.Version == "" { profile.Version = "v1" }
	profile.Settings.Website121 = snapshot
	if _, err := s.store.SaveVersionProfile(ctx, profile); err != nil { return Current{}, err }
	return s.GetCurrent(ctx, projectID)
}

func (s *Service) SyncStyleTypes(ctx context.Context, projectID int64) (Current, error) {
	if s.source == nil { return Current{}, errors.New("style/type snapshot source unavailable") }
	profile, err := s.store.GetVersionProfile(ctx, projectID)
	if err != nil { return Current{}, err }
	snapshot, err := s.source.GetStyleTypeSnapshot(ctx, projectID)
	if err != nil { return Current{}, err }
	profile.ProjectID = projectID
	if profile.Name == "" { profile.Name = "默认版本配置档" }
	if profile.Version == "" { profile.Version = "v1" }
	profile.Settings.StyleTypes = snapshot
	if _, err := s.store.SaveVersionProfile(ctx, profile); err != nil { return Current{}, err }
	return s.GetCurrent(ctx, projectID)
}
