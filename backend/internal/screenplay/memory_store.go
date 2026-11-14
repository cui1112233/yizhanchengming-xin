package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryStore 是进程内 Store 实现，用于单测与本地串联。
// 它刻意复刻 MySQL 实现的两个并发语义：revision 冲突与显式互斥锁。
type MemoryStore struct {
	storyboardGenerations map[string]StoryboardGenerationState
	storyboardPreferences map[string]StoryboardPreferences
	mu                    sync.Mutex
	projects              map[string]Project
	userStyles            map[string]UserStyle
	documents             map[string][]Document
	shots                 map[string][]Shot
	characters            map[string][]Character
	scenes                map[string][]Scene
	assets                map[string][]Asset
	assetBatches          map[string][]AssetBatch
	storyboards           map[string][]Storyboard
	batches               map[string]*Batch
	logs                  map[string][]BatchLog
	locks                 map[string]memoryLock
	revision              int64
}

type memoryLock struct {
	token   string
	expires time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects:     map[string]Project{},
		documents:    map[string][]Document{},
		shots:        map[string][]Shot{},
		characters:   map[string][]Character{},
		scenes:       map[string][]Scene{},
		assets:       map[string][]Asset{},
		assetBatches: map[string][]AssetBatch{},
		storyboards:  map[string][]Storyboard{},
		batches:      map[string]*Batch{},
		logs:         map[string][]BatchLog{},
		locks:        map[string]memoryLock{},
	}
}

func documentKey(owner, projectID, documentID string) string {
	return owner + "\x00" + projectID + "\x00" + documentID
}

func (m *MemoryStore) SaveScreenplay(_ context.Context, doc Document, shots []Shot, characters []Character, scenes []Scene) (Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(doc.OwnerUsername, doc.ProjectID)
	if _, ok := m.projects[key]; !ok {
		return Document{}, fmt.Errorf("%w: 项目 %s", ErrNotFound, doc.ProjectID)
	}
	existing := m.documents[key]
	maxVersion := 0
	for i := range existing {
		if existing[i].Version > maxVersion {
			maxVersion = existing[i].Version
		}
		existing[i].Status = "superseded"
	}
	now := time.Now().UTC()
	doc.Version = maxVersion + 1
	doc.Status = "active"
	doc.ContentSHA256 = contentDigest(doc.Content)
	doc.Revision = 1
	doc.CreatedAt = now
	doc.UpdatedAt = now
	doc.EpisodeCount = len(uniqueEpisodes(shots))
	m.documents[key] = append(existing, doc)

	docKey := documentKey(doc.OwnerUsername, doc.ProjectID, doc.ID)
	m.shots[docKey] = append([]Shot(nil), shots...)
	m.characters[docKey] = append([]Character(nil), characters...)
	m.scenes[docKey] = append([]Scene(nil), scenes...)
	return doc, nil
}

func (m *MemoryStore) ActiveDocument(_ context.Context, owner, projectID string) (Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	documents := m.documents[projectKey(owner, projectID)]
	for i := len(documents) - 1; i >= 0; i-- {
		if documents[i].Status == "active" {
			return documents[i], nil
		}
	}
	return Document{}, ErrNoActiveScript
}

func (m *MemoryStore) GetDocument(_ context.Context, owner, projectID, documentID string) (Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, document := range m.documents[projectKey(owner, projectID)] {
		if document.ID == documentID {
			return document, nil
		}
	}
	return Document{}, fmt.Errorf("%w: 剧本版本 %s", ErrNotFound, documentID)
}

func (m *MemoryStore) ListDocuments(_ context.Context, owner, projectID string) ([]Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := m.documents[projectKey(owner, projectID)]
	out := make([]Document, 0, len(stored))
	for i := len(stored) - 1; i >= 0; i-- {
		doc := stored[i]
		doc.Content = ""
		out = append(out, doc)
	}
	return out, nil
}

func (m *MemoryStore) ListShots(_ context.Context, owner, projectID, documentID string) ([]Shot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := m.shots[documentKey(owner, projectID, documentID)]
	out := make([]Shot, len(stored))
	copy(out, stored)
	return out, nil
}

func (m *MemoryStore) CreateShot(_ context.Context, shot Shot) (Shot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if strings.TrimSpace(shot.OwnerUsername) == "" || strings.TrimSpace(shot.ProjectID) == "" || strings.TrimSpace(shot.DocumentID) == "" {
		return Shot{}, fmt.Errorf("%w: 项目、剧本版本与用户不能为空", ErrInvalidArgument)
	}
	if strings.TrimSpace(shot.Heading) == "" && strings.TrimSpace(shot.ScriptText) == "" {
		return Shot{}, fmt.Errorf("%w: 场次标题和内容不能同时为空", ErrInvalidArgument)
	}
	if shot.EpisodeNo <= 0 {
		shot.EpisodeNo = 1
	}
	if shot.ID == "" {
		shot.ID = newScreenplayID("shot")
	}
	key := documentKey(shot.OwnerUsername, shot.ProjectID, shot.DocumentID)
	if shot.SortOrder <= 0 {
		for _, existing := range m.shots[key] {
			if existing.EpisodeNo == shot.EpisodeNo && existing.SortOrder >= shot.SortOrder {
				shot.SortOrder = existing.SortOrder + 1
			}
		}
	}
	m.shots[key] = append(m.shots[key], shot)
	sort.SliceStable(m.shots[key], func(i, j int) bool {
		if m.shots[key][i].EpisodeNo != m.shots[key][j].EpisodeNo {
			return m.shots[key][i].EpisodeNo < m.shots[key][j].EpisodeNo
		}
		return m.shots[key][i].SortOrder < m.shots[key][j].SortOrder
	})
	return shot, nil
}

func (m *MemoryStore) UpdateShot(_ context.Context, owner, projectID, shotID string, changes Shot) (Shot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, rows := range m.shots {
		for index := range rows {
			if rows[index].ID != shotID || rows[index].OwnerUsername != owner || rows[index].ProjectID != projectID {
				continue
			}
			if strings.TrimSpace(changes.Heading) == "" && strings.TrimSpace(changes.ScriptText) == "" {
				return Shot{}, fmt.Errorf("%w: 场次标题和内容不能同时为空", ErrInvalidArgument)
			}
			if changes.EpisodeNo > 0 {
				rows[index].EpisodeNo = changes.EpisodeNo
			}
			rows[index].Heading = strings.TrimSpace(changes.Heading)
			rows[index].ScriptText = strings.TrimSpace(changes.ScriptText)
			rows[index].CharacterNames = append([]string(nil), changes.CharacterNames...)
			rows[index].SceneName = strings.TrimSpace(changes.SceneName)
			rows[index].PropNames = append([]string(nil), changes.PropNames...)
			rows[index].LocationReference = strings.TrimSpace(changes.LocationReference)
			m.shots[key] = rows
			return rows[index], nil
		}
	}
	return Shot{}, fmt.Errorf("%w: 镜头 %s", ErrNotFound, shotID)
}

func (m *MemoryStore) MoveShot(_ context.Context, owner, projectID, shotID string, direction int) ([]Shot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if direction != -1 && direction != 1 {
		return nil, fmt.Errorf("%w: 移动方向无效", ErrInvalidArgument)
	}
	for key, rows := range m.shots {
		index := -1
		for i := range rows {
			if rows[i].ID == shotID && rows[i].OwnerUsername == owner && rows[i].ProjectID == projectID {
				index = i
				break
			}
		}
		if index < 0 {
			continue
		}
		target := index + direction
		if target < 0 || target >= len(rows) || rows[target].EpisodeNo != rows[index].EpisodeNo {
			return nil, fmt.Errorf("%w: 已到当前集边界", ErrInvalidState)
		}
		rows[index].SortOrder, rows[target].SortOrder = rows[target].SortOrder, rows[index].SortOrder
		rows[index], rows[target] = rows[target], rows[index]
		m.shots[key] = rows
		out := make([]Shot, len(rows))
		copy(out, rows)
		return out, nil
	}
	return nil, fmt.Errorf("%w: 镜头 %s", ErrNotFound, shotID)
}

func (m *MemoryStore) DeleteShot(_ context.Context, owner, projectID, shotID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for key, rows := range m.shots {
		for index := range rows {
			if rows[index].ID != shotID || rows[index].OwnerUsername != owner || rows[index].ProjectID != projectID {
				continue
			}
			m.shots[key] = append(rows[:index], rows[index+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%w: 镜头 %s", ErrNotFound, shotID)
}

func (m *MemoryStore) ListCharacters(_ context.Context, owner, projectID, documentID string) ([]Character, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := m.characters[documentKey(owner, projectID, documentID)]
	out := make([]Character, len(stored))
	copy(out, stored)
	return out, nil
}

func (m *MemoryStore) ListScenes(_ context.Context, owner, projectID, documentID string) ([]Scene, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := m.scenes[documentKey(owner, projectID, documentID)]
	out := make([]Scene, len(stored))
	copy(out, stored)
	return out, nil
}

func projectKey(owner, id string) string { return owner + "\x00" + id }

func (m *MemoryStore) CreateProject(_ context.Context, project Project) (Project, error) {
	if project.AspectRatio != "" && !ValidCreativeAspectRatio(project.AspectRatio) {
		return Project{}, fmt.Errorf("%w: 不支持的剧本比例%s", ErrInvalidArgument, project.AspectRatio)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	project.ID = strings.TrimSpace(project.ID)
	project.OwnerUsername = strings.TrimSpace(project.OwnerUsername)
	project.Title = strings.TrimSpace(project.Title)
	if project.ID == "" || project.OwnerUsername == "" {
		return Project{}, fmt.Errorf("%w: 项目 ID 与 owner 不能为空", ErrInvalidArgument)
	}
	if project.Title == "" {
		project.Title = "未命名漫剧"
	}
	if project.AspectRatio == "" {
		project.AspectRatio = "9:16"
	}
	if project.ScriptMode == "" {
		project.ScriptMode = "b"
	}
	if project.PromptID == "" {
		project.PromptID = "b_director_screenplay"
	}
	if project.AssetDegree == "" {
		project.AssetDegree = DegreeMaster
	}
	if project.PermissionMode == "" {
		project.PermissionMode = PermissionAuto
	}
	now := time.Now().UTC()
	if project.CreatedAt.IsZero() {
		project.CreatedAt = now
	}
	project.UpdatedAt = now
	key := projectKey(project.OwnerUsername, project.ID)
	if _, exists := m.projects[key]; exists {
		return Project{}, fmt.Errorf("%w: 项目已存在", ErrInvalidArgument)
	}
	m.projects[key] = project
	return project, nil
}

func (m *MemoryStore) ListProjects(_ context.Context, owner string) ([]Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Project, 0)
	prefix := strings.TrimSpace(owner) + "\x00"
	for key, project := range m.projects {
		if strings.HasPrefix(key, prefix) {
			out = append(out, project)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (m *MemoryStore) UpdateProject(_ context.Context, owner, id string, changes Project, expectedRevision int64) (Project, error) {
	if changes.AspectRatio != "" && !ValidCreativeAspectRatio(changes.AspectRatio) {
		return Project{}, fmt.Errorf("%w: 不支持的剧本比例%s", ErrInvalidArgument, changes.AspectRatio)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(owner, id)
	current, ok := m.projects[key]
	if !ok {
		return Project{}, fmt.Errorf("%w: 项目 %s", ErrNotFound, id)
	}
	if current.Revision != expectedRevision {
		return Project{}, ErrRevisionConflict
	}
	if title := strings.TrimSpace(changes.Title); title != "" {
		current.Title = title
	}
	if changes.ScriptMode != "" {
		current.ScriptMode = changes.ScriptMode
	}
	if changes.AspectRatio != "" {
		current.AspectRatio = changes.AspectRatio
	}
	if changes.ArtStyle != "" {
		current.ArtStyle = changes.ArtStyle
	}
	current.StyleDesc = changes.StyleDesc
	current.SelectedStyle = changes.SelectedStyle
	current.Synopsis = changes.Synopsis
	if changes.ModelID != "" {
		current.ModelID = changes.ModelID
	}
	if changes.PromptID != "" {
		current.PromptID = changes.PromptID
	}
	if changes.AssetDegree != "" {
		current.AssetDegree = changes.AssetDegree
	}
	if changes.PermissionMode != "" {
		current.PermissionMode = changes.PermissionMode
	}
	current.Revision++
	current.UpdatedAt = time.Now().UTC()
	m.projects[key] = current
	return current, nil
}

func (m *MemoryStore) SaveProjectSource(_ context.Context, owner, id, content string) (Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(owner, id)
	current, ok := m.projects[key]
	if !ok {
		return Project{}, fmt.Errorf("%w: 项目 %s", ErrNotFound, id)
	}
	current.SourceText = content
	current.UpdatedAt = time.Now().UTC()
	m.projects[key] = current
	return current, nil
}

func (m *MemoryStore) DeleteProject(_ context.Context, owner, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(owner, id)
	if _, ok := m.projects[key]; !ok {
		return fmt.Errorf("%w: 项目 %s", ErrNotFound, id)
	}
	delete(m.projects, key)
	delete(m.documents, key)
	delete(m.assets, key)
	delete(m.storyboards, key)
	delete(m.batches, key)
	return nil
}

func (m *MemoryStore) GetProject(_ context.Context, owner, id string) (Project, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	project, ok := m.projects[projectKey(owner, id)]
	if !ok {
		return Project{}, fmt.Errorf("%w: 项目 %s", ErrNotFound, id)
	}
	project.CreativeConfig = cloneCreativeConfig(project.CreativeConfig)
	return project, nil
}

func (m *MemoryStore) CreateBatch(_ context.Context, batch *Batch) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(batch.OwnerUsername, batch.ProjectID)
	if _, exists := m.batches[key]; exists {
		return fmt.Errorf("%w: 项目已存在批次", ErrInvalidArgument)
	}
	m.batches[key] = cloneBatch(batch)
	return nil
}

func (m *MemoryStore) GetBatch(_ context.Context, owner, projectID string) (*Batch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	batch, ok := m.batches[projectKey(owner, projectID)]
	if !ok {
		return nil, fmt.Errorf("%w: 项目批次", ErrNotFound)
	}
	return cloneBatch(batch), nil
}

func (m *MemoryStore) UpdateBatch(_ context.Context, batch *Batch, expectedRevision int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(batch.OwnerUsername, batch.ProjectID)
	stored, ok := m.batches[key]
	if !ok {
		return fmt.Errorf("%w: 项目批次", ErrNotFound)
	}
	if stored.Revision != expectedRevision {
		return ErrRevisionConflict
	}
	next := cloneBatch(batch)
	next.Revision = expectedRevision + 1
	m.batches[key] = next
	batch.Revision = next.Revision
	return nil
}

func (m *MemoryStore) MutateBatch(_ context.Context, owner, projectID string, fn func(*Batch) error) (*Batch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(owner, projectID)
	stored, ok := m.batches[key]
	if !ok {
		return nil, fmt.Errorf("%w: 项目批次", ErrNotFound)
	}
	working := cloneBatch(stored)
	if err := fn(working); err != nil {
		return nil, err
	}
	working.Revision = stored.Revision + 1
	m.batches[key] = cloneBatch(working)
	return working, nil
}

func (m *MemoryStore) AppendBatchLog(_ context.Context, batchID, owner, level, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revision++
	m.logs[owner+"\x00"+batchID] = append(m.logs[owner+"\x00"+batchID], BatchLog{
		ID: m.revision, Level: level, Message: message, CreatedAt: time.Now().UTC(),
	})
	return nil
}

func (m *MemoryStore) ListBatchLogs(_ context.Context, batchID, owner string, limit int) ([]BatchLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entries := m.logs[owner+"\x00"+batchID]
	if limit <= 0 || limit > len(entries) {
		limit = len(entries)
	}
	out := make([]BatchLog, 0, limit)
	for i := len(entries) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, entries[i])
	}
	return out, nil
}

func (m *MemoryStore) AcquireLock(_ context.Context, name, token string, ttl time.Duration) error {
	if name == "" || token == "" {
		return fmt.Errorf("%w: 锁名与 token 不能为空", ErrInvalidArgument)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	lock, exists := m.locks[name]
	if exists && lock.expires.After(now) && lock.token != token {
		return fmt.Errorf("%w: %s", ErrLockBusy, name)
	}
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	m.locks[name] = memoryLock{token: token, expires: now.Add(ttl)}
	return nil
}

func (m *MemoryStore) ReleaseLock(_ context.Context, name, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if lock, ok := m.locks[name]; ok && lock.token == token {
		delete(m.locks, name)
	}
	return nil
}

func (m *MemoryStore) SaveAssetBatch(_ context.Context, batch AssetBatch) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(batch.OwnerUsername, batch.ProjectID)
	if m.assetBatches == nil {
		m.assetBatches = map[string][]AssetBatch{}
	}
	stored := m.assetBatches[key]
	for i := range stored {
		if stored[i].Type == batch.Type && stored[i].Degree == batch.Degree && stored[i].BatchIndex == batch.BatchIndex {
			stored[i] = batch
			m.assetBatches[key] = stored
			return nil
		}
	}
	m.assetBatches[key] = append(stored, batch)
	return nil
}

func (m *MemoryStore) MergeAssets(_ context.Context, owner, projectID string, assets []Asset) ([]Asset, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.assets == nil {
		m.assets = map[string][]Asset{}
	}
	key := projectKey(owner, projectID)
	merged := m.assets[key]
	index := map[string]int{}
	for i, asset := range merged {
		index[string(asset.Type)+"\x00"+asset.Name] = i
	}
	for i := range assets {
		assets[i].ProjectID = projectID
		assets[i].OwnerUsername = owner
		assets[i].SortOrder = i + 1
		if position, ok := index[string(assets[i].Type)+"\x00"+assets[i].Name]; ok {
			merged[position] = assets[i]
			continue
		}
		index[string(assets[i].Type)+"\x00"+assets[i].Name] = len(merged)
		merged = append(merged, assets[i])
	}
	m.assets[key] = merged
	return assets, nil
}

func (m *MemoryStore) ListAssets(_ context.Context, owner, projectID string, assetType AssetType) ([]Asset, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := m.assets[projectKey(owner, projectID)]
	out := make([]Asset, 0, len(stored))
	for _, asset := range stored {
		if assetType == "" || asset.Type == assetType {
			out = append(out, asset)
		}
	}
	return out, nil
}

func (m *MemoryStore) CreateAsset(_ context.Context, asset Asset) (Asset, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	asset.OwnerUsername = strings.TrimSpace(asset.OwnerUsername)
	asset.ProjectID = strings.TrimSpace(asset.ProjectID)
	asset.Name = strings.TrimSpace(asset.Name)
	if _, ok := m.projects[projectKey(asset.OwnerUsername, asset.ProjectID)]; !ok {
		return Asset{}, fmt.Errorf("%w: 项目 %s", ErrNotFound, asset.ProjectID)
	}
	if !validAssetType(asset.Type) || asset.Name == "" {
		return Asset{}, fmt.Errorf("%w: 资产类型和名称不能为空", ErrInvalidArgument)
	}
	if asset.Degree == "" {
		asset.Degree = DegreeStandard
	}
	key := projectKey(asset.OwnerUsername, asset.ProjectID)
	for _, existing := range m.assets[key] {
		if existing.Type == asset.Type && strings.EqualFold(existing.Name, asset.Name) {
			return Asset{}, fmt.Errorf("%w: 同类型资产名称已存在", ErrInvalidArgument)
		}
	}
	if asset.ID == "" {
		asset.ID = newScreenplayID("asset")
	}
	asset.SortOrder = len(m.assets[key]) + 1
	asset.Revision = 1
	asset.CreatedAt = time.Now().UTC()
	asset.UpdatedAt = asset.CreatedAt
	m.assets[key] = append(m.assets[key], asset)
	return asset, nil
}

func (m *MemoryStore) UpdateAsset(_ context.Context, owner, projectID, assetID string, changes Asset, expectedRevision int64) (Asset, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(owner, projectID)
	rows := m.assets[key]
	index := -1
	for i := range rows {
		if rows[i].ID == assetID {
			index = i
			break
		}
	}
	if index < 0 {
		return Asset{}, fmt.Errorf("%w: 资产 %s", ErrNotFound, assetID)
	}
	if rows[index].Revision != expectedRevision {
		return Asset{}, ErrRevisionConflict
	}
	changes.Name = strings.TrimSpace(changes.Name)
	if !validAssetType(changes.Type) || changes.Name == "" {
		return Asset{}, fmt.Errorf("%w: 资产类型和名称不能为空", ErrInvalidArgument)
	}
	for i, existing := range rows {
		if i != index && existing.Type == changes.Type && strings.EqualFold(existing.Name, changes.Name) {
			return Asset{}, fmt.Errorf("%w: 同类型资产名称已存在", ErrInvalidArgument)
		}
	}
	current := rows[index]
	current.Type = changes.Type
	current.Name = changes.Name
	current.Identity = strings.TrimSpace(changes.Identity)
	current.Description = strings.TrimSpace(changes.Description)
	current.Prompt = strings.TrimSpace(changes.Prompt)
	current.Voice = strings.TrimSpace(changes.Voice)
	current.Episodes = strings.TrimSpace(changes.Episodes)
	current.Tag = strings.TrimSpace(changes.Tag)
	if changes.Degree != "" {
		current.Degree = changes.Degree
	}
	current.Extra = changes.Extra
	current.Revision++
	current.UpdatedAt = time.Now().UTC()
	rows[index] = current
	m.assets[key] = rows
	return current, nil
}

func (m *MemoryStore) DeleteAsset(_ context.Context, owner, projectID, assetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := projectKey(owner, projectID)
	rows := m.assets[key]
	for index := range rows {
		if rows[index].ID == assetID {
			m.assets[key] = append(rows[:index], rows[index+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%w: 资产 %s", ErrNotFound, assetID)
}

func (m *MemoryStore) SaveStoryboard(_ context.Context, storyboard Storyboard) (Storyboard, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveStoryboardLocked(storyboard)
}
func (m *MemoryStore) saveStoryboardLocked(storyboard Storyboard) (Storyboard, error) {
	key := projectKey(storyboard.OwnerUsername, storyboard.ProjectID)
	if m.storyboards == nil {
		m.storyboards = map[string][]Storyboard{}
	}
	stored := m.storyboards[key]
	now := time.Now().UTC()
	storyboard.ShotCount = len(storyboard.Shots)
	if storyboard.Status == "" {
		storyboard.Status = "done"
	}
	for index := range storyboard.Shots {
		shot := &storyboard.Shots[index]
		if shot.ID == "" {
			shot.ID = newScreenplayID("sb_shot")
		}
		if shot.Seq <= 0 {
			shot.Seq = index + 1
		}
		shot.ProjectID = storyboard.ProjectID
		shot.OwnerUsername = storyboard.OwnerUsername
	}
	for index := range stored {
		if stored[index].EpisodeNo != storyboard.EpisodeNo || stored[index].Engine != storyboard.Engine {
			continue
		}
		storyboard.ID = stored[index].ID
		if storyboard.ID == "" {
			storyboard.ID = newScreenplayID("sb")
		}
		storyboard.Revision = stored[index].Revision + 1
		storyboard.CreatedAt = stored[index].CreatedAt
		storyboard.UpdatedAt = now
		for shotIndex := range storyboard.Shots {
			storyboard.Shots[shotIndex].StoryboardID = storyboard.ID
		}
		stored[index] = cloneStoryboard(storyboard)
		m.storyboards[key] = stored
		return cloneStoryboard(stored[index]), nil
	}
	if storyboard.ID == "" {
		storyboard.ID = newScreenplayID("sb")
	}
	for shotIndex := range storyboard.Shots {
		storyboard.Shots[shotIndex].StoryboardID = storyboard.ID
	}
	storyboard.Revision = 1
	storyboard.CreatedAt = now
	storyboard.UpdatedAt = now
	m.storyboards[key] = append(stored, cloneStoryboard(storyboard))
	return cloneStoryboard(storyboard), nil
}

func (m *MemoryStore) GetStoryboard(_ context.Context, owner, projectID string, episodeNo int, engine StoryboardEngine) (Storyboard, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, storyboard := range m.storyboards[projectKey(owner, projectID)] {
		if storyboard.EpisodeNo == episodeNo && storyboard.Engine == engine {
			return cloneStoryboard(storyboard), nil
		}
	}
	return Storyboard{}, fmt.Errorf("%w: 第 %d 集 %s 分镜", ErrNotFound, episodeNo, engine)
}

func (m *MemoryStore) ListStoryboards(_ context.Context, owner, projectID string, engine StoryboardEngine) ([]Storyboard, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored := m.storyboards[projectKey(owner, projectID)]
	out := make([]Storyboard, 0, len(stored))
	for _, storyboard := range stored {
		if engine != "" && storyboard.Engine != engine {
			continue
		}
		out = append(out, cloneStoryboard(storyboard))
	}
	return out, nil
}

func cloneStoryboard(storyboard Storyboard) Storyboard {
	out := storyboard
	out.Shots = append([]StoryboardShot(nil), storyboard.Shots...)
	return out
}

func cloneBatch(batch *Batch) *Batch {
	if batch == nil {
		return nil
	}
	raw, err := json.Marshal(batch)
	if err != nil {
		panic(err)
	}
	var out Batch
	if err := json.Unmarshal(raw, &out); err != nil {
		panic(err)
	}
	if out.Steps == nil {
		out.Steps = map[string]*StepState{}
	}
	return &out
}
