package agentchat

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Store is the persistence contract for workspaces, sessions and messages.
// The file implementation below mirrors the original on-disk layout so data
// can be migrated or compared directly.
type Store interface {
	ListWorkspaces(ctx context.Context, owner string) ([]Workspace, error)
	Workspace(ctx context.Context, owner, id string) (Workspace, error)
	CreateWorkspace(ctx context.Context, owner string, ws Workspace) (Workspace, error)
	UpdateWorkspace(ctx context.Context, owner, id string, patch WorkspacePatch) (Workspace, error)
	DeleteWorkspace(ctx context.Context, owner, id string) error

	ListSessions(ctx context.Context, owner, workspaceID string) ([]SessionInfo, error)
	Session(ctx context.Context, owner string, id int64) (SessionInfo, error)
	CreateSession(ctx context.Context, owner string, info SessionInfo) (SessionInfo, error)
	UpdateSession(ctx context.Context, owner string, id int64, patch SessionPatch) (SessionInfo, error)
	DeleteSession(ctx context.Context, owner string, id int64) error

	Messages(ctx context.Context, owner string, sessionID int64) ([]Message, error)
	SaveMessages(ctx context.Context, owner string, sessionID int64, messages []Message) error
}

// WorkspacePatch carries the mutable workspace fields.
type WorkspacePatch struct {
	Name            *string
	Pinned          *bool
	Archived        *bool
	EnabledSkillIDs *[]string
	Stage           *string
}

// SessionPatch carries the mutable session fields.
type SessionPatch struct {
	Title      *string
	Model      *string
	Permission *string
	Pinned     *bool
}

// Errors used by the HTTP layer to pick a Chinese message.
var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrSessionNotFound   = errors.New("session not found")
	ErrWorkspaceExists   = errors.New("workspace already exists")
)

// FileStore persists everything under <baseDir>/<owner>.
type FileStore struct {
	baseDir string
	now     func() time.Time
	mu      sync.Mutex
}

// NewFileStore creates the file-backed store. baseDir is the same users
// directory the other Agent stores already use (UserConfigStore.BaseDir).
func NewFileStore(baseDir string, clock func() time.Time) *FileStore {
	if clock == nil {
		clock = time.Now
	}
	return &FileStore{baseDir: baseDir, now: clock}
}

func (s *FileStore) agentDir(owner string) string {
	return filepath.Join(s.baseDir, safeSegment(owner), "agent")
}

func (s *FileStore) sessionsDir(owner string) string {
	return filepath.Join(s.baseDir, safeSegment(owner), "AI创作")
}

func (s *FileStore) indexPath(owner string) string {
	return filepath.Join(s.agentDir(owner), "workspaces.json")
}

// UserAgentRoot returns the server-owned workspace directory for one user.
// Client-supplied workspace roots must resolve inside this directory.
func (s *FileStore) UserAgentRoot(owner string) string {
	return s.agentDir(owner)
}

// DefaultWorkspaceRoot is where the auto-created default workspace lives.
func (s *FileStore) DefaultWorkspaceRoot(owner string) string {
	return filepath.Join(s.agentDir(owner), "default-workspace")
}

// RootPath exposes the resolved root of a workspace for the tool layer.
func (s *FileStore) RootPath(ctx context.Context, owner, id string) (string, error) {
	ws, err := s.Workspace(ctx, owner, id)
	if err != nil {
		return "", err
	}
	return ws.RootPath, nil
}

func safeSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "." || value == ".." || strings.ContainsAny(value, `/\`) {
		return "invalid-user"
	}
	return value
}

// ReadIndex reads workspaces.json, creating the default workspace on first use.
func (s *FileStore) readIndex(owner string) (workspaceIndex, error) {
	path := s.indexPath(owner)
	raw, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return workspaceIndex{}, err
		}
		index := workspaceIndex{Version: 1, Workspaces: []Workspace{}}
		defaultRoot := s.DefaultWorkspaceRoot(owner)
		if err := os.MkdirAll(defaultRoot, 0o700); err != nil {
			return workspaceIndex{}, err
		}
		now := s.now().UnixMilli()
		defaultWorkspace := Workspace{
			ID:              newID(""),
			Name:            "默认工作区",
			Slug:            "默认工作区",
			RootPath:        defaultRoot,
			IsDefault:       true,
			CreatedAt:       now,
			UpdatedAt:       now,
			EnabledSkillIDs: []string{},
			Stage:           "input",
		}
		index.Workspaces = append(index.Workspaces, defaultWorkspace)
		if err := s.syncWorkspaceMeta(defaultWorkspace); err != nil {
			return workspaceIndex{}, err
		}
		if err := writeJSONFile(path, index); err != nil {
			return workspaceIndex{}, err
		}
		return index, nil
	}
	var index workspaceIndex
	if err := decodeJSON(raw, &index); err != nil {
		return workspaceIndex{}, err
	}
	if index.Workspaces == nil {
		index.Workspaces = []Workspace{}
	}
	return index, nil
}

func (s *FileStore) writeIndex(owner string, index workspaceIndex) error {
	if index.Version == 0 {
		index.Version = 1
	}
	if index.Workspaces == nil {
		index.Workspaces = []Workspace{}
	}
	return writeJSONFile(s.indexPath(owner), index)
}

func (s *FileStore) syncWorkspaceMeta(ws Workspace) error {
	luoshui := filepath.Join(ws.RootPath, ".luoshui")
	for _, dir := range []string{
		luoshui,
		filepath.Join(luoshui, "skills"),
		filepath.Join(luoshui, "context"),
		filepath.Join(luoshui, "context", "automation"),
		filepath.Join(luoshui, "sessions"),
		filepath.Join(luoshui, "runs"),
		filepath.Join(ws.RootPath, "input"),
		filepath.Join(ws.RootPath, "output"),
	} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	return writeJSONFile(filepath.Join(luoshui, "workspace.json"), workspaceMeta{
		Name:            ws.Name,
		Version:         1,
		EnabledSkillIDs: nonNilStrings(ws.EnabledSkillIDs),
		Stage:           ws.Stage,
	})
}

func (s *FileStore) ListWorkspaces(ctx context.Context, owner string) ([]Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	index, err := s.readIndex(owner)
	if err != nil {
		return nil, err
	}
	workspaces := append([]Workspace(nil), index.Workspaces...)
	sort.SliceStable(workspaces, func(i, j int) bool {
		if workspaces[i].Pinned != workspaces[j].Pinned {
			return workspaces[i].Pinned
		}
		return workspaces[i].CreatedAt < workspaces[j].CreatedAt
	})
	return workspaces, nil
}

func (s *FileStore) Workspace(ctx context.Context, owner, id string) (Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.workspaceLocked(owner, id)
}

func (s *FileStore) workspaceLocked(owner, id string) (Workspace, error) {
	index, err := s.readIndex(owner)
	if err != nil {
		return Workspace{}, err
	}
	for _, ws := range index.Workspaces {
		if ws.ID == id {
			return ws, nil
		}
	}
	return Workspace{}, ErrWorkspaceNotFound
}

func (s *FileStore) CreateWorkspace(ctx context.Context, owner string, ws Workspace) (Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	index, err := s.readIndex(owner)
	if err != nil {
		return Workspace{}, err
	}
	for _, existing := range index.Workspaces {
		if strings.EqualFold(existing.Name, ws.Name) {
			return Workspace{}, ErrWorkspaceExists
		}
	}
	now := s.now().UnixMilli()
	if ws.ID == "" {
		ws.ID = newID("")
	}
	if ws.Slug == "" {
		ws.Slug = ws.Name
	}
	if ws.RootPath == "" {
		ws.RootPath = filepath.Join(s.agentDir(owner), slugForPath(ws.Name))
	} else {
		rootPath, err := ResolveWorkspaceRoot(s.agentDir(owner), ws.RootPath)
		if err != nil {
			return Workspace{}, err
		}
		ws.RootPath = rootPath
	}
	if ws.EnabledSkillIDs == nil {
		ws.EnabledSkillIDs = []string{}
	}
	if ws.Stage == "" {
		ws.Stage = "input"
	}
	ws.CreatedAt = now
	ws.UpdatedAt = now
	if err := s.syncWorkspaceMeta(ws); err != nil {
		return Workspace{}, err
	}
	index.Workspaces = append(index.Workspaces, ws)
	if err := s.writeIndex(owner, index); err != nil {
		return Workspace{}, err
	}
	return ws, nil
}

func (s *FileStore) UpdateWorkspace(ctx context.Context, owner, id string, patch WorkspacePatch) (Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	index, err := s.readIndex(owner)
	if err != nil {
		return Workspace{}, err
	}
	for i := range index.Workspaces {
		ws := &index.Workspaces[i]
		if ws.ID != id {
			continue
		}
		if patch.Name != nil && strings.TrimSpace(*patch.Name) != "" {
			ws.Name = strings.TrimSpace(*patch.Name)
			ws.Slug = ws.Name
		}
		if patch.Pinned != nil {
			ws.Pinned = *patch.Pinned
		}
		if patch.Archived != nil {
			ws.Archived = *patch.Archived
		}
		if patch.EnabledSkillIDs != nil {
			ws.EnabledSkillIDs = *patch.EnabledSkillIDs
		}
		if patch.Stage != nil {
			ws.Stage = *patch.Stage
		}
		ws.UpdatedAt = s.now().UnixMilli()
		if err := s.syncWorkspaceMeta(*ws); err != nil {
			return Workspace{}, err
		}
		if err := s.writeIndex(owner, index); err != nil {
			return Workspace{}, err
		}
		return *ws, nil
	}
	return Workspace{}, ErrWorkspaceNotFound
}

func (s *FileStore) DeleteWorkspace(ctx context.Context, owner, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	index, err := s.readIndex(owner)
	if err != nil {
		return err
	}
	next := make([]Workspace, 0, len(index.Workspaces))
	for _, ws := range index.Workspaces {
		if ws.ID != id {
			next = append(next, ws)
			continue
		}
		if ws.IsDefault {
			return errors.New("默认工作区不可删除")
		}
	}
	if len(next) == len(index.Workspaces) {
		return ErrWorkspaceNotFound
	}
	index.Workspaces = next
	return s.writeIndex(owner, index)
}

func (s *FileStore) ListSessions(ctx context.Context, owner, workspaceID string) ([]SessionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	root := s.sessionsDir(owner)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []SessionInfo{}, nil
		}
		return nil, err
	}
	sessions := make([]SessionInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := readSessionInfo(filepath.Join(root, entry.Name(), "info.json"))
		if err != nil {
			continue
		}
		if workspaceID != "" && info.WorkspaceID != workspaceID {
			continue
		}
		sessions = append(sessions, info)
	}
	sort.SliceStable(sessions, func(i, j int) bool { return sessions[i].UpdatedAt > sessions[j].UpdatedAt })
	return sessions, nil
}

func (s *FileStore) Session(ctx context.Context, owner string, id int64) (SessionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir, err := s.sessionDir(owner, id)
	if err != nil {
		return SessionInfo{}, err
	}
	info, err := readSessionInfo(filepath.Join(dir, "info.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return SessionInfo{}, ErrSessionNotFound
		}
		return SessionInfo{}, err
	}
	return info, nil
}

func (s *FileStore) sessionDir(owner string, id int64) (string, error) {
	root := s.sessionsDir(owner)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrSessionNotFound
		}
		return "", err
	}
	suffix := "_" + strconv.FormatInt(id, 10)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			return filepath.Join(root, entry.Name()), nil
		}
	}
	return "", ErrSessionNotFound
}

func (s *FileStore) CreateSession(ctx context.Context, owner string, info SessionInfo) (SessionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if info.ID == 0 {
		info.ID = now.UnixMilli()
	}
	if strings.TrimSpace(info.Title) == "" {
		info.Title = "新 Agent 会话"
	}
	if info.Kind == "" {
		info.Kind = "agent"
	}
	if info.ChatMode == "" {
		info.ChatMode = "agent"
	}
	info.PermissionMode = NormalizePermission(info.PermissionMode)
	info.AgentPermissionMode = info.PermissionMode
	info.CreatedAt = nowStamp(now)
	info.UpdatedAt = info.CreatedAt
	info.MessageCount = 0
	dir := filepath.Join(s.sessionsDir(owner), sessionDirName(info.Title, info.ID))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return SessionInfo{}, err
	}
	if err := writeJSONFile(filepath.Join(dir, "info.json"), info); err != nil {
		return SessionInfo{}, err
	}
	if err := writeJSONFile(filepath.Join(dir, "messages.json"), []Message{}); err != nil {
		return SessionInfo{}, err
	}
	return info, nil
}

func (s *FileStore) UpdateSession(ctx context.Context, owner string, id int64, patch SessionPatch) (SessionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir, err := s.sessionDir(owner, id)
	if err != nil {
		return SessionInfo{}, err
	}
	info, err := readSessionInfo(filepath.Join(dir, "info.json"))
	if err != nil {
		return SessionInfo{}, err
	}
	oldDir := dir
	if patch.Title != nil && strings.TrimSpace(*patch.Title) != "" {
		info.Title = strings.TrimSpace(*patch.Title)
	}
	if patch.Model != nil {
		info.Model = strings.TrimSpace(*patch.Model)
	}
	if patch.Permission != nil {
		mode := NormalizePermission(*patch.Permission)
		info.PermissionMode = mode
		info.AgentPermissionMode = mode
	}
	info.UpdatedAt = nowStamp(s.now())
	newDir := filepath.Join(s.sessionsDir(owner), sessionDirName(info.Title, info.ID))
	if newDir != oldDir {
		if err := os.Rename(oldDir, newDir); err != nil {
			return SessionInfo{}, err
		}
	}
	if err := writeJSONFile(filepath.Join(newDir, "info.json"), info); err != nil {
		return SessionInfo{}, err
	}
	return info, nil
}

func (s *FileStore) DeleteSession(ctx context.Context, owner string, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir, err := s.sessionDir(owner, id)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func (s *FileStore) Messages(ctx context.Context, owner string, sessionID int64) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir, err := s.sessionDir(owner, sessionID)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, "messages.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return []Message{}, nil
		}
		return nil, err
	}
	var messages []Message
	if err := decodeJSON(raw, &messages); err != nil {
		return nil, err
	}
	if messages == nil {
		messages = []Message{}
	}
	return messages, nil
}

func (s *FileStore) SaveMessages(ctx context.Context, owner string, sessionID int64, messages []Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveMessagesLocked(owner, sessionID, messages)
}

func (s *FileStore) saveMessagesLocked(owner string, sessionID int64, messages []Message) error {
	dir, err := s.sessionDir(owner, sessionID)
	if err != nil {
		return err
	}
	if messages == nil {
		messages = []Message{}
	}
	if err := writeJSONFile(filepath.Join(dir, "messages.json"), messages); err != nil {
		return err
	}
	info, err := readSessionInfo(filepath.Join(dir, "info.json"))
	if err != nil {
		return err
	}
	info.MessageCount = len(messages)
	info.UpdatedAt = nowStamp(s.now())
	return writeJSONFile(filepath.Join(dir, "info.json"), info)
}

// AppendRunAudit appends one JSON line to <workspace>/.luoshui/runs/<run>.jsonl.
// A failure here never blocks the reply.
func (s *FileStore) AppendRunAudit(workspaceRoot, runID string, line []byte) error {
	dir := filepath.Join(workspaceRoot, ".luoshui", "runs")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dir, strings.TrimSpace(runID)+".jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}

func readSessionInfo(path string) (SessionInfo, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return SessionInfo{}, err
	}
	var info SessionInfo
	if err := decodeJSON(raw, &info); err != nil {
		return SessionInfo{}, err
	}
	return info, nil
}

func sessionDirName(title string, id int64) string {
	title = sanitizeFileName(title)
	if len([]rune(title)) > 60 {
		title = string([]rune(title)[:60])
	}
	if title == "" {
		title = "未命名"
	}
	return fmt.Sprintf("%s_%d", title, id)
}

func sanitizeFileName(value string) string {
	replacer := strings.NewReplacer(`\`, "_", "/", "_", ":", "_", "*", "_", "?", "_", `"`, "_", "<", "_", ">", "_", "|", "_")
	return strings.TrimSpace(replacer.Replace(value))
}

func slugForPath(name string) string {
	value := sanitizeFileName(name)
	if value == "" {
		return "workspace"
	}
	return value
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
