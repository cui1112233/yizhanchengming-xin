package providerconfig

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestStorePutPreservesCredentialWhenAPIKeyIsOmitted(t *testing.T) {
	cipher, _ := NewCipher(bytes.Repeat([]byte{0x71}, 32))
	exec := &execFake{}
	store := newSQLStore(exec, nil, cipher)
	if _, err := store.Put(context.Background(), PutInput{Owner:"owner",MediaKind:KindVideo,Provider:"personal_api",Model:"yd2.0-mini",Enabled:true}); err != nil { t.Fatal(err) }
	if !strings.Contains(exec.query, "credential_nonce=COALESCE(VALUES(credential_nonce), credential_nonce)") {
		t.Fatalf("metadata-only update would overwrite encrypted nonce: %s", exec.query)
	}
	if !strings.Contains(exec.query, "credential_ciphertext=COALESCE(VALUES(credential_ciphertext), credential_ciphertext)") {
		t.Fatalf("metadata-only update would overwrite encrypted credential: %s", exec.query)
	}
}
