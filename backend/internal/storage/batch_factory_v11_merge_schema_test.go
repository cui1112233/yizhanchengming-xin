package storage

import (
	"strings"
	"testing"
)

func TestV11MergeMigrationAddsDurableMergeJobs(t *testing.T) {
	joined := strings.ToLower(strings.Join(V11MergeStatements(), "\n"))
	for _, table := range []string{
		"batch_factory_v11_merge_jobs",
		"batch_factory_v11_merge_sources",
	} {
		if !strings.Contains(joined, table) {
			t.Fatalf("missing merge table %s", table)
		}
	}
	for _, clause := range []string{
		"unique key uq_bfv11_merge_request",
		"foreign key (batch_id) references batch_factory_v11_batches(id)",
		"foreign key (job_id) references batch_factory_v11_merge_jobs(id)",
	} {
		if !strings.Contains(joined, clause) {
			t.Fatalf("missing merge durability clause %s", clause)
		}
	}
}

func TestV11MigrationsRegisterMergeAfterProduction(t *testing.T) {
	migrations := V11Migrations()
	if len(migrations) < 8 || migrations[5].Version != 1100006 || migrations[6].Version != 1100007 {
		t.Fatalf("merge migration missing: %+v", migrations)
	}
	if migrations[5].CallbackChecksum != "batch-factory-v11-merge-v1" || migrations[6].CallbackChecksum != "batch-factory-v11-merge-poller-v1" {
		t.Fatalf("unexpected merge checksum: %+v", migrations[6])
	}
	if migrations[7].Version != 1100008 || migrations[7].CallbackChecksum != "batch-factory-v11-external-publish-v1" {
		t.Fatalf("external publish migration missing: %+v", migrations[7])
	}
}

func TestV11MergeAudioMigrationPersistsDeduplicationInputs(t *testing.T) {
	joined := strings.ToLower(strings.Join(V11MergeAudioStatements(), "\n"))
	for _, required := range []string{"audio_source", "audio_artifact_id", "alter table batch_factory_v11_merge_jobs"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("merge audio migration misses %q: %s", required, joined)
		}
	}
}
