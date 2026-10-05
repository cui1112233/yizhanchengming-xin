package unifiedsettings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
)

type MySQLStore struct { db *sql.DB }
func NewMySQLStore(db *sql.DB) *MySQLStore { return &MySQLStore{db: db} }

func encode(value any) ([]byte, error) { return json.Marshal(value) }
func decodeSettings(raw []byte, out *Settings) error {
	if len(raw) == 0 { return nil }
	return json.Unmarshal(raw, out)
}

func (s *MySQLStore) GetProjectSettings(ctx context.Context, projectID int64) (Settings, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT settings_json FROM batch_project_settings WHERE batch_project_id = ?`, projectID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) { return Settings{}, nil }
	if err != nil { return Settings{}, err }
	var settings Settings
	if err := decodeSettings(raw, &settings); err != nil { return Settings{}, err }
	return settings, nil
}

func (s *MySQLStore) SaveProjectSettings(ctx context.Context, projectID int64, settings Settings) (Settings, error) {
	raw, err := encode(settings)
	if err != nil { return Settings{}, err }
	_, err = s.db.ExecContext(ctx, `INSERT INTO batch_project_settings (batch_project_id, settings_json) VALUES (?, ?) ON DUPLICATE KEY UPDATE settings_json = VALUES(settings_json), updated_at = CURRENT_TIMESTAMP(6)`, projectID, raw)
	if err != nil { return Settings{}, err }
	return settings, nil
}

func (s *MySQLStore) GetVersionProfile(ctx context.Context, projectID int64) (VersionProfile, error) {
	var profile VersionProfile
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT id, profile_name, version, settings_json, created_at, updated_at FROM batch_version_config_profiles WHERE batch_project_id = ?`, projectID).
		Scan(&profile.ID, &profile.Name, &profile.Version, &raw, &profile.CreatedAt, &profile.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) { return VersionProfile{ProjectID: projectID, Name: "默认版本配置档", Version: "v1"}, nil }
	if err != nil { return VersionProfile{}, err }
	profile.ProjectID = projectID
	if err := decodeSettings(raw, &profile.Settings); err != nil { return VersionProfile{}, err }
	return profile, nil
}

func (s *MySQLStore) SaveVersionProfile(ctx context.Context, profile VersionProfile) (VersionProfile, error) {
	raw, err := encode(profile.Settings)
	if err != nil { return VersionProfile{}, err }
	_, err = s.db.ExecContext(ctx, `INSERT INTO batch_version_config_profiles (batch_project_id, profile_name, version, settings_json) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE profile_name = VALUES(profile_name), version = VALUES(version), settings_json = VALUES(settings_json), updated_at = CURRENT_TIMESTAMP(6)`, profile.ProjectID, profile.Name, profile.Version, raw)
	if err != nil { return VersionProfile{}, err }
	return s.GetVersionProfile(ctx, profile.ProjectID)
}

func (s *MySQLStore) Get121Snapshot(ctx context.Context, projectID int64) (map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT b.source, b.platform_id FROM books b JOIN batch_projects p ON p.intake_id = b.intake_id WHERE p.id = ? ORDER BY b.source, b.platform_id`, projectID)
	if err != nil { return nil, err }
	defer rows.Close()
	items := make([]map[string]string, 0)
	for rows.Next() {
		var source, platformID string
		if err := rows.Scan(&source, &platformID); err != nil { return nil, err }
		items = append(items, map[string]string{"source": source, "platformId": platformID})
	}
	if err := rows.Err(); err != nil { return nil, err }
	return map[string]any{"sources": items}, nil
}

func (s *MySQLStore) GetStyleTypeSnapshot(ctx context.Context, projectID int64) (map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT b.genre, b.gender, b.style FROM books b JOIN batch_projects p ON p.intake_id = b.intake_id WHERE p.id = ?`, projectID)
	if err != nil { return nil, err }
	defer rows.Close()
	genres, genders, styles := map[string]struct{}{}, map[string]struct{}{}, map[string]struct{}{}
	for rows.Next() {
		var genre, gender, style string
		if err := rows.Scan(&genre, &gender, &style); err != nil { return nil, err }
		if genre != "" { genres[genre] = struct{}{} }
		if gender != "" { genders[gender] = struct{}{} }
		if style != "" { styles[style] = struct{}{} }
	}
	if err := rows.Err(); err != nil { return nil, err }
	keys := func(values map[string]struct{}) []string {
		out := make([]string, 0, len(values)); for value := range values { out = append(out, value) }; sort.Strings(out); return out
	}
	return map[string]any{"genres": keys(genres), "genders": keys(genders), "styles": keys(styles)}, nil
}
