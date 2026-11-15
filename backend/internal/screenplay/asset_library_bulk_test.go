package screenplay

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestLibraryBulkAtomicReplayAndClearCAS(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryAssetLibraryRepository()
	a := AssetLibraryEntry{ID: "bulk-a", Owner: "alice", Type: AssetCharacter, Name: "无图", Snapshot: Asset{Extra: map[string]string{LibraryArchiveJSON: `{"id":"original","custom":{"x":1}}`, LibraryArchiveID: "original", LibraryMetadataOnly: "true"}}}
	b := a
	b.ID = "bulk-b"
	result, err := repo.ApplyAssetLibraryBulk(ctx, "alice", "request-one", "digest-one", "import", []AssetLibraryEntry{a, b}, nil)
	if err != nil || len(result.Entries) != 2 {
		t.Fatalf("import: %+v %v", result, err)
	}
	if result.Entries[0].Snapshot.Extra[LibraryArchiveID] == result.Entries[1].Snapshot.Extra[LibraryArchiveID] {
		t.Fatal("logical IDs collided")
	}
	replay, err := repo.ApplyAssetLibraryBulk(ctx, "alice", "request-one", "digest-one", "import", []AssetLibraryEntry{a, b}, nil)
	if err != nil || replay.Entries[0].ID != result.Entries[0].ID {
		t.Fatalf("replay: %v", err)
	}
	if _, err = repo.ApplyAssetLibraryBulk(ctx, "alice", "request-one", "different", "import", []AssetLibraryEntry{a}, nil); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
	if _, err = repo.ApplyAssetLibraryBulk(ctx, "alice", "clear-one", "clear", "clear", nil, []AssetLibraryExpected{{ID: a.ID, Revision: 1}}); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
	rows, _ := repo.ListAssetLibrary(ctx, "alice", AssetCharacter, "", "")
	if len(rows) != 2 {
		t.Fatal("partial clear")
	}
	expected := []AssetLibraryExpected{{ID: a.ID, Revision: 1}, {ID: b.ID, Revision: 1}}
	cleared, err := repo.ApplyAssetLibraryBulk(ctx, "alice", "clear-two", "clear2", "clear", nil, expected)
	if err != nil || cleared.ClearedCount != 2 {
		t.Fatal(cleared, err)
	}
	_, err = repo.ApplyAssetLibraryBulk(ctx, "alice", "clear-two", "clear2", "clear", nil, expected)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.SaveAssetLibraryEntry(ctx, a, 0); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("tombstone: %v", err)
	}
}

func TestLibraryBulkConcurrentClearRequiresFreshWholeSnapshot(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryAssetLibraryRepository()
	entry, err := repo.SaveAssetLibraryEntry(ctx, assetLibraryFixture(), 0)
	if err != nil {
		t.Fatal(err)
	}
	expected := []AssetLibraryExpected{{ID: entry.ID, Revision: entry.Revision}}
	var wg sync.WaitGroup
	out := make(chan error, 2)
	for _, key := range []string{"clear-one", "clear-two"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			_, err := repo.ApplyAssetLibraryBulk(ctx, "alice", key, key, "clear", nil, expected)
			out <- err
		}(key)
	}
	wg.Wait()
	close(out)
	success, conflicts := 0, 0
	for err := range out {
		if err == nil {
			success++
		} else if errors.Is(err, ErrRevisionConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatal(success, conflicts)
	}
}

func TestLibraryBulkMetadataOnlyDoesNotRelaxOrdinaryUploadValidation(t *testing.T) {
	repo := NewMemoryAssetLibraryRepository()
	entry := AssetLibraryEntry{ID: "empty-image", Owner: "alice", Type: AssetCharacter, Name: "missing"}
	if _, err := repo.SaveAssetLibraryEntry(context.Background(), entry, 0); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	entry.Snapshot.Extra = map[string]string{LibraryMetadataOnly: "true"}
	if _, err := repo.SaveAssetLibraryEntry(context.Background(), entry, 0); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	entry.Snapshot.Extra[LibraryArchiveJSON] = `{"name":"missing"}`
	entry.Type = AssetScene
	if _, err := repo.SaveAssetLibraryEntry(context.Background(), entry, 0); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
}

func TestLibraryBulkPendingIntentBindsRequestBeforeCommit(t *testing.T) {
	repo := NewMemoryAssetLibraryRepository()
	ctx := context.Background()
	if err := repo.ReserveAssetLibraryBulk(ctx, "alice", "pending-key", "first"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.LookupAssetLibraryBulk(ctx, "alice", "pending-key", "first"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := repo.ReserveAssetLibraryBulk(ctx, "alice", "pending-key", "changed"); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
	entry := assetLibraryFixture()
	result, err := repo.ApplyAssetLibraryBulk(ctx, "alice", "pending-key", "first", "import", []AssetLibraryEntry{entry}, nil)
	if err != nil || len(result.Entries) != 1 {
		t.Fatal(result, err)
	}
	if _, err := repo.LookupAssetLibraryBulk(ctx, "alice", "pending-key", "first"); err != nil {
		t.Fatal(err)
	}
}
