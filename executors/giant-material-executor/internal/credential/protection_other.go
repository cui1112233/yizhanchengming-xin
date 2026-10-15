//go:build !windows

package credential

// Development platforms use the file store so the executor package remains
// testable on macOS/Linux. The production Windows binary uses DPAPI.
func protect(data []byte) ([]byte, error)   { return data, nil }
func unprotect(data []byte) ([]byte, error) { return data, nil }
