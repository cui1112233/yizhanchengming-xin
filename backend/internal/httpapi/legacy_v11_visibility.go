package httpapi

import "os"

// legacyV11Public is intentionally opt-in. Production and local review
// runtimes expose V12 only; tests can opt in while the old handlers are
// migrated internally.
func legacyV11Public() bool {
	return os.Getenv("QIANTIE_ENABLE_LEGACY_V11") == "1"
}
