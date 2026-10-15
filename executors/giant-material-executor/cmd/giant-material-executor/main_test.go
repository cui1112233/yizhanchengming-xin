package main

import (
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"qiantie/giant-material-executor/internal/agent"
	"qiantie/giant-material-executor/internal/credential"
)

func TestDefaultPythonCommandMatchesPlatform(t *testing.T) {
	want := "python3"
	if runtime.GOOS == "windows" {
		want = "python"
	}
	if got := defaultPythonCommand(); got != want {
		t.Fatalf("python command=%q, want %q", got, want)
	}
}

func TestDefaultPythonCommandHonorsOverride(t *testing.T) {
	t.Setenv("GIANT_MATERIAL_PYTHON", "/custom/python")
	if got := defaultPythonCommand(); got != "/custom/python" {
		t.Fatalf("python command=%q, want override", got)
	}
}

func TestPublicAPIClientTimeoutLeavesRoomForServerClaimLongPoll(t *testing.T) {
	if publicAPIClientTimeout <= 25*time.Second {
		t.Fatalf("public API timeout=%s, must exceed the server's 25s claim long-poll window", publicAPIClientTimeout)
	}
}

func TestNormalizePublicAPIURLRequiresHTTPHost(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trims trailing slash", input: " https://factory.example.com/ ", want: "https://factory.example.com"},
		{name: "allows API path", input: "https://factory.example.com/control/", want: "https://factory.example.com/control"},
		{name: "rejects missing scheme", input: "factory.example.com", wantErr: true},
		{name: "rejects unsupported scheme", input: "ftp://factory.example.com", wantErr: true},
		{name: "rejects missing host", input: "https://", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizePublicAPIURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalize URL: %v", err)
			}
			if got != tc.want {
				t.Fatalf("normalized URL=%q, want %q", got, tc.want)
			}
		})
	}
}

func TestConfiguredPairingCodeTrimsEnvironment(t *testing.T) {
	t.Setenv("GIANT_MATERIAL_EXECUTOR_PAIRING_CODE", "  ABCD-1234  ")
	if got := configuredPairingCode(); got != "ABCD-1234" {
		t.Fatalf("pairing code=%q, want trimmed code", got)
	}
}

func TestConfiguredPairingCodeIsEmptyWithoutEnvironment(t *testing.T) {
	t.Setenv("GIANT_MATERIAL_EXECUTOR_PAIRING_CODE", "  ")
	if got := configuredPairingCode(); got != "" {
		t.Fatalf("pairing code=%q, want empty", got)
	}
}

func TestUnauthorizedRunClearsSavedCredential(t *testing.T) {
	store := credential.NewFileStore(filepath.Join(t.TempDir(), "executor.credential"))
	if err := savePairResult(store, agent.PairResult{ExecutorID: "executor-1", Token: "long-lived-token"}); err != nil {
		t.Fatalf("save pair result: %v", err)
	}
	if got := handleAgentRunError(store, agent.ErrUnauthorized); got != agent.BindingNeedsPairing {
		t.Fatalf("binding state=%s, want %s", got, agent.BindingNeedsPairing)
	}
	if _, err := store.Load(); !errors.Is(err, credential.ErrNotFound) {
		t.Fatalf("credential after unauthorized=%v, want ErrNotFound", err)
	}
}

func TestPreferLegacyCredentialStoreKeepsExistingPairing(t *testing.T) {
	legacy := &memoryCredentialStore{record: credential.Record{ExecutorID: "executor-1", Token: "token-1"}}
	device := &memoryCredentialStore{loadErr: credential.ErrNotFound}
	store, err := preferLegacyCredentialStore(device, legacy)
	if err != nil {
		t.Fatalf("select credential store: %v", err)
	}
	if store != legacy {
		t.Fatal("expected existing legacy pairing to be retained")
	}
}

func TestPreferLegacyCredentialStoreUsesDeviceStoreWhenNoLegacyRecordExists(t *testing.T) {
	device := &memoryCredentialStore{record: credential.Record{ExecutorID: "new-executor", Token: "new-token"}}
	legacy := &memoryCredentialStore{loadErr: credential.ErrNotFound}
	store, err := preferLegacyCredentialStore(device, legacy)
	if err != nil {
		t.Fatalf("select credential store: %v", err)
	}
	if store != device {
		t.Fatal("expected device store when no legacy pairing exists")
	}
}

func TestSavePairResultAndRestoreCredential(t *testing.T) {
	store := credential.NewFileStore(filepath.Join(t.TempDir(), "executor.credential"))
	result := agent.PairResult{ExecutorID: "executor-1", Token: "long-lived-token"}

	if err := savePairResult(store, result); err != nil {
		t.Fatalf("save pair result: %v", err)
	}
	got, err := loadExecutorCredential(store)
	if err != nil {
		t.Fatalf("load credential: %v", err)
	}
	if got.ExecutorID != result.ExecutorID || got.Token != result.Token {
		t.Fatalf("credential=%+v, want executor=%q token=%q", got, result.ExecutorID, result.Token)
	}
}

func TestLoadExecutorCredentialKeepsFirstLaunchUnpaired(t *testing.T) {
	store := credential.NewFileStore(filepath.Join(t.TempDir(), "executor.credential"))
	if _, err := loadExecutorCredential(store); err != credential.ErrNotFound {
		t.Fatalf("load error=%v, want ErrNotFound", err)
	}
}

func TestSavePairResultRejectsMissingServerCredential(t *testing.T) {
	store := credential.NewFileStore(filepath.Join(t.TempDir(), "executor.credential"))
	if err := savePairResult(store, agent.PairResult{ExecutorID: "executor-1"}); err == nil {
		t.Fatal("expected missing token to fail")
	}
}

type memoryCredentialStore struct {
	record  credential.Record
	loadErr error
}

func (s *memoryCredentialStore) Load() (credential.Record, error) {
	if s.loadErr != nil {
		return credential.Record{}, s.loadErr
	}
	return s.record, nil
}

func (s *memoryCredentialStore) Save(record credential.Record) error {
	s.record = record
	s.loadErr = nil
	return nil
}

func (s *memoryCredentialStore) Clear() error {
	s.record = credential.Record{}
	s.loadErr = credential.ErrNotFound
	return nil
}
