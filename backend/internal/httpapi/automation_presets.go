package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

var (
	ErrAutomationPresetNotFound = errors.New("automation preset not found")
	ErrAutomationPresetConflict = errors.New("automation preset version conflict")
	ErrAutomationPresetInvalid  = errors.New("automation preset is invalid")
)

type AutomationPreset struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Version   int            `json:"version"`
	Config    map[string]any `json:"config"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

type AutomationPresetStore interface {
	List(string) ([]AutomationPreset, error)
	Get(string, string) (*AutomationPreset, error)
	Create(string, string, map[string]any) (AutomationPreset, error)
	Update(string, string, int, string, map[string]any) (AutomationPreset, error)
	Remove(string, string) error
}

type automationPresetFile struct {
	Version int                           `json:"version"`
	Presets map[string][]AutomationPreset `json:"presets"`
}

type FileAutomationPresetStore struct {
	path string
	now  func() time.Time
	mu   sync.Mutex
	data automationPresetFile
}

func NewFileAutomationPresetStore(path string, now func() time.Time) (*FileAutomationPresetStore, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("%w: path is required", ErrAutomationPresetInvalid)
	}
	if now == nil {
		now = time.Now
	}
	store := &FileAutomationPresetStore{path: path, now: now, data: automationPresetFile{Version: 1, Presets: map[string][]AutomationPreset{}}}
	raw, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(raw, &store.data); err != nil {
			return nil, fmt.Errorf("read automation presets: %w", err)
		}
		if store.data.Version == 0 {
			store.data.Version = 1
		}
		if store.data.Presets == nil {
			store.data.Presets = map[string][]AutomationPreset{}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read automation presets: %w", err)
	}
	return store, nil
}

func (s *FileAutomationPresetStore) List(owner string) ([]AutomationPreset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := append([]AutomationPreset(nil), s.data.Presets[strings.TrimSpace(owner)]...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].UpdatedAt > items[j].UpdatedAt })
	return cloneAutomationPresets(items), nil
}

func (s *FileAutomationPresetStore) Get(owner, id string) (*AutomationPreset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.data.Presets[strings.TrimSpace(owner)] {
		if item.ID == strings.TrimSpace(id) {
			copy := cloneAutomationPreset(item)
			return &copy, nil
		}
	}
	return nil, ErrAutomationPresetNotFound
}

func (s *FileAutomationPresetStore) Create(owner, name string, config map[string]any) (AutomationPreset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owner, name = strings.TrimSpace(owner), strings.TrimSpace(name)
	if owner == "" || name == "" || len(config) == 0 {
		return AutomationPreset{}, ErrAutomationPresetInvalid
	}
	now := s.now().UTC().Format(time.RFC3339Nano)
	item := AutomationPreset{ID: newAutomationPresetID(), Name: name, Version: 1, Config: cloneAutomationMap(config), CreatedAt: now, UpdatedAt: now}
	s.data.Presets[owner] = append(s.data.Presets[owner], item)
	if err := s.persistLocked(); err != nil {
		return AutomationPreset{}, err
	}
	return cloneAutomationPreset(item), nil
}

func (s *FileAutomationPresetStore) Update(owner, id string, expectedVersion int, name string, config map[string]any) (AutomationPreset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owner, id = strings.TrimSpace(owner), strings.TrimSpace(id)
	items := s.data.Presets[owner]
	for index := range items {
		if items[index].ID != id {
			continue
		}
		if expectedVersion != items[index].Version {
			return AutomationPreset{}, ErrAutomationPresetConflict
		}
		if strings.TrimSpace(name) == "" {
			name = items[index].Name
		}
		if len(config) == 0 {
			config = items[index].Config
		}
		if strings.TrimSpace(name) == "" || len(config) == 0 {
			return AutomationPreset{}, ErrAutomationPresetInvalid
		}
		items[index].Name = strings.TrimSpace(name)
		items[index].Config = cloneAutomationMap(config)
		items[index].Version++
		items[index].UpdatedAt = s.now().UTC().Format(time.RFC3339Nano)
		s.data.Presets[owner] = items
		if err := s.persistLocked(); err != nil {
			return AutomationPreset{}, err
		}
		return cloneAutomationPreset(items[index]), nil
	}
	return AutomationPreset{}, ErrAutomationPresetNotFound
}

func (s *FileAutomationPresetStore) Remove(owner, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	owner, id = strings.TrimSpace(owner), strings.TrimSpace(id)
	items := s.data.Presets[owner]
	for index := range items {
		if items[index].ID != id {
			continue
		}
		s.data.Presets[owner] = append(items[:index], items[index+1:]...)
		return s.persistLocked()
	}
	return ErrAutomationPresetNotFound
}

func (s *FileAutomationPresetStore) persistLocked() error {
	encoded, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode automation presets: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create automation preset directory: %w", err)
	}
	temporary := fmt.Sprintf("%s.%d.tmp", s.path, time.Now().UnixNano())
	if err := os.WriteFile(temporary, encoded, 0o600); err != nil {
		return fmt.Errorf("write automation presets: %w", err)
	}
	if err := os.Rename(temporary, s.path); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("replace automation presets: %w", err)
	}
	return nil
}

func newAutomationPresetID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("auto-preset-%d", time.Now().UnixNano())
	}
	return "auto-preset-" + hex.EncodeToString(buf)
}

func cloneAutomationPreset(item AutomationPreset) AutomationPreset {
	item.Config = cloneAutomationMap(item.Config)
	return item
}

func cloneAutomationPresets(items []AutomationPreset) []AutomationPreset {
	cloned := make([]AutomationPreset, len(items))
	for index, item := range items {
		cloned[index] = cloneAutomationPreset(item)
	}
	return cloned
}

func cloneAutomationMap(input map[string]any) map[string]any {
	if input == nil {
		return map[string]any{}
	}
	encoded, err := json.Marshal(input)
	if err != nil {
		return map[string]any{}
	}
	var output map[string]any
	if json.Unmarshal(encoded, &output) != nil || output == nil {
		return map[string]any{}
	}
	return output
}

func registerAutomationPresetRoutes(mux *http.ServeMux, auth BridgeAuth, store AutomationPresetStore, version string) {
	if mux == nil || store == nil {
		return
	}
	prefix := "/api/batch-factory/" + version + "/automation-presets"
	mux.Handle(prefix, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		switch r.Method {
		case http.MethodGet:
			items, err := store.List(owner)
			if err != nil {
				writeAutomationPresetError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"presets": items})
		case http.MethodPost:
			var input struct {
				Name   string         `json:"name"`
				Config map[string]any `json:"config"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			item, err := store.Create(owner, input.Name, input.Config)
			if err != nil {
				writeAutomationPresetError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"preset": item})
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})))
	mux.Handle(prefix+"/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		id := strings.TrimPrefix(r.URL.Path, prefix+"/")
		if strings.Contains(id, "/") || strings.TrimSpace(id) == "" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		switch r.Method {
		case http.MethodPut:
			var input struct {
				Name            string         `json:"name"`
				Config          map[string]any `json:"config"`
				ExpectedVersion int            `json:"expectedVersion"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			item, err := store.Update(owner, id, input.ExpectedVersion, input.Name, input.Config)
			if err != nil {
				writeAutomationPresetError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"preset": item})
		case http.MethodDelete:
			if err := store.Remove(owner, id); err != nil {
				writeAutomationPresetError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})))
}

func registerBatchAutomationPresetRoute(mux *http.ServeMux, auth BridgeAuth, batchStore batchfactoryv11.Store, presets AutomationPresetStore, version string) {
	if mux == nil || batchStore == nil || presets == nil {
		return
	}
	mux.Handle("POST /api/batch-factory/"+version+"/batches/{batchId}/automation-presets", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			Name string `json:"name"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		batch, err := batchStore.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		config := map[string]any{}
		for key, raw := range batch.SettingsState.Patch {
			var value any
			if json.Unmarshal(raw, &value) == nil {
				config[key] = value
			}
		}
		item, err := presets.Create(owner, input.Name, config)
		if err != nil {
			writeAutomationPresetError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"preset": item})
	})))
}

func writeAutomationPresetError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	switch {
	case errors.Is(err, ErrAutomationPresetNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrAutomationPresetConflict):
		status = http.StatusConflict
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
