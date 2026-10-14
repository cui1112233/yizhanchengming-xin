package httpapi

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("QIANTIE_ENABLE_LEGACY_V11", "1")
	os.Exit(m.Run())
}
