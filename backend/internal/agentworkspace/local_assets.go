package agentworkspace

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalAssetObjectStore keeps local review uploads on disk. Production keeps
// using TOS whenever the TOS configuration is present.
type LocalAssetObjectStore struct{ Root string }

func NewLocalAssetObjectStore(root string) *LocalAssetObjectStore {
	return &LocalAssetObjectStore{Root: filepath.Clean(strings.TrimSpace(root))}
}

func (s *LocalAssetObjectStore) objectPath(asset Asset) (string, error) {
	if s == nil || strings.TrimSpace(s.Root) == "" || strings.TrimSpace(asset.ObjectKey) == "" {
		return "", ErrInvalidAsset
	}
	key := filepath.FromSlash(asset.ObjectKey)
	if filepath.IsAbs(key) {
		return "", ErrInvalidAsset
	}
	root, err := filepath.Abs(s.Root)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(root, key))
	if err != nil {
		return "", err
	}
	if target == root || !strings.HasPrefix(target, root+string(os.PathSeparator)) {
		return "", ErrInvalidAsset
	}
	return target, nil
}

func (s *LocalAssetObjectStore) Put(ctx context.Context, asset Asset, body io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	target, err := s.objectPath(asset)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".agent-asset-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err = io.Copy(temporary, body); err == nil {
		err = temporary.Close()
	} else {
		_ = temporary.Close()
	}
	if err != nil {
		return err
	}
	return os.Rename(temporaryPath, target)
}

func (s *LocalAssetObjectStore) Open(ctx context.Context, asset Asset) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	target, err := s.objectPath(asset)
	if err != nil {
		return nil, err
	}
	body, err := os.Open(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrAssetNotFound
	}
	return body, err
}

func (s *LocalAssetObjectStore) Delete(ctx context.Context, asset Asset) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	target, err := s.objectPath(asset)
	if err != nil {
		return err
	}
	err = os.Remove(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
