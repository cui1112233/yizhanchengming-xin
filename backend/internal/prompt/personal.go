package prompt

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var validPromptCategories = map[string]bool{"prefix": true, "quality": true, "restriction": true, "negative": true}

type PersonalPrompt struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Category   string     `json:"category"`
	Name       *string    `json:"name"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

type personalPromptState struct {
	Version int              `json:"version"`
	Prompts []PersonalPrompt `json:"prompts"`
}

type PersonalStore struct {
	path string
	mu   sync.Mutex
}

func NewPersonalStore(path string) *PersonalStore {
	return &PersonalStore{path: strings.TrimSpace(path)}
}

func (s *PersonalStore) loadLocked() (personalPromptState, error) {
	if s == nil || s.path == "" {
		return personalPromptState{}, errors.New("personal prompt path is not configured")
	}
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return personalPromptState{Version: 1, Prompts: []PersonalPrompt{}}, nil
	}
	if err != nil {
		return personalPromptState{}, err
	}
	var state personalPromptState
	if err := json.Unmarshal(data, &state); err != nil || state.Version != 1 {
		return personalPromptState{}, errors.New("invalid personal prompt store")
	}
	if state.Prompts == nil {
		state.Prompts = []PersonalPrompt{}
	}
	for _, item := range state.Prompts {
		if err := validatePersonalPrompt(item); err != nil {
			return personalPromptState{}, err
		}
	}
	return state, nil
}

func (s *PersonalStore) saveLocked(state personalPromptState) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".script-constraint-prompts-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}

func (s *PersonalStore) List(username, category string) ([]PersonalPrompt, error) {
	if !validPromptUsername(username) || !validPromptCategories[category] {
		return nil, errors.New("invalid personal prompt")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	result := make([]PersonalPrompt, 0)
	for _, item := range state.Prompts {
		if item.Username == username && item.Category == category {
			result = append(result, item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].LastUsedAt != nil && result[j].LastUsedAt != nil && !result[i].LastUsedAt.Equal(*result[j].LastUsedAt) {
			return result[i].LastUsedAt.After(*result[j].LastUsedAt)
		}
		if result[i].LastUsedAt != nil {
			return true
		}
		if result[j].LastUsedAt != nil {
			return false
		}
		return result[i].UpdatedAt.After(result[j].UpdatedAt)
	})
	return result, nil
}

func (s *PersonalStore) GetOwned(username, id string) (PersonalPrompt, bool, error) {
	if !validPromptUsername(username) || strings.TrimSpace(id) == "" {
		return PersonalPrompt{}, false, errors.New("invalid personal prompt")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.loadLocked()
	if err != nil {
		return PersonalPrompt{}, false, err
	}
	for _, item := range state.Prompts {
		if item.Username == username && item.ID == id {
			return item, true, nil
		}
	}
	return PersonalPrompt{}, false, nil
}

func (s *PersonalStore) CreateOrSaveDraft(username, category string, name *string, body string) (PersonalPrompt, error) {
	if err := validatePersonalInput(username, category, name, body); err != nil {
		return PersonalPrompt{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.loadLocked()
	if err != nil {
		return PersonalPrompt{}, err
	}
	now := time.Now().UTC()
	if name == nil {
		for index := range state.Prompts {
			if state.Prompts[index].Username == username && state.Prompts[index].Category == category && state.Prompts[index].Name == nil {
				state.Prompts[index].Body = strings.TrimSpace(body)
				state.Prompts[index].UpdatedAt = now
				if err := s.saveLocked(state); err != nil {
					return PersonalPrompt{}, err
				}
				return state.Prompts[index], nil
			}
		}
	}
	if name != nil {
		count := 0
		for _, item := range state.Prompts {
			if item.Username == username && item.Category == category && item.Name != nil {
				count++
			}
		}
		if count >= 100 {
			return PersonalPrompt{}, errors.New("too many personal prompts")
		}
		value := strings.TrimSpace(*name)
		name = &value
	}
	item := PersonalPrompt{ID: newPromptID(), Username: username, Category: category, Name: name, Body: strings.TrimSpace(body), CreatedAt: now, UpdatedAt: now}
	state.Prompts = append(state.Prompts, item)
	if err := s.saveLocked(state); err != nil {
		return PersonalPrompt{}, err
	}
	return item, nil
}

func (s *PersonalStore) Update(username, id string, name *string, body *string) (PersonalPrompt, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.loadLocked()
	if err != nil {
		return PersonalPrompt{}, false, err
	}
	for index := range state.Prompts {
		item := &state.Prompts[index]
		if item.Username != username || item.ID != id {
			continue
		}
		updatedName := item.Name
		updatedBody := item.Body
		if name != nil {
			updatedName = name
		}
		if body != nil {
			updatedBody = *body
		}
		if err := validatePersonalInput(username, item.Category, updatedName, updatedBody); err != nil {
			return PersonalPrompt{}, false, err
		}
		if updatedName != nil {
			value := strings.TrimSpace(*updatedName)
			updatedName = &value
		}
		item.Name, item.Body, item.UpdatedAt = updatedName, strings.TrimSpace(updatedBody), time.Now().UTC()
		if err := s.saveLocked(state); err != nil {
			return PersonalPrompt{}, false, err
		}
		return *item, true, nil
	}
	return PersonalPrompt{}, false, nil
}

func (s *PersonalStore) Remove(username, id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.loadLocked()
	if err != nil {
		return false, err
	}
	for index, item := range state.Prompts {
		if item.Username == username && item.ID == id {
			state.Prompts = append(state.Prompts[:index], state.Prompts[index+1:]...)
			return true, s.saveLocked(state)
		}
	}
	return false, nil
}

func (s *PersonalStore) MarkUsed(username string, ids []string) error {
	if !validPromptUsername(username) {
		return errors.New("invalid personal prompt")
	}
	wanted := make(map[string]bool)
	for _, id := range ids {
		wanted[id] = true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.loadLocked()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	changed := false
	for index := range state.Prompts {
		if state.Prompts[index].Username == username && wanted[state.Prompts[index].ID] {
			state.Prompts[index].LastUsedAt = &now
			changed = true
		}
	}
	if changed {
		return s.saveLocked(state)
	}
	return nil
}

func validatePersonalInput(username, category string, name *string, body string) error {
	if !validPromptUsername(username) || !validPromptCategories[category] || (name != nil && (strings.TrimSpace(*name) == "" || len(*name) > 80)) || strings.TrimSpace(body) == "" || len(body) > 12000 {
		return errors.New("invalid personal prompt")
	}
	return nil
}

func validatePersonalPrompt(item PersonalPrompt) error {
	if err := validatePersonalInput(item.Username, item.Category, item.Name, item.Body); err != nil || item.ID == "" || item.CreatedAt.IsZero() || item.UpdatedAt.IsZero() {
		return errors.New("invalid personal prompt store")
	}
	return nil
}

func validPromptUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}
	for _, char := range username {
		if !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_' || char == '-') {
			return false
		}
	}
	return true
}

func newPromptID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("prompt-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes[:])
}
