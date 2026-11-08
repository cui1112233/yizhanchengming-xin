package productionmedia

import (
	"errors"
	"testing"

	"qiantie/backend/internal/localartifact"
)

func TestDiskWriteGuardBlocksAtEightyFivePercent(t *testing.T) {
	guard := NewDiskWriteGuard(func() (float64, error) { return 85, nil })
	if err := guard(); !errors.Is(err, localartifact.ErrDiskPressure) {
		t.Fatalf("guard error=%v", err)
	}
}

func TestDiskWriteGuardAllowsBelowEightyFivePercent(t *testing.T) {
	guard := NewDiskWriteGuard(func() (float64, error) { return 84.9, nil })
	if err := guard(); err != nil {
		t.Fatalf("guard blocked safe disk usage: %v", err)
	}
}
