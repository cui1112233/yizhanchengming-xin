package videomanagementconnector

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

type SessionStore interface {
	Save(context.Context, string, Session) error
	Load(context.Context, string) (Session, error)
}

// SessionRemover is optional so existing stores and test doubles remain
// compatible while the Go connector takes over the browser-worker lifecycle.
type SessionRemover interface {
	Remove(context.Context, string) error
}

type MemorySessionStore struct{ values map[string]Session }

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{values: map[string]Session{}}
}
func (s *MemorySessionStore) Save(_ context.Context, owner string, value Session) error {
	s.values[owner] = value
	return nil
}
func (s *MemorySessionStore) Load(_ context.Context, owner string) (Session, error) {
	value, ok := s.values[owner]
	if !ok {
		return Session{}, os.ErrNotExist
	}
	return value, nil
}
func (s *MemorySessionStore) Remove(_ context.Context, owner string) error {
	delete(s.values, owner)
	return nil
}
func (s *MemorySessionStore) RawForTest(owner string) string {
	value, _ := s.values[owner]
	raw, _ := json.Marshal(value)
	return string(raw)
}

type FileSessionStore struct {
	root string
	key  []byte
}

func NewFileSessionStore(root string, secret []byte) (*FileSessionStore, error) {
	if root == "" || len(secret) == 0 {
		return nil, errors.New("session store root and key are required")
	}
	hash := sha256.Sum256(secret)
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	return &FileSessionStore{root: root, key: hash[:]}, nil
}

func (s *FileSessionStore) pathForTest(owner string) string {
	return filepath.Join(s.root, safeOwner(owner)+".session")
}
func (s *FileSessionStore) Save(_ context.Context, owner string, value Session) error {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	encoded := append(nonce, gcm.Seal(nil, nonce, payload, nil)...)
	return os.WriteFile(s.pathForTest(owner), encoded, 0o600)
}
func (s *FileSessionStore) Load(_ context.Context, owner string) (Session, error) {
	raw, err := os.ReadFile(s.pathForTest(owner))
	if err != nil {
		return Session{}, err
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return Session{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return Session{}, err
	}
	if len(raw) < gcm.NonceSize() {
		return Session{}, errors.New("invalid encrypted session")
	}
	payload, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return Session{}, fmt.Errorf("decrypt session: %w", err)
	}
	var value Session
	if err := json.Unmarshal(payload, &value); err != nil {
		return Session{}, err
	}
	return value, nil
}
func (s *FileSessionStore) Remove(_ context.Context, owner string) error {
	err := os.Remove(s.pathForTest(owner))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

var ownerPattern = regexp.MustCompile(`[^a-zA-Z0-9_.-]+`)

func safeOwner(owner string) string {
	value := ownerPattern.ReplaceAllString(owner, "_")
	if value == "" {
		return "unknown"
	}
	return value
}
