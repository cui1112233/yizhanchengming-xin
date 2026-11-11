// Package safepath centralizes validation for identity values (mainly
// usernames) that are joined into filesystem paths. Every caller that builds a
// per-user path must go through this package so traversal, separators, drive
// prefixes and symlinked escapes are rejected consistently.
package safepath

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ErrInvalid marks identity values that are not safe to use as a filesystem
// path segment.
var ErrInvalid = errors.New("invalid path segment")

// Username trims and validates a username before it is joined into any
// filesystem path. It rejects empty values, parent-directory segments, path
// separators, Windows drive/volume prefixes and control characters.
func Username(raw string) (string, error) {
	username := strings.TrimSpace(raw)
	if username == "" || username == "." || username == ".." {
		return "", ErrInvalid
	}
	if strings.Contains(username, "..") ||
		strings.ContainsAny(username, `/\`) ||
		strings.ContainsAny(username, `:*?"<>|`) ||
		strings.ContainsRune(username, 0) {
		return "", ErrInvalid
	}
	for _, char := range username {
		if char < 0x20 {
			return "", ErrInvalid
		}
	}
	return username, nil
}

// ValidUsername reports whether raw is safe to use as a path segment.
func ValidUsername(raw string) bool {
	_, err := Username(raw)
	return err == nil
}

// UserPath validates rawUsername and joins it under baseDir. WithinDir then
// confirms the resulting path is still inside baseDir after symlink resolution.
func UserPath(baseDir, rawUsername string, parts ...string) (string, error) {
	username, err := Username(rawUsername)
	if err != nil {
		return "", err
	}
	base, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	elements := append([]string{base, username}, parts...)
	target := filepath.Join(elements...)
	if err := WithinDir(base, target); err != nil {
		return "", err
	}
	return target, nil
}

// WithinDir verifies that target stays inside baseDir. Existing ancestors are
// resolved with EvalSymlinks first, so a symlinked user directory cannot be
// used to escape baseDir. A target equal to baseDir is rejected because callers
// always append at least a username segment.
func WithinDir(baseDir, target string) error {
	base, err := filepath.Abs(filepath.Clean(baseDir))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	targetPath, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	resolvedBase, err := resolveExisting(base)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	resolvedTarget, err := resolveExisting(targetPath)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	rel, err := filepath.Rel(resolvedBase, resolvedTarget)
	if err != nil || rel == "." || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return ErrInvalid
	}
	return nil
}

// resolveExisting resolves symlinks for the deepest existing ancestor of path
// and re-appends the not-yet-created suffix. This lets containment checks run
// before the user directory is created while still following existing links.
func resolveExisting(path string) (string, error) {
	path = filepath.Clean(path)
	current := path
	remainder := ""
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			if remainder == "" {
				return resolved, nil
			}
			return filepath.Join(resolved, remainder), nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return filepath.Join(current, remainder), nil
		}
		if remainder == "" {
			remainder = filepath.Base(current)
		} else {
			remainder = filepath.Join(filepath.Base(current), remainder)
		}
		current = parent
	}
}
