package authn

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSecurityAuditSessionTokenHashesCannotBeSerialized(t *testing.T) {
	record := SessionRecord{
		UserID:           7,
		AccessTokenHash:  "audit-access-token-hash",
		RefreshTokenHash: "audit-refresh-token-hash",
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	body := string(encoded)
	for _, forbidden := range []string{"audit-access-token-hash", "audit-refresh-token-hash", "AccessTokenHash", "RefreshTokenHash"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("session token hash is JSON-serializable: %s", body)
		}
	}
}
