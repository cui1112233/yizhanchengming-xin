package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// MySQLAutomationPresetStore 是自动化预设的版本化实现：HEAD 表记录当前
// 修订号，revisions 表保存每次更新的不可变配置快照。批次通过
// automationPresetSnapshot（id+version）钉版，预设更新不再静默覆盖历史。
type MySQLAutomationPresetStore struct {
	db  *sql.DB
	now func() time.Time
}

// NewMySQLAutomationPresetStore 构造版本化预设库；legacyJSONPath 指向旧的
// JSON 文件存储，存在时做一次性导入（缺失的预设按其当前版本号落一条初始
// 修订），保证升级后用户已保存的预设不消失。
func NewMySQLAutomationPresetStore(db *sql.DB, legacyJSONPath string, now func() time.Time) (*MySQLAutomationPresetStore, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: mysql db is required", ErrAutomationPresetInvalid)
	}
	if now == nil {
		now = time.Now
	}
	store := &MySQLAutomationPresetStore{db: db, now: now}
	if err := store.importLegacyJSON(legacyJSONPath); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *MySQLAutomationPresetStore) importLegacyJSON(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read legacy automation presets: %w", err)
	}
	var legacy automationPresetFile
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return fmt.Errorf("parse legacy automation presets: %w", err)
	}
	for owner, items := range legacy.Presets {
		owner = strings.TrimSpace(owner)
		if owner == "" {
			continue
		}
		for _, item := range items {
			if strings.TrimSpace(item.ID) == "" || len(item.Config) == 0 {
				continue
			}
			var exists int
			err := s.db.QueryRow(`SELECT 1 FROM batch_factory_v12_automation_presets WHERE id=? AND owner_username=?`, item.ID, owner).Scan(&exists)
			if err == nil {
				continue
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("check legacy automation preset %s: %w", item.ID, err)
			}
			revision := item.Version
			if revision < 1 {
				revision = 1
			}
			name := strings.TrimSpace(item.Name)
			if name == "" {
				name = item.ID
			}
			createdAt := parseAutomationPresetTime(item.CreatedAt, s.now)
			updatedAt := parseAutomationPresetTime(item.UpdatedAt, s.now)
			encoded, err := json.Marshal(item.Config)
			if err != nil {
				return fmt.Errorf("encode legacy automation preset %s: %w", item.ID, err)
			}
			tx, err := s.db.Begin()
			if err != nil {
				return err
			}
			if _, err := tx.Exec(`INSERT INTO batch_factory_v12_automation_presets(id,owner_username,name,current_revision,created_at,updated_at) VALUES(?,?,?,?,?,?)`, item.ID, owner, name, revision, createdAt, updatedAt); err != nil {
				tx.Rollback()
				return fmt.Errorf("import legacy automation preset %s: %w", item.ID, err)
			}
			if _, err := tx.Exec(`INSERT INTO batch_factory_v12_automation_preset_revisions(preset_id,owner_username,revision,name,config_json,created_at) VALUES(?,?,?,?,?,?)`, item.ID, owner, revision, name, string(encoded), updatedAt); err != nil {
				tx.Rollback()
				return fmt.Errorf("import legacy automation preset %s revision: %w", item.ID, err)
			}
			if err := tx.Commit(); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseAutomationPresetTime(value string, now func() time.Time) time.Time {
	if parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(value)); err == nil {
		return parsed.UTC()
	}
	return now().UTC()
}

func (s *MySQLAutomationPresetStore) List(owner string) ([]AutomationPreset, error) {
	rows, err := s.db.Query(`SELECT id,name,current_revision,config_json,created_at,updated_at FROM batch_factory_v12_automation_presets p JOIN batch_factory_v12_automation_preset_revisions r ON r.preset_id=p.id AND r.revision=p.current_revision WHERE p.owner_username=?`, strings.TrimSpace(owner))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AutomationPreset{}
	for rows.Next() {
		item, err := scanAutomationPreset(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].UpdatedAt > items[j].UpdatedAt })
	return items, nil
}

type automationPresetScanner interface {
	Scan(...any) error
}

func scanAutomationPreset(row automationPresetScanner) (AutomationPreset, error) {
	var item AutomationPreset
	var config string
	var createdAt, updatedAt time.Time
	if err := row.Scan(&item.ID, &item.Name, &item.Version, &config, &createdAt, &updatedAt); err != nil {
		return AutomationPreset{}, err
	}
	parsed := map[string]any{}
	if err := json.Unmarshal([]byte(config), &parsed); err != nil {
		return AutomationPreset{}, fmt.Errorf("decode automation preset %s config: %w", item.ID, err)
	}
	item.Config = parsed
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339Nano)
	item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339Nano)
	return item, nil
}

func (s *MySQLAutomationPresetStore) Get(owner, id string) (*AutomationPreset, error) {
	row := s.db.QueryRow(`SELECT p.id,p.name,p.current_revision,r.config_json,p.created_at,p.updated_at FROM batch_factory_v12_automation_presets p JOIN batch_factory_v12_automation_preset_revisions r ON r.preset_id=p.id AND r.revision=p.current_revision WHERE p.owner_username=? AND p.id=?`, strings.TrimSpace(owner), strings.TrimSpace(id))
	item, err := scanAutomationPreset(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAutomationPresetNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// GetRevision 读取某个不可变历史修订，批次钉版回溯与升级 diff 用它。
func (s *MySQLAutomationPresetStore) GetRevision(owner, id string, revision int) (*AutomationPreset, error) {
	row := s.db.QueryRow(`SELECT r.preset_id,r.name,r.revision,r.config_json,r.created_at,r.created_at FROM batch_factory_v12_automation_preset_revisions r WHERE r.owner_username=? AND r.preset_id=? AND r.revision=?`, strings.TrimSpace(owner), strings.TrimSpace(id), revision)
	item, err := scanAutomationPreset(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAutomationPresetNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// History 按修订号升序返回预设的全部不可变版本。
func (s *MySQLAutomationPresetStore) History(owner, id string) ([]AutomationPreset, error) {
	rows, err := s.db.Query(`SELECT preset_id,name,revision,config_json,created_at,created_at FROM batch_factory_v12_automation_preset_revisions WHERE owner_username=? AND preset_id=? ORDER BY revision`, strings.TrimSpace(owner), strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AutomationPreset{}
	for rows.Next() {
		item, err := scanAutomationPreset(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrAutomationPresetNotFound
	}
	return items, nil
}

func (s *MySQLAutomationPresetStore) Create(owner, name string, config map[string]any) (AutomationPreset, error) {
	owner, name = strings.TrimSpace(owner), strings.TrimSpace(name)
	if owner == "" || name == "" || len(config) == 0 {
		return AutomationPreset{}, ErrAutomationPresetInvalid
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return AutomationPreset{}, err
	}
	now := s.now().UTC()
	item := AutomationPreset{ID: newAutomationPresetID(), Name: name, Version: 1, Config: cloneAutomationMap(config), CreatedAt: now.Format(time.RFC3339Nano), UpdatedAt: now.Format(time.RFC3339Nano)}
	tx, err := s.db.Begin()
	if err != nil {
		return AutomationPreset{}, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO batch_factory_v12_automation_presets(id,owner_username,name,current_revision,created_at,updated_at) VALUES(?,?,?,1,?,?)`, item.ID, owner, name, now, now); err != nil {
		return AutomationPreset{}, err
	}
	if _, err := tx.Exec(`INSERT INTO batch_factory_v12_automation_preset_revisions(preset_id,owner_username,revision,name,config_json,created_at) VALUES(?,?,1,?,?,?)`, item.ID, owner, name, string(encoded), now); err != nil {
		return AutomationPreset{}, err
	}
	if err := tx.Commit(); err != nil {
		return AutomationPreset{}, err
	}
	return item, nil
}

func (s *MySQLAutomationPresetStore) Update(owner, id string, expectedVersion int, name string, config map[string]any) (AutomationPreset, error) {
	owner, id = strings.TrimSpace(owner), strings.TrimSpace(id)
	tx, err := s.db.Begin()
	if err != nil {
		return AutomationPreset{}, err
	}
	defer tx.Rollback()
	var currentName string
	var currentRevision int
	var createdAt time.Time
	err = tx.QueryRow(`SELECT name,current_revision,created_at FROM batch_factory_v12_automation_presets WHERE owner_username=? AND id=? FOR UPDATE`, owner, id).Scan(&currentName, &currentRevision, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return AutomationPreset{}, ErrAutomationPresetNotFound
	}
	if err != nil {
		return AutomationPreset{}, err
	}
	if expectedVersion != currentRevision {
		return AutomationPreset{}, ErrAutomationPresetConflict
	}
	if strings.TrimSpace(name) == "" || len(config) == 0 {
		var currentConfig string
		if err := tx.QueryRow(`SELECT config_json FROM batch_factory_v12_automation_preset_revisions WHERE owner_username=? AND preset_id=? AND revision=?`, owner, id, currentRevision).Scan(&currentConfig); err != nil {
			return AutomationPreset{}, err
		}
		if strings.TrimSpace(name) == "" {
			name = currentName
		}
		if len(config) == 0 {
			if err := json.Unmarshal([]byte(currentConfig), &config); err != nil {
				return AutomationPreset{}, err
			}
		}
	}
	name = strings.TrimSpace(name)
	if name == "" || len(config) == 0 {
		return AutomationPreset{}, ErrAutomationPresetInvalid
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return AutomationPreset{}, err
	}
	now := s.now().UTC()
	next := currentRevision + 1
	if _, err := tx.Exec(`INSERT INTO batch_factory_v12_automation_preset_revisions(preset_id,owner_username,revision,name,config_json,created_at) VALUES(?,?,?,?,?,?)`, id, owner, next, name, string(encoded), now); err != nil {
		return AutomationPreset{}, err
	}
	if _, err := tx.Exec(`UPDATE batch_factory_v12_automation_presets SET name=?,current_revision=?,updated_at=? WHERE owner_username=? AND id=?`, name, next, now, owner, id); err != nil {
		return AutomationPreset{}, err
	}
	if err := tx.Commit(); err != nil {
		return AutomationPreset{}, err
	}
	return AutomationPreset{ID: id, Name: name, Version: next, Config: cloneAutomationMap(config), CreatedAt: createdAt.UTC().Format(time.RFC3339Nano), UpdatedAt: now.Format(time.RFC3339Nano)}, nil
}

func (s *MySQLAutomationPresetStore) Remove(owner, id string) error {
	owner, id = strings.TrimSpace(owner), strings.TrimSpace(id)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`DELETE FROM batch_factory_v12_automation_presets WHERE owner_username=? AND id=?`, owner, id)
	if err != nil {
		return err
	}
	if count, _ := res.RowsAffected(); count != 1 {
		return ErrAutomationPresetNotFound
	}
	if _, err := tx.Exec(`DELETE FROM batch_factory_v12_automation_preset_revisions WHERE owner_username=? AND preset_id=?`, owner, id); err != nil {
		return err
	}
	return tx.Commit()
}
