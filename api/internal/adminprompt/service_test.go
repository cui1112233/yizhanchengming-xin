package adminprompt

import (
	"context"
	"errors"
	"testing"
)

func TestDraftIsNotResolvableAndPublishArchivesPriorVersionWithRedactedAudit(t *testing.T) {
	store := NewMemoryStore(Version{ID: 1, Key: "script.default", Version: 1, Content: "published", Lifecycle: Published})
	service := NewService(store)
	draft, err := service.CreateDraft(context.Background(), 7, "script.default", "draft body", "server-request")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ResolveRuntime(context.Background(), "script.default"); err != nil {
		t.Fatal(err)
	}
	if err := service.Publish(context.Background(), 7, "script.default", draft.Version, "server-request"); err != nil {
		t.Fatal(err)
	}
	versions := store.Versions("script.default")
	if versions[0].Lifecycle != Archived || versions[1].Lifecycle != Published {
		t.Fatalf("versions=%+v", versions)
	}
	audit := store.Audits()[1]
	if audit.ContentSHA256 == "" || audit.Summary == "draft body" {
		t.Fatalf("audit leaked or missed hash: %+v", audit)
	}
}

func TestPublishRollsBackWhenAuditFails(t *testing.T) {
	store := NewMemoryStore(Version{ID: 1, Key: "script.default", Version: 1, Content: "old", Lifecycle: Published})
	draft, _ := NewService(store).CreateDraft(context.Background(), 7, "script.default", "new", "server-request")
	store.auditErr = errors.New("audit unavailable")
	if err := NewService(store).Publish(context.Background(), 7, "script.default", draft.Version, "server-request"); err == nil {
		t.Fatal("expected audit failure")
	}
	if got, _ := NewService(store).ResolveRuntime(context.Background(), "script.default"); got.Content != "old" {
		t.Fatalf("runtime=%+v", got)
	}
}

func TestRestoreCreatesHigherPublishedVersion(t *testing.T) {
	store := NewMemoryStore(Version{ID: 1, Key: "script.default", Version: 1, Content: "old", Lifecycle: Archived}, Version{ID: 2, Key: "script.default", Version: 2, Content: "new", Lifecycle: Published})
	restored, err := NewService(store).Restore(context.Background(), 7, "script.default", 1, "server-request")
	if err != nil || restored.Version != 3 || restored.Content != "old" || restored.Lifecycle != Published {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}
}
