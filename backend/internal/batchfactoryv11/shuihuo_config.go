package batchfactoryv11

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

func (s *MySQLStore) GetShuihuoConfig(ctx context.Context, owner string) (map[string]any, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT config_json FROM batch_factory_v12_shuihuo_configs WHERE owner_username=?`, owner).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	config := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &config); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func (s *MySQLStore) SaveShuihuoConfig(ctx context.Context, owner string, config map[string]any) error {
	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, `INSERT INTO batch_factory_v12_shuihuo_configs(owner_username,config_json,revision,created_at,updated_at) VALUES(?,?,1,?,?) ON DUPLICATE KEY UPDATE config_json=?,revision=batch_factory_v12_shuihuo_configs.revision+1,updated_at=?`, owner, string(raw), now, now, string(raw), now)
	return err
}

func (s *MemoryStore) GetShuihuoConfig(_ context.Context, owner string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	config := s.shuihuoConfigs[owner]
	if config == nil {
		return map[string]any{}, nil
	}
	return cloneShuihuoConfig(config), nil
}

func (s *MemoryStore) SaveShuihuoConfig(_ context.Context, owner string, config map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shuihuoConfigs[owner] = cloneShuihuoConfig(config)
	return nil
}

func cloneShuihuoConfig(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
