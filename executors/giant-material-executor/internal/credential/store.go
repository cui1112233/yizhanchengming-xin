package credential

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrNotFound = errors.New("executor credential not found")

type Record struct {
	ExecutorID string    `json:"executorId"`
	Token      string    `json:"token"`
	SavedAt    time.Time `json:"savedAt"`
}

type Store interface {
	Load() (Record, error)
	Save(Record) error
	Clear() error
}

type FileStore struct {
	path string
}

type protectedStore struct {
	file *FileStore
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: strings.TrimSpace(path)}
}

func (s *FileStore) Load() (Record, error) {
	data, err := s.read()
	if err != nil {
		return Record{}, err
	}
	return decodeRecord(data)
}

func (s *FileStore) Save(record Record) error {
	if err := record.validate(); err != nil {
		return err
	}
	if record.SavedAt.IsZero() {
		record.SavedAt = time.Now().UTC()
	}
	data, err := json.Marshal(record)
	if err != nil {
		return errors.New("marshal executor credential")
	}
	return s.write(data)
}

func NewProtectedStore(path string) Store {
	return &protectedStore{file: NewFileStore(path)}
}

func (s *protectedStore) Load() (Record, error) {
	data, err := s.file.read()
	if err != nil {
		return Record{}, err
	}
	plain, err := unprotect(data)
	if err != nil {
		return Record{}, errors.New("executor credential is malformed")
	}
	return decodeRecord(plain)
}

func (s *protectedStore) Save(record Record) error {
	if err := record.validate(); err != nil {
		return err
	}
	if record.SavedAt.IsZero() {
		record.SavedAt = time.Now().UTC()
	}
	plain, err := json.Marshal(record)
	if err != nil {
		return errors.New("marshal executor credential")
	}
	data, err := protect(plain)
	if err != nil {
		return errors.New("protect executor credential")
	}
	return s.file.write(data)
}

func (s *protectedStore) Clear() error {
	return s.file.Clear()
}

func (s *FileStore) read() ([]byte, error) {
	if s.path == "" {
		return nil, errors.New("executor credential path is required")
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read executor credential: %w", err)
	}
	return data, nil
}

func (s *FileStore) write(data []byte) error {
	if s.path == "" {
		return errors.New("executor credential path is required")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create executor credential directory: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".executor-credential-*")
	if err != nil {
		return fmt.Errorf("create executor credential temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("protect executor credential temp file: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write executor credential temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync executor credential temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close executor credential temp file: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		_ = os.Remove(s.path)
		if retryErr := os.Rename(tmpName, s.path); retryErr != nil {
			return fmt.Errorf("replace executor credential: %w", retryErr)
		}
	}
	return nil
}

func decodeRecord(data []byte) (Record, error) {
	var record Record
	if err := json.Unmarshal(data, &record); err != nil || record.validate() != nil {
		return Record{}, errors.New("executor credential is malformed")
	}
	return record, nil
}

func (s *FileStore) Clear() error {
	if err := os.Remove(s.path); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("clear executor credential: %w", err)
	}
	return nil
}

func (record Record) validate() error {
	if strings.TrimSpace(record.ExecutorID) == "" {
		return errors.New("executor ID is required")
	}
	if strings.TrimSpace(record.Token) == "" {
		return errors.New("executor token is required")
	}
	return nil
}
