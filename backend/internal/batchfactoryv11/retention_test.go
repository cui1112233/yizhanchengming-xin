package batchfactoryv11

import (
	"testing"
	"time"
)

func TestRetentionExpiryUsesConfiguredDays(t *testing.T) {
	completed := time.Date(2026, 11, 1, 8, 30, 0, 0, time.UTC)
	for _, days := range []int{3, 7, 14, 30} {
		want := completed.AddDate(0, 0, days)
		if got := retentionExpiry(completed, days); !got.Equal(want) {
			t.Fatalf("retentionExpiry(%d)=%s want=%s", days, got, want)
		}
	}
}

func TestRetentionExpiryFallsBackToDefaultWindow(t *testing.T) {
	completed := time.Date(2026, 11, 1, 8, 30, 0, 0, time.UTC)
	want := completed.AddDate(0, 0, DefaultLocalMergedArtifactRetentionDays)
	if got := retentionExpiry(completed, 0); !got.Equal(want) {
		t.Fatalf("retentionExpiry(0)=%s want=%s", got, want)
	}
}

func TestMergeMediaStorageRefAcceptsOnlyGeneratedMergeVideos(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		ok   bool
	}{
		{"/api/batch-factory/v12/batches/batch_1/merge-media/merge_abc123.mp4", "merge_abc123.mp4", true},
		{"/api/batch-factory/v11/batches/batch_1/merge-media/merge_abc123", "merge_abc123.mp4", true},
		{"https://example.com/video/final.mp4", "", false},
		{"/api/batch-factory/v12/batches/batch_1/merge-media/other_abc.mp4", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := mergeMediaStorageRef(tc.raw)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("mergeMediaStorageRef(%q)=(%q,%v) want=(%q,%v)", tc.raw, got, ok, tc.want, tc.ok)
		}
	}
}

func TestMergeTOSObjectKeyRestoresUploadedObjectKey(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		ok   bool
	}{
		{"https://cdn.example.com/batch-merged/merge_task_1.mp4", "batch-merged/merge_task_1.mp4", true},
		{"https://cdn.example.com/batch-uploaded/upload_1.mp4", "batch-uploaded/upload_1.mp4", true},
		{"https://cdn.example.com/production-media/comic/image.png", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := mergeTOSObjectKey(tc.raw)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("mergeTOSObjectKey(%q)=(%q,%v) want=(%q,%v)", tc.raw, got, ok, tc.want, tc.ok)
		}
	}
}

func TestLocalExecutorArtifactIDExtractsArtifactIdentity(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		ok   bool
	}{
		{"http://127.0.0.1:4001/api/local-executor/v1/artifacts/art_abc?token=x", "art_abc", true},
		{"https://example.com/api/local-executor/v1/artifacts/art_x1y2", "art_x1y2", true},
		{"https://example.com/api/batch-factory/v12/batches/b1/merge-media/merge_a.mp4", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := parseLocalExecutorArtifactID(tc.raw)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("parseLocalExecutorArtifactID(%q)=(%q,%v) want=(%q,%v)", tc.raw, got, ok, tc.want, tc.ok)
		}
	}
}
