package screenplay

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var _ AssetCheckRepository = (*MemoryStore)(nil)
var _ AssetCheckRepository = (*MySQLStore)(nil)

func assetCheckBlocksNew(j AssetCheckJob) bool {
	return j.Status == "queued" || j.Status == "running" || j.Status == "unknown"
}

func validAssetCheckStatus(status string) bool {
	switch status {
	case "queued", "running", "completed", "cancelled", "failed", "unknown":
		return true
	}
	return false
}
func prepareAssetCheckJob(j AssetCheckJob) (AssetCheckJob, error) {
	if strings.TrimSpace(j.Owner) == "" || strings.TrimSpace(j.ProjectID) == "" || strings.TrimSpace(j.RequestKey) == "" || strings.TrimSpace(j.InputDigest) == "" {
		return j, ErrInvalidArgument
	}
	if j.ID == "" {
		j.ID = newScreenplayID("asset_check")
	}
	if j.Status == "" {
		j.Status = "queued"
	}
	if !validAssetCheckStatus(j.Status) {
		return j, ErrInvalidArgument
	}
	j.Revision = 1
	j.CreatedAt = time.Now().UTC()
	j.UpdatedAt = j.CreatedAt
	if _, err := json.Marshal(j); err != nil {
		return j, err
	}
	return j, nil
}
func preserveAssetCheckIdentity(j *AssetCheckJob, old AssetCheckJob) {
	j.ID = old.ID
	j.Owner = old.Owner
	j.ProjectID = old.ProjectID
	j.RequestKey = old.RequestKey
	j.InputDigest = old.InputDigest
	j.CreatedAt = old.CreatedAt
	j.Revision = old.Revision + 1
	j.UpdatedAt = time.Now().UTC()
}
func assetCheckApplyReady(j AssetCheckJob, revision int64) error {
	if j.Discarded {
		return fmt.Errorf("%w: 已放弃的检查结果不可添加", ErrInvalidState)
	}
	if j.Applied {
		return nil
	}
	if j.Revision != revision {
		return ErrRevisionConflict
	}
	if (j.Status != "completed" && j.Status != "cancelled") || len(j.Candidates) == 0 {
		return fmt.Errorf("%w: 没有可确认合并的检查草稿", ErrInvalidState)
	}
	return nil
}
func assetCheckNameKey(a Asset) string { return string(a.Type) + "\x00" + strings.TrimSpace(a.Name) }
func assetCheckNewAssets(j AssetCheckJob, existing []Asset) ([]Asset, int, error) {
	seen := map[string]bool{}
	order := 0
	for _, a := range existing {
		seen[assetCheckNameKey(a)] = true
		if a.SortOrder > order {
			order = a.SortOrder
		}
	}
	out := []Asset{}
	skipped := 0
	now := time.Now().UTC()
	for _, candidate := range j.Candidates {
		a := candidate.Asset
		a.Name = strings.TrimSpace(a.Name)
		if !validAssetType(a.Type) || a.Name == "" {
			return nil, 0, ErrInvalidArgument
		}
		key := assetCheckNameKey(a)
		if seen[key] {
			skipped++
			continue
		}
		seen[key] = true
		if len(candidate.Original) > 0 && !json.Valid(candidate.Original) {
			return nil, 0, ErrInvalidArgument
		}
		extra := map[string]string{}
		for k, v := range a.Extra {
			extra[k] = v
		}
		extra["original_asset_check"] = string(candidate.Original)
		a.Extra = extra
		a.ID = newScreenplayID("asset")
		a.OwnerUsername = j.Owner
		a.ProjectID = j.ProjectID
		if a.Degree == "" {
			a.Degree = j.Degree
		}
		if a.Degree == "" {
			a.Degree = DegreeStandard
		}
		order++
		a.SortOrder = order
		a.Revision = 1
		a.CreatedAt = now
		a.UpdatedAt = now
		out = append(out, a)
	}
	return out, skipped, nil
}
func (m *MemoryStore) CreateAssetCheckJob(_ context.Context, j AssetCheckJob) (AssetCheckJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.projects[projectKey(j.Owner, j.ProjectID)]; !ok {
		return AssetCheckJob{}, ErrNotFound
	}
	for _, old := range m.assetCheckJobs {
		if old.Owner == j.Owner && old.ProjectID == j.ProjectID && old.RequestKey == j.RequestKey {
			if old.InputDigest != j.InputDigest {
				return old, ErrRevisionConflict
			}
			return cloneAssetCheckJob(old), nil
		}
	}
	for _, old := range m.assetCheckJobs {
		if old.Owner == j.Owner && old.ProjectID == j.ProjectID && assetCheckBlocksNew(old) {
			return cloneAssetCheckJob(old), fmt.Errorf("%w: 请复用已有资产检查任务 %s", ErrInvalidState, old.ID)
		}
	}
	var err error
	j, err = prepareAssetCheckJob(j)
	if err != nil {
		return j, err
	}
	if _, exists := m.assetCheckJobs[j.ID]; exists {
		return AssetCheckJob{}, ErrRevisionConflict
	}
	if m.assetCheckJobs == nil {
		m.assetCheckJobs = map[string]AssetCheckJob{}
	}
	m.assetCheckJobs[j.ID] = cloneAssetCheckJob(j)
	return cloneAssetCheckJob(j), nil
}
func (m *MemoryStore) GetAssetCheckJob(_ context.Context, o, p, id string) (AssetCheckJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.assetCheckJobs[id]
	if !ok || j.Owner != o || j.ProjectID != p {
		return AssetCheckJob{}, ErrNotFound
	}
	return cloneAssetCheckJob(j), nil
}
func (m *MemoryStore) ListAssetCheckJobs(_ context.Context, o, p string, limit int) ([]AssetCheckJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []AssetCheckJob{}
	for _, j := range m.assetCheckJobs {
		if j.Owner == o && j.ProjectID == p {
			out = append(out, cloneAssetCheckJob(j))
		}
	}
	sort.Slice(out, func(i, k int) bool { return out[i].UpdatedAt.After(out[k].UpdatedAt) })
	limit = assetCheckLimit(limit)
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (m *MemoryStore) ListRunnableAssetCheckJobs(_ context.Context) ([]AssetCheckJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []AssetCheckJob{}
	for _, j := range m.assetCheckJobs {
		if j.Status == "queued" || j.Status == "running" {
			out = append(out, cloneAssetCheckJob(j))
		}
	}
	sort.Slice(out, func(i, k int) bool { return out[i].UpdatedAt.Before(out[k].UpdatedAt) })
	if len(out) > 256 {
		out = out[:256]
	}
	return out, nil
}
func (m *MemoryStore) MutateAssetCheckJob(_ context.Context, o, p, id string, fn func(*AssetCheckJob) error) (AssetCheckJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	old, ok := m.assetCheckJobs[id]
	if !ok || old.Owner != o || old.ProjectID != p {
		return AssetCheckJob{}, ErrNotFound
	}
	j := cloneAssetCheckJob(old)
	if err := fn(&j); err != nil {
		return AssetCheckJob{}, err
	}
	preserveAssetCheckIdentity(&j, old)
	if !validAssetCheckStatus(j.Status) {
		return AssetCheckJob{}, ErrInvalidArgument
	}
	if _, err := json.Marshal(j); err != nil {
		return AssetCheckJob{}, err
	}
	m.assetCheckJobs[id] = cloneAssetCheckJob(j)
	return cloneAssetCheckJob(j), nil
}
func (m *MemoryStore) ApplyAssetCheckJob(_ context.Context, o, p, id string, revision int64) (AssetCheckJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.projects[projectKey(o, p)]; !ok {
		return AssetCheckJob{}, ErrNotFound
	}
	old, ok := m.assetCheckJobs[id]
	if !ok || old.Owner != o || old.ProjectID != p {
		return AssetCheckJob{}, ErrNotFound
	}
	if err := assetCheckApplyReady(old, revision); err != nil {
		return old, err
	}
	if old.Applied {
		return cloneAssetCheckJob(old), nil
	}
	assets, skipped, err := assetCheckNewAssets(old, m.assets[projectKey(o, p)])
	if err != nil {
		return AssetCheckJob{}, err
	}
	j := cloneAssetCheckJob(old)
	j.Applied = true
	j.AddedCount = len(assets)
	j.SkippedCount = skipped
	preserveAssetCheckIdentity(&j, old)
	if m.assets == nil {
		m.assets = map[string][]Asset{}
	}
	m.assets[projectKey(o, p)] = append(m.assets[projectKey(o, p)], assets...)
	m.assetCheckJobs[id] = cloneAssetCheckJob(j)
	return cloneAssetCheckJob(j), nil
}
func assetCheckLimit(n int) int {
	if n <= 0 {
		return 30
	}
	if n > 200 {
		return 200
	}
	return n
}
func scanAssetCheck(scanner interface{ Scan(...any) error }) (AssetCheckJob, error) {
	var raw []byte
	err := scanner.Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return AssetCheckJob{}, ErrNotFound
	}
	if err != nil {
		return AssetCheckJob{}, err
	}
	var j AssetCheckJob
	err = json.Unmarshal(raw, &j)
	return j, err
}
func lockAssetCheckProject(ctx context.Context, tx *sql.Tx, o, p string) error {
	var id string
	err := tx.QueryRowContext(ctx, `SELECT id FROM screenplay_projects WHERE id=? AND owner_username=? AND deleted_at IS NULL FOR UPDATE`, p, o).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
func saveAssetCheckTx(ctx context.Context, tx *sql.Tx, j AssetCheckJob) error {
	if !validAssetCheckStatus(j.Status) {
		return ErrInvalidArgument
	}
	raw, err := json.Marshal(j)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE screenplay_asset_check_jobs SET status=?,applied=?,candidate_count=?,state_json=?,updated_at=? WHERE id=? AND owner_username=? AND project_id=?`, j.Status, j.Applied, len(j.Candidates), string(raw), j.UpdatedAt, j.ID, j.Owner, j.ProjectID)
	return err
}
func (s *MySQLStore) CreateAssetCheckJob(ctx context.Context, j AssetCheckJob) (AssetCheckJob, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AssetCheckJob{}, err
	}
	defer tx.Rollback()
	if err = lockAssetCheckProject(ctx, tx, j.Owner, j.ProjectID); err != nil {
		return AssetCheckJob{}, err
	}
	old, err := scanAssetCheck(tx.QueryRowContext(ctx, `SELECT state_json FROM screenplay_asset_check_jobs WHERE owner_username=? AND project_id=? AND request_key=? FOR UPDATE`, j.Owner, j.ProjectID, j.RequestKey))
	if err == nil {
		if old.InputDigest != j.InputDigest {
			return old, ErrRevisionConflict
		}
		return old, tx.Commit()
	}
	if !errors.Is(err, ErrNotFound) {
		return AssetCheckJob{}, err
	}
	var active string
	err = tx.QueryRowContext(ctx, `SELECT id FROM screenplay_asset_check_jobs WHERE owner_username=? AND project_id=? AND status IN ('queued','running','unknown') LIMIT 1 FOR UPDATE`, j.Owner, j.ProjectID).Scan(&active)
	if err == nil {
		return AssetCheckJob{}, fmt.Errorf("%w: 请复用已有资产检查任务 %s", ErrInvalidState, active)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return AssetCheckJob{}, err
	}
	j, err = prepareAssetCheckJob(j)
	if err != nil {
		return j, err
	}
	raw, err := json.Marshal(j)
	if err != nil {
		return j, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO screenplay_asset_check_jobs(id,owner_username,project_id,request_key,input_digest,status,applied,candidate_count,state_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, j.ID, j.Owner, j.ProjectID, j.RequestKey, j.InputDigest, j.Status, j.Applied, len(j.Candidates), string(raw), j.CreatedAt, j.UpdatedAt)
	if err != nil {
		return AssetCheckJob{}, err
	}
	if err = tx.Commit(); err != nil {
		return AssetCheckJob{}, err
	}
	return j, nil
}
func (s *MySQLStore) GetAssetCheckJob(ctx context.Context, o, p, id string) (AssetCheckJob, error) {
	return scanAssetCheck(s.db.QueryRowContext(ctx, `SELECT state_json FROM screenplay_asset_check_jobs WHERE id=? AND owner_username=? AND project_id=?`, id, o, p))
}
func (s *MySQLStore) listAssetChecks(ctx context.Context, query string, args ...any) ([]AssetCheckJob, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AssetCheckJob{}
	for rows.Next() {
		j, err := scanAssetCheck(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
func (s *MySQLStore) ListAssetCheckJobs(ctx context.Context, o, p string, limit int) ([]AssetCheckJob, error) {
	return s.listAssetChecks(ctx, `SELECT state_json FROM screenplay_asset_check_jobs WHERE owner_username=? AND project_id=? ORDER BY updated_at DESC LIMIT ?`, o, p, assetCheckLimit(limit))
}
func (s *MySQLStore) ListRunnableAssetCheckJobs(ctx context.Context) ([]AssetCheckJob, error) {
	return s.listAssetChecks(ctx, `SELECT state_json FROM screenplay_asset_check_jobs WHERE status IN ('queued','running') ORDER BY updated_at LIMIT 256`)
}
func (s *MySQLStore) MutateAssetCheckJob(ctx context.Context, o, p, id string, fn func(*AssetCheckJob) error) (AssetCheckJob, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AssetCheckJob{}, err
	}
	defer tx.Rollback()
	old, err := scanAssetCheck(tx.QueryRowContext(ctx, `SELECT state_json FROM screenplay_asset_check_jobs WHERE id=? AND owner_username=? AND project_id=? FOR UPDATE`, id, o, p))
	if err != nil {
		return AssetCheckJob{}, err
	}
	j := cloneAssetCheckJob(old)
	if err = fn(&j); err != nil {
		return AssetCheckJob{}, err
	}
	preserveAssetCheckIdentity(&j, old)
	if err = saveAssetCheckTx(ctx, tx, j); err != nil {
		return AssetCheckJob{}, err
	}
	if err = tx.Commit(); err != nil {
		return AssetCheckJob{}, err
	}
	return j, nil
}
func (s *MySQLStore) ApplyAssetCheckJob(ctx context.Context, o, p, id string, revision int64) (AssetCheckJob, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AssetCheckJob{}, err
	}
	defer tx.Rollback()
	if err = lockAssetCheckProject(ctx, tx, o, p); err != nil {
		return AssetCheckJob{}, err
	}
	old, err := scanAssetCheck(tx.QueryRowContext(ctx, `SELECT state_json FROM screenplay_asset_check_jobs WHERE id=? AND owner_username=? AND project_id=? FOR UPDATE`, id, o, p))
	if err != nil {
		return AssetCheckJob{}, err
	}
	if err = assetCheckApplyReady(old, revision); err != nil {
		return old, err
	}
	if old.Applied {
		return old, tx.Commit()
	}
	rows, err := tx.QueryContext(ctx, `SELECT asset_type,name,sort_order FROM screenplay_assets WHERE owner_username=? AND project_id=?`, o, p)
	if err != nil {
		return AssetCheckJob{}, err
	}
	existing := []Asset{}
	for rows.Next() {
		var a Asset
		if err = rows.Scan(&a.Type, &a.Name, &a.SortOrder); err != nil {
			rows.Close()
			return AssetCheckJob{}, err
		}
		existing = append(existing, a)
	}
	err = rows.Err()
	closeErr := rows.Close()
	if err != nil {
		return AssetCheckJob{}, err
	}
	if closeErr != nil {
		return AssetCheckJob{}, closeErr
	}
	assets, skipped, err := assetCheckNewAssets(old, existing)
	if err != nil {
		return AssetCheckJob{}, err
	}
	for _, a := range assets {
		extra, err := json.Marshal(a.Extra)
		if err != nil {
			return AssetCheckJob{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO screenplay_assets (id,project_id,owner_username,asset_type,degree,name,identity_text,description,prompt,voice_text,episodes,tag,extra,sort_order,source_batch,revision,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1,?,?)`, a.ID, p, o, string(a.Type), string(a.Degree), a.Name, nullString(a.Identity), nullString(a.Description), nullString(a.Prompt), nullString(a.Voice), a.Episodes, a.Tag, string(extra), a.SortOrder, a.SourceBatch, a.CreatedAt, a.UpdatedAt)
		if err != nil {
			return AssetCheckJob{}, err
		}
	}
	j := cloneAssetCheckJob(old)
	j.Applied = true
	j.AddedCount = len(assets)
	j.SkippedCount = skipped
	preserveAssetCheckIdentity(&j, old)
	if err = saveAssetCheckTx(ctx, tx, j); err != nil {
		return AssetCheckJob{}, err
	}
	if err = tx.Commit(); err != nil {
		return AssetCheckJob{}, err
	}
	return j, nil
}

func (m *MemoryStore) FindAssetCheckJobByRequestKey(_ context.Context, owner, projectID, key string) (AssetCheckJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, job := range m.assetCheckJobs {
		if job.Owner == owner && job.ProjectID == projectID && job.RequestKey == key {
			return cloneAssetCheckJob(job), nil
		}
	}
	return AssetCheckJob{}, ErrNotFound
}
func (s *MySQLStore) FindAssetCheckJobByRequestKey(ctx context.Context, owner, projectID, key string) (AssetCheckJob, error) {
	return scanAssetCheck(s.db.QueryRowContext(ctx, `SELECT state_json FROM screenplay_asset_check_jobs WHERE owner_username=? AND project_id=? AND request_key=?`, owner, projectID, key))
}
