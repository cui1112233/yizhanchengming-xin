package httpapi

import (
	"qiantie/backend/internal/safepath"
)

// ErrInvalidUsername marks identity values that are not safe to use as a
// filesystem path segment.
var ErrInvalidUsername = safepath.ErrInvalid

// SafeUsername trims and validates a username before it is joined into any
// filesystem path. It rejects empty values, parent-directory segments, path
// separators, Windows drive/volume prefixes and control characters.
func SafeUsername(raw string) (string, error) {
	return safepath.Username(raw)
}

// ValidUsername reports whether raw is safe to use as a path segment.
func ValidUsername(raw string) bool {
	return safepath.ValidUsername(raw)
}

// UserScopedPath validates username and joins it under baseDir while verifying
// that the result stays inside baseDir after symlink resolution.
func UserScopedPath(baseDir, rawUsername string, parts ...string) (string, error) {
	return safepath.UserPath(baseDir, rawUsername, parts...)
}

// WithinDir verifies that target stays inside baseDir after symlink resolution.
func WithinDir(baseDir, target string) error {
	return safepath.WithinDir(baseDir, target)
}
