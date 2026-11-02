package httpapi

import "testing"

func TestMergeOutputMatchesArtifactAcceptsV11AndV12LocalURLs(t *testing.T) {
	for _, outputURL := range []string{
		"/api/batch-factory/v11/batches/batch-1/merge-media/merge_cover",
		"/api/batch-factory/v12/batches/batch-1/merge-media/merge_cover",
	} {
		if !mergeOutputMatchesArtifact("batch-1", "merge_cover", outputURL) {
			t.Fatalf("expected local merge output %q to match its cover artifact", outputURL)
		}
	}
	if mergeOutputMatchesArtifact("batch-1", "merge_cover", "/api/batch-factory/v12/batches/other/merge-media/merge_cover") {
		t.Fatal("a different batch must not authorize the cover artifact")
	}
}
