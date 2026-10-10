package httpapi

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMFATOTPMatchesRFC6238Vector(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	if !mfaTOTP(secret, "287082", time.Unix(59, 0)) {
		t.Fatal("expected SHA-1 TOTP vector to verify")
	}
	if mfaTOTP(secret, "000000", time.Unix(59, 0)) {
		t.Fatal("unexpected invalid TOTP verification")
	}
}

func TestVerifyMFARecoveryCodeConsumesSingleUseCode(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "mfa.json")
	rows := []mfaFileRow{{Username: "alice", Enabled: true, Secret: "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", RecoveryHashes: []string{mfaRecoveryHash("ABCDE-12345")}}}
	if err := writeMFAFile(path, rows); err != nil {
		t.Fatal(err)
	}
	required, valid, used, err := verifyMFAForLogin(path, "alice", "ABCDE-12345", time.Unix(59, 0))
	if err != nil || !required || !valid || !used {
		t.Fatalf("recovery verification = required %v valid %v used %v err %v", required, valid, used, err)
	}
	rowsAfter, err := readMFAFile(path)
	if err != nil || len(rowsAfter) != 1 || len(rowsAfter[0].RecoveryHashes) != 0 {
		t.Fatalf("recovery code was not consumed: %#v err=%v", rowsAfter, err)
	}
	_, valid, _, err = verifyMFAForLogin(path, "alice", "ABCDE-12345", time.Unix(59, 0))
	if err != nil || valid {
		t.Fatal("recovery code was accepted more than once")
	}
	_ = os.Remove(path)
}
