package auth

import "testing"

func TestDecodeLegacyAccountsSupportsArrayAndWrapper(t *testing.T) {
	for _, input := range []string{
		`[{"username":"alice","passwordHash":"scrypt$aa$bb","isOwner":true,"active":true,"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}]`,
		`{"accounts":[{"username":"alice","passwordHash":"scrypt$aa$bb","isOwner":true,"active":true,"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"}]}`,
	} {
		accounts, err := decodeLegacyAccounts([]byte(input))
		if err != nil || len(accounts) != 1 || accounts[0].Username != "alice" {
			t.Fatalf("decodeLegacyAccounts() = %+v, %v", accounts, err)
		}
	}
}

func TestValidateLegacyAccountRejectsMissingHashOrTime(t *testing.T) {
	if validateLegacyAccount(LegacyAccount{Username: "alice"}) == nil {
		t.Fatal("missing legacy fields should be rejected")
	}
}
