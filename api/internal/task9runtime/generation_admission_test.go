package task9runtime

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestGenerationSnapshotNormalizesSelectionAndWhitelistsOptions(t *testing.T) {
	req := GenerationRequest{BatchProjectID: 7, BookIDs: []int64{9, 2, 9}, RequestID: "Case-Sensitive", RequestedByUserID: 4, HookEnabled: true}
	snap, err := normalizeGenerationRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snap.BookIDs, []int64{2, 9}) || snap.DirectorMode != "normal" || snap.ShotDurationLimitSec != 15 || snap.SelectionAll {
		t.Fatalf("normalized=%+v", snap)
	}
	b, _ := json.Marshal(snap)
	var fields map[string]any
	if err := json.Unmarshal(b, &fields); err != nil {
		t.Fatal(err)
	}
	want := []string{"schema_version", "batch_project_id", "book_ids", "selection_all", "hook_enabled", "plot_mode", "director_mode", "match_audio", "shot_duration_limit_sec", "requested_by_user_id"}
	if len(fields) != len(want) {
		t.Fatalf("snapshot includes unapproved fields: %s", b)
	}
	for _, k := range want {
		if _, ok := fields[k]; !ok {
			t.Fatalf("missing %s", k)
		}
	}
}

func TestGenerationAdmissionRejectsInvalidInputBeforeDatabase(t *testing.T) {
	base := GenerationRequest{BatchProjectID: 7, BookIDs: []int64{2}, RequestID: "key", RequestedByUserID: 4}
	for _, change := range []func(*GenerationRequest){
		func(r *GenerationRequest) { r.BatchProjectID = 0 }, func(r *GenerationRequest) { r.RequestID = "" },
		func(r *GenerationRequest) { r.BookIDs = []int64{-1} }, func(r *GenerationRequest) { r.DirectorMode = "unknown" },
		func(r *GenerationRequest) { r.ShotDurationLimitSec = 12 }, func(r *GenerationRequest) { r.RequestedByUserID = 0 },
	} {
		r := base
		change(&r)
		if _, err := NewMySQLStore(nil).AdmitGeneration(context.Background(), r); !errors.Is(err, ErrInvalidGenerationRequest) {
			t.Fatalf("request=%+v err=%v", r, err)
		}
	}
}

func TestGenerationSnapshotRejectsTamperingUnsupportedAndUnknownFields(t *testing.T) {
	snap := GenerationSnapshot{SchemaVersion: 1, BatchProjectID: 7, BookIDs: []int64{2}, DirectorMode: "normal", ShotDurationLimitSec: 15, RequestedByUserID: 4}
	b, hash, err := encodeGenerationSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeGenerationSnapshot(1, b, hash, 7, 4); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		version int
		body    []byte
		hash    string
	}{
		{2, b, hash}, {1, b, "incorrect"}, {1, []byte(`{"schema_version":1,"prompt":"secret"}`), hash},
	} {
		if _, err := decodeGenerationSnapshot(tc.version, tc.body, tc.hash, 7, 4); !errors.Is(err, ErrInvalidGenerationRequest) {
			t.Fatalf("invalid snapshot accepted: %s %v", tc.body, err)
		}
	}
}
