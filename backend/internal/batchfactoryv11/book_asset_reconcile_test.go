package batchfactoryv11

import (
	"context"
	"testing"
	"time"
)

func TestSplitAssetNameCoreAndNote(t *testing.T) {
	cases := []struct {
		input    string
		wantCore string
		wantNote string
	}{
		{"前女友", "前女友", ""},
		{"前女友（林薇）", "前女友", "林薇"},
		{"儿子 (小冰山)", "儿子", "小冰山"},
		{"丈夫（秦衍）（总裁）", "丈夫（秦衍）", "总裁"},
		{"（只有括号）", "（只有括号）", ""},
	}
	for _, tc := range cases {
		core, note := splitAssetName(tc.input)
		if core != tc.wantCore || note != tc.wantNote {
			t.Fatalf("splitAssetName(%q) = (%q,%q), want (%q,%q)", tc.input, core, note, tc.wantCore, tc.wantNote)
		}
	}
}

func TestReconcileDirectorAssetsFreezesMatchedRow(t *testing.T) {
	existing := []existingAssetRow{
		{ID: "asset-old-ex", Kind: "character", Name: "前女友", Source: "director"},
	}
	desired := []BookAsset{
		{Kind: "character", Name: "前女友（林薇）", Prompt: "AI擅自重写的描述"},
	}
	plan := reconcileDirectorAssets(existing, desired, map[string]bool{})
	// Same person: reuse the old row, but its saved prompt is frozen — there is
	// no update payload carrying the AI's reworded text.
	if len(plan.MatchedIDs) != 1 || plan.MatchedIDs[0] != "asset-old-ex" {
		t.Fatalf("renamed asset must match the old row: %+v", plan)
	}
	if len(plan.Inserts) != 0 || len(plan.Deletes) != 0 {
		t.Fatalf("matched asset must not insert or delete: %+v", plan)
	}
}

func TestReconcileDirectorAssetsPreservesExistingRowsWhenOutputOmitsThem(t *testing.T) {
	existing := []existingAssetRow{
		{ID: "asset-keep", Kind: "character", Name: "林溪", Source: "director"},
		{ID: "asset-stale", Kind: "character", Name: "路人甲", Source: "director"},
	}
	desired := []BookAsset{
		{Kind: "character", Name: "林溪", Prompt: "短发女主"},
	}
	plan := reconcileDirectorAssets(existing, desired, map[string]bool{})
	if len(plan.MatchedIDs) != 1 || plan.MatchedIDs[0] != "asset-keep" {
		t.Fatalf("kept asset missing: %+v", plan)
	}
	if len(plan.Deletes) != 0 {
		t.Fatalf("an incomplete asset output must not delete existing rows: %+v", plan)
	}
}

func TestReconcileDirectorAssetsKeepsStaleRowsWithImages(t *testing.T) {
	existing := []existingAssetRow{
		{ID: "asset-stale-image", Kind: "scene", Name: "旧花园", Source: "director"},
	}
	plan := reconcileDirectorAssets(existing, nil, map[string]bool{"asset-stale-image": true})
	if len(plan.Deletes) != 0 || len(plan.MatchedIDs) != 0 {
		t.Fatalf("stale rows with images must be kept: %+v", plan)
	}
}

func TestReconcileDirectorAssetsFreezesManualRows(t *testing.T) {
	existing := []existingAssetRow{
		{ID: "asset-manual", Kind: "character", Name: "前女友", Source: "manual"},
	}
	desired := []BookAsset{
		{Kind: "character", Name: "前女友（林薇）", Prompt: "AI又造的同名人"},
	}
	plan := reconcileDirectorAssets(existing, desired, map[string]bool{})
	// The manual row is recognised as the same person and frozen; no duplicate
	// insert and no prompt overwrite.
	if len(plan.MatchedIDs) != 1 || plan.MatchedIDs[0] != "asset-manual" {
		t.Fatalf("manual row must swallow the AI duplicate: %+v", plan)
	}
	if len(plan.Inserts) != 0 || len(plan.Deletes) != 0 {
		t.Fatalf("manual row must not insert or delete: %+v", plan)
	}
}

func TestReconcileDirectorAssetsReplacesDifferentNotesWhenNoImages(t *testing.T) {
	existing := []existingAssetRow{
		{ID: "asset-son-a", Kind: "character", Name: "儿子（小冰山）", Source: "director"},
	}
	desired := []BookAsset{
		{Kind: "character", Name: "儿子（小火山）", Prompt: "另一个儿子"},
	}
	plan := reconcileDirectorAssets(existing, desired, map[string]bool{})
	// Two different notes are never merged in place: with no image on the old
	// row it is replaced (delete + insert) so the ledger still holds one row
	// instead of stacking two.
	if len(plan.Inserts) != 1 || plan.Inserts[0].Name != "儿子（小火山）" {
		t.Fatalf("new different-note character must be inserted: %+v", plan)
	}
	if len(plan.Deletes) != 0 {
		t.Fatalf("different-note output must not delete the old row during a merge: %+v", plan)
	}

	// If the old row has a generated/uploaded image, it is protected and both
	// rows coexist as potentially different people.
	planImage := reconcileDirectorAssets(existing, desired, map[string]bool{"asset-son-a": true})
	if len(planImage.Deletes) != 0 || len(planImage.Inserts) != 1 {
		t.Fatalf("image-backed old row must be kept: %+v", planImage)
	}
}

func TestReconcileDirectorAssetsSkipsAmbiguousCoreMatches(t *testing.T) {
	existing := []existingAssetRow{
		{ID: "asset-ex1", Kind: "character", Name: "前女友", Source: "director"},
		{ID: "asset-ex2", Kind: "character", Name: "前女友（某人）", Source: "director"},
	}
	desired := []BookAsset{
		{Kind: "character", Name: "前女友（林薇）", Prompt: "林薇"},
	}
	plan := reconcileDirectorAssets(existing, desired, map[string]bool{})
	// Only the empty-note candidate qualifies, so the match is still unique.
	if len(plan.MatchedIDs) != 1 || plan.MatchedIDs[0] != "asset-ex1" {
		t.Fatalf("empty-note candidate should match uniquely: %+v", plan)
	}

	// Desired-side ambiguity: two renamed variants in one run cannot be bound
	// to a single frozen row, so neither matches in place; the imageless old row
	// is replaced once and both variants insert.
	plan2 := reconcileDirectorAssets(
		[]existingAssetRow{{ID: "asset-ex1", Kind: "character", Name: "前女友", Source: "director"}},
		[]BookAsset{
			{Kind: "character", Name: "前女友（林薇）", Prompt: "林薇"},
			{Kind: "character", Name: "前女友（某人）", Prompt: "某人"},
		},
		map[string]bool{},
	)
	if len(plan2.MatchedIDs) != 0 {
		t.Fatalf("ambiguous desired core must not match in place: %+v", plan2)
	}
	if len(plan2.Inserts) != 2 || len(plan2.Deletes) != 0 {
		t.Fatalf("ambiguous desired rows insert; old row remains available: %+v", plan2)
	}
}

func TestAddCoreNameLookupsUniqueAndAmbiguous(t *testing.T) {
	ids := map[string]string{"character\x00前女友": "id-exact"}
	addCoreNameLookups(ids, "character", [][2]string{{"前女友", "id-exact"}})
	// Exact key untouched; a core key is added so a variant name still binds.
	if ids["character\x00前女友"] != "id-exact" {
		t.Fatalf("exact key must not change: %q", ids["character\x00前女友"])
	}
	// Ambiguous core (two different rows): no guessed binding.
	ambiguous := map[string]string{}
	addCoreNameLookups(ambiguous, "character", [][2]string{{"儿子（小冰山）", "id-a"}, {"儿子（小火山）", "id-b"}})
	if _, ok := ambiguous["character\x00儿子"]; ok {
		t.Fatalf("ambiguous core must stay unindexed: %+v", ambiguous)
	}
}

// End-to-end through the memory store: a second director run that renames one
// character and drops another must neither grow the ledger nor change the
// frozen saved prompts.
func TestMemoryDirectorRunsDoNotStackOrRewriteAssets(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)

	first := []BookAsset{
		{Kind: "character", Name: "前女友", Prompt: "唯一有效的旧描述", Source: "director"},
		{Kind: "character", Name: "林溪", Prompt: "女主", Source: "director"},
	}
	for _, asset := range first {
		if _, err := store.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, CreateBookAssetInput{Kind: asset.Kind, Name: asset.Name, Prompt: asset.Prompt}); err != nil {
			t.Fatal(err)
		}
	}
	// Simulate what a second director run writes: one renamed (AI reworded the
	// prompt), one kept.
	reconcile := func(desired []BookAsset) {
		existingRows := []existingAssetRow{}
		hasImage := map[string]bool{}
		for assetID, owned := range store.bookAssets {
			asset := owned.Value
			if owned.Owner == "alice" && asset.BatchID == batch.ID && asset.BookID == book.ID {
				existingRows = append(existingRows, existingAssetRow{ID: assetID, Kind: asset.Kind, Name: asset.Name, Source: asset.Source})
			}
		}
		plan := reconcileDirectorAssets(existingRows, desired, hasImage)
		for _, id := range plan.Deletes {
			delete(store.bookAssets, id)
		}
		// Matched rows: deliberately untouched, no overwrite.
		for _, seed := range plan.Inserts {
			now := time.Now()
			seed.ID, seed.Revision, seed.CreatedAt, seed.UpdatedAt = store.id("asset"), 1, now, now
			store.bookAssets[seed.ID] = memoryOwned[BookAsset]{Owner: "alice", Value: seed}
		}
	}
	reconcile([]BookAsset{
		{Kind: "character", Name: "前女友（林薇）", Prompt: "AI擅自改写的新描述", Source: "director"},
		{Kind: "character", Name: "林溪", Prompt: "女主", Source: "director"},
	})

	var count int
	var frozenPrompt string
	for _, owned := range store.bookAssets {
		asset := owned.Value
		if owned.Owner == "alice" && asset.BatchID == batch.ID && asset.BookID == book.ID {
			count++
			if asset.Name == "前女友" {
				frozenPrompt = asset.Prompt
			}
			if asset.Name == "前女友（林薇）" {
				t.Fatalf("renamed variant must not be inserted: %+v", asset)
			}
		}
	}
	if count != 2 {
		t.Fatalf("asset ledger stacked to %d rows, want 2", count)
	}
	if frozenPrompt != "唯一有效的旧描述" {
		t.Fatalf("matched prompt was rewritten to %q", frozenPrompt)
	}
}
