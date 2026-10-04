package providerconfig

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

type execFake struct { query string; args []any }
func (f *execFake) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	f.query = query; f.args = append([]any(nil), args...); return configResult(1), nil
}
type configResult int64
func (r configResult) LastInsertId() (int64,error) { return int64(r),nil }
func (r configResult) RowsAffected() (int64,error) { return int64(r),nil }

type rowFake struct { values []any; err error }
func (r rowFake) Scan(dest ...any) error {
	if r.err != nil { return r.err }
	for i, value := range r.values {
		switch target := dest[i].(type) {
		case *string: *target = value.(string)
		case *[]byte: if value == nil { *target = nil } else { *target = append([]byte(nil), value.([]byte)...) }
		case *bool: *target = value.(bool)
		case *time.Time: *target = value.(time.Time)
		default: return errors.New("unsupported scan target")
		}
	}
	return nil
}

func TestStorePutEncryptsCredentialBeforeSQL(t *testing.T) {
	cipher, _ := NewCipher(bytes.Repeat([]byte{0x41}, 32))
	exec := &execFake{}
	store := newSQLStore(exec, nil, cipher)
	record, err := store.Put(context.Background(), PutInput{
		Owner: "owner", MediaKind: KindVideo, Provider: "personal_api", Model: "yd2.0-mini", APIKey: "plain-secret",
		CreateURL: "https://create", TasksURL: "https://tasks", ResultURL: "https://result/{id}", Enabled: true,
	})
	if err != nil { t.Fatal(err) }
	if record.ID == "" || record.Configured != true { t.Fatalf("record=%#v", record) }
	for _, arg := range exec.args {
		if value, ok := arg.(string); ok && value == "plain-secret" { t.Fatal("plaintext API key reached SQL arguments") }
		if value, ok := arg.([]byte); ok && bytes.Equal(value, []byte("plain-secret")) { t.Fatal("plaintext API key reached SQL arguments") }
	}
}

func TestStoreResolveDecryptsCredentialAndScopesOwner(t *testing.T) {
	cipher, _ := NewCipher(bytes.Repeat([]byte{0x24}, 32))
	nonce, encrypted, _ := cipher.Encrypt("alice", KindVideo, "personal_api", []byte("secret"))
	now := time.Date(2026,10,4,12,0,0,0,time.UTC)
	var gotArgs []any
	store := newSQLStore(&execFake{}, func(_ context.Context, query string, args ...any) rowScanner {
		gotArgs = append([]any(nil), args...)
		return rowFake{values: []any{"cfg_1","alice",KindVideo,"personal_api","yd2.0-mini","https://create","https://tasks","https://result/{id}",nonce,encrypted,[]byte(`{"resolution":"720p"}`),true,now,now}}
	}, cipher)
	resolved, err := store.Resolve(context.Background(), "alice", KindVideo, "personal_api")
	if err != nil { t.Fatal(err) }
	if resolved.APIKey != "secret" || resolved.Model != "yd2.0-mini" { t.Fatalf("resolved=%#v", resolved) }
	if len(gotArgs) != 3 || gotArgs[0] != "alice" || gotArgs[1] != KindVideo || gotArgs[2] != "personal_api" { t.Fatalf("args=%#v", gotArgs) }
}

func TestStoreResolveMapsMissingProvider(t *testing.T) {
	cipher, _ := NewCipher(bytes.Repeat([]byte{0x55}, 32))
	store := newSQLStore(&execFake{}, func(context.Context,string,...any) rowScanner { return rowFake{err: sql.ErrNoRows} }, cipher)
	if _, err := store.Resolve(context.Background(), "owner", KindVideo, "missing"); !errors.Is(err, ErrNotFound) { t.Fatalf("err=%v", err) }
}

func TestStoreRejectsUnsupportedMediaKind(t *testing.T) {
	cipher, _ := NewCipher(bytes.Repeat([]byte{0x66}, 32))
	store := newSQLStore(&execFake{}, nil, cipher)
	if _, err := store.Put(context.Background(), PutInput{Owner:"owner", MediaKind:"audio", Provider:"x", Enabled:true}); err == nil { t.Fatal("expected media kind error") }
}
