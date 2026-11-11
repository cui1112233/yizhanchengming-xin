package storage

import (
	"strings"
	"testing"
)

func TestV11ExternalSchemaStoresCiphertextAndRedactedAuditOnly(t *testing.T) {
	joined := strings.ToLower(strings.Join(V11ExternalStatements(), "\n"))
	for _, table := range []string{"batch_factory_v11_external_credentials", "batch_factory_v11_external_intents", "batch_factory_v11_external_audits"} {
		if !strings.Contains(joined, table) {
			t.Fatalf("missing %s", table)
		}
	}
	for _, column := range []string{"ciphertext blob", "nonce varbinary", "payload_digest", "reference_value", "message varchar(255)"} {
		if !strings.Contains(joined, column) {
			t.Fatalf("missing %s", column)
		}
	}
	if strings.Contains(joined, "password") || strings.Contains(joined, "token text") || strings.Contains(joined, "secret text") {
		t.Fatal("external schema must not contain plaintext credential fields")
	}
}
