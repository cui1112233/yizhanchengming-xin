package storage

import (
	"strings"
	"testing"
)

func TestV11BookAssetsMigrationIsAdditiveAndDurable(t *testing.T) {
	migrations := V11Migrations()
	if len(migrations) < 13 {
		t.Fatalf("migrations=%d", len(migrations))
	}
	assetMigration := migrations[11]
	if assetMigration.Version != 1100012 || assetMigration.CallbackChecksum != "batch-factory-v11-book-assets-v1" {
		t.Fatalf("asset migration=%+v", assetMigration)
	}
	if got := assetMigration.LegacyChecksums; len(got) != 1 || got[0] != "b5e1e45e1ae47240a5d303efe70ad8b29afa21d906b482774d696c8424600624" {
		t.Fatalf("legacy book-assets checksum=%#v", got)
	}
	if assetMigration.Reconcile == nil {
		t.Fatal("book-assets migration must repair the verified historical ledger/table split before later foreign keys run")
	}
	var detailsMigration Migration
	for _, candidate := range migrations {
		if candidate.Version == 1100024 {
			detailsMigration = candidate
			break
		}
	}
	if detailsMigration.Version != 1100024 || !strings.Contains(strings.ToLower(strings.Join(detailsMigration.SQL, "\n")), "formal_name") || !strings.Contains(strings.ToLower(strings.Join(detailsMigration.SQL, "\n")), "relationships") {
		t.Fatalf("structured asset details migration=%+v", detailsMigration)
	}
	if migrations[15].Version != 1100016 || migrations[15].Adopt == nil {
		t.Fatal("book-merge migration must adopt only a verified pre-existing book_id column")
	}
	if migrations[19].Version != 1100020 || migrations[19].Adopt == nil {
		t.Fatal("merge-progress migration must adopt only its complete verified column set")
	}
	sql := strings.ToLower(strings.Join(V11BookAssetsStatements(), "\n"))
	for _, required := range []string{"batch_factory_v11_book_assets", "owner_username", "book_id", "kind", "prompt", "source", "revision"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("asset migration misses %q: %s", required, sql)
		}
	}
}

func TestV11BookAssetImageOverridesMigrationIsAdditive(t *testing.T) {
	var imageOverridesMigration Migration
	for _, candidate := range V11Migrations() {
		if candidate.Version == 1100027 {
			imageOverridesMigration = candidate
			break
		}
	}
	if imageOverridesMigration.Version != 1100027 || imageOverridesMigration.CallbackChecksum != "batch-factory-v11-book-asset-image-overrides-v1" {
		t.Fatalf("image override migration=%+v", imageOverridesMigration)
	}
	sql := strings.ToLower(strings.Join(imageOverridesMigration.SQL, "\n"))
	for _, required := range []string{"batch_factory_v11_book_assets", "image_model_id", "image_aspect_ratio", " null"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("image override migration misses %q: %s", required, sql)
		}
	}
}
