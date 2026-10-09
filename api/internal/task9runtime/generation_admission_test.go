package task9runtime

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type admissionServiceStore struct {
	result AdmissionResult
	err    error
	run    GenerationRunStatus
	admits int
}

func (s *admissionServiceStore) AdmitGeneration(context.Context, GenerationRequest) (AdmissionResult, error) {
	s.admits++
	return s.result, s.err
}
func (s *admissionServiceStore) GenerationRun(context.Context, int64, int64) (GenerationRunStatus, error) {
	return s.run, s.err
}

type admissionCoordinator struct {
	items []WorkItem
	fail  int
}

type mutableRuntimeReadiness struct{ ready bool }

func (r *mutableRuntimeReadiness) Ready() bool { return r.ready }

func (c *admissionCoordinator) Enqueue(_ context.Context, item WorkItem) error {
	c.items = append(c.items, item)
	if c.fail > 0 && len(c.items) == c.fail {
		return errors.New("queue unavailable")
	}
	return nil
}

func TestGenerationAdmissionServiceQueuesOnlyDispatchItemsAndKeepsStableTaskIDs(t *testing.T) {
	store := &admissionServiceStore{result: AdmissionResult{Run: RunRecord{ID: 9}, Items: []WorkItem{{BookRunID: 31, BookID: 2, Attempt: 1}}, TaskIDs: []int64{31, 32}, Created: false}}
	coordinator := &admissionCoordinator{}
	service := NewGenerationAdmissionService(store, coordinator, StaticRuntimeReadiness(true))
	result, err := service.AdmitGeneration(context.Background(), GenerationRequest{})
	if err != nil || result.Dispatch != "queued" || !reflect.DeepEqual(result.TaskIDs, []int64{31, 32}) || len(coordinator.items) != 1 {
		t.Fatalf("result=%+v queued=%+v err=%v", result, coordinator.items, err)
	}
}

func TestGenerationAdmissionServiceReturnsPendingAfterDurableQueueFailure(t *testing.T) {
	store := &admissionServiceStore{result: AdmissionResult{Run: RunRecord{ID: 9}, Items: []WorkItem{{BookRunID: 31, BookID: 2, Attempt: 1}, {BookRunID: 32, BookID: 3, Attempt: 1}}, TaskIDs: []int64{31, 32}, Created: true}}
	coordinator := &admissionCoordinator{fail: 2}
	service := NewGenerationAdmissionService(store, coordinator, StaticRuntimeReadiness(true))
	result, err := service.AdmitGeneration(context.Background(), GenerationRequest{})
	if err != nil || result.Dispatch != "pending" || len(coordinator.items) != 2 {
		t.Fatalf("result=%+v queued=%+v err=%v", result, coordinator.items, err)
	}
}

func TestGenerationAdmissionServiceFailsClosedBeforeStoreWhenNotReady(t *testing.T) {
	service := NewGenerationAdmissionService(&admissionServiceStore{}, &admissionCoordinator{}, StaticRuntimeReadiness(false))
	if _, err := service.AdmitGeneration(context.Background(), GenerationRequest{}); !errors.Is(err, ErrExecutorUnavailable) {
		t.Fatalf("err=%v", err)
	}
}

func TestGenerationAdmissionServiceReadsRuntimeReadinessForEveryAdmission(t *testing.T) {
	store := &admissionServiceStore{result: AdmissionResult{Run: RunRecord{ID: 9}}}
	readiness := &mutableRuntimeReadiness{}
	service := NewGenerationAdmissionService(store, &admissionCoordinator{}, readiness)

	if service.Ready() {
		t.Fatal("service reported ready before the supervised runtime started")
	}
	if _, err := service.AdmitGeneration(context.Background(), GenerationRequest{}); !errors.Is(err, ErrExecutorUnavailable) || store.admits != 0 {
		t.Fatalf("err=%v admits=%d", err, store.admits)
	}

	readiness.ready = true
	if !service.Ready() {
		t.Fatal("service did not observe the supervised runtime becoming ready")
	}
	if _, err := service.AdmitGeneration(context.Background(), GenerationRequest{}); err != nil || store.admits != 1 {
		t.Fatalf("err=%v admits=%d", err, store.admits)
	}

	readiness.ready = false
	if service.Ready() {
		t.Fatal("service did not observe the supervised runtime stopping")
	}
}

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
	want := []string{"schema_version", "batch_project_id", "book_ids", "selection_all", "hook_enabled", "plot_mode", "director_mode", "match_audio", "shot_duration_limit_sec", "requested_by_user_id", "action", "config"}
	if len(fields) != len(want) {
		t.Fatalf("snapshot includes unapproved fields: %s", b)
	}
	for _, k := range want {
		if _, ok := fields[k]; !ok {
			t.Fatalf("missing %s", k)
		}
	}
}

func TestGenerationSnapshotV2ActionsAreCanonicalAndStrict(t *testing.T) {
	full, err := normalizeGenerationRequest(GenerationRequest{BatchProjectID: 7, BookIDs: []int64{2}, RequestID: "full", RequestedByUserID: 4})
	if err != nil {
		t.Fatal(err)
	}
	if full.SchemaVersion != 2 || full.Action != GenerationActionFull || full.RetryStage != "" || full.SourceBookRunID != 0 {
		t.Fatalf("full snapshot=%+v", full)
	}
	retry, err := normalizeGenerationRequest(GenerationRequest{BatchProjectID: 7, BookIDs: []int64{2}, RequestID: "retry", RequestedByUserID: 4, Action: GenerationActionStageRetry, RetryStage: "DIRECTOR", SourceBookRunID: 81})
	if err != nil {
		t.Fatal(err)
	}
	if retry.SchemaVersion != 2 || retry.Action != GenerationActionStageRetry || retry.RetryStage != "DIRECTOR" || retry.SourceBookRunID != 81 {
		t.Fatalf("retry snapshot=%+v", retry)
	}
	body, hash, err := encodeGenerationSnapshot(retry)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeGenerationSnapshot(2, body, hash, 7, 4)
	if err != nil || !reflect.DeepEqual(decoded, retry) {
		t.Fatalf("decoded=%+v err=%v", decoded, err)
	}
	for _, mutate := range []func(*GenerationRequest){
		func(r *GenerationRequest) { r.Action = "unknown" },
		func(r *GenerationRequest) { r.Action = GenerationActionFull; r.RetryStage = "SCRIPT" },
		func(r *GenerationRequest) { r.Action = GenerationActionFull; r.SourceBookRunID = 1 },
		func(r *GenerationRequest) {
			r.Action = GenerationActionStageRetry
			r.RetryStage = ""
			r.SourceBookRunID = 1
		},
		func(r *GenerationRequest) {
			r.Action = GenerationActionStageRetry
			r.RetryStage = "UNKNOWN"
			r.SourceBookRunID = 1
		},
		func(r *GenerationRequest) {
			r.Action = GenerationActionStageRetry
			r.RetryStage = "SCRIPT"
			r.SourceBookRunID = 0
		},
		func(r *GenerationRequest) {
			r.Action = GenerationActionStageRetry
			r.RetryStage = "SCRIPT"
			r.SourceBookRunID = 1
			r.BookIDs = []int64{2, 3}
		},
	} {
		req := GenerationRequest{BatchProjectID: 7, BookIDs: []int64{2}, RequestID: "invalid", RequestedByUserID: 4}
		mutate(&req)
		if _, err := normalizeGenerationRequest(req); !errors.Is(err, ErrInvalidGenerationRequest) {
			t.Fatalf("invalid request accepted: %+v err=%v", req, err)
		}
	}
}

func TestGenerationSnapshotV1RemainsReadableAsFull(t *testing.T) {
	legacy := GenerationSnapshot{SchemaVersion: 1, BatchProjectID: 7, BookIDs: []int64{2}, DirectorMode: "normal", ShotDurationLimitSec: 15, RequestedByUserID: 4}
	body, hash, err := encodeGenerationSnapshot(legacy)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeGenerationSnapshot(1, body, hash, 7, 4)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Action != GenerationActionFull || decoded.SchemaVersion != 1 {
		t.Fatalf("legacy snapshot=%+v", decoded)
	}
}

func TestGenerationConfigSnapshotUsesOnlyAllowlistedServerSettings(t *testing.T) {
	project := []byte(`{"production":{"scriptWorkspace":{"constraints":"RULES","characters":"CHARACTERS","scenes":"SCENES","model":"MODEL","apiKey":"secret-canary"},"providerToken":"secret-token"},"knowledgePromptRef":"project-kb"}`)
	profile := []byte(`{"processingRulePromptRef":"rules-v3","knowledgePromptRef":"kb-v7","production":{"scriptWorkspace":{"constraints":"OLD"}},"credentials":{"password":"secret"}}`)
	config, err := generationConfigFromSettings(project, profile, 42, "女频短剧版", "v3")
	if err != nil {
		t.Fatal(err)
	}
	if config.Constraints != "RULES" || config.ProcessingRulePromptRef != "rules-v3" || config.KnowledgePromptRef != "project-kb" || config.Model != "MODEL" || config.Characters != "CHARACTERS" || config.Scenes != "SCENES" {
		t.Fatalf("config=%+v", config)
	}
	encoded, _ := json.Marshal(config)
	for _, forbidden := range []string{"apiKey", "secret-canary", "providerToken", "secret-token", "password", "credentials"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("settings snapshot leaked %q: %s", forbidden, encoded)
		}
	}
}

func TestGenerationConfigSnapshotRejectsOversizedFields(t *testing.T) {
	project := []byte(`{"production":{"scriptWorkspace":{"constraints":"` + strings.Repeat("x", generationConfigFieldLimit+1) + `"}}}`)
	if _, err := generationConfigFromSettings(project, nil, 0, "", ""); !errors.Is(err, ErrInvalidGenerationRequest) {
		t.Fatalf("oversized settings err=%v", err)
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
		{3, b, hash}, {1, b, "incorrect"}, {1, []byte(`{"schema_version":1,"prompt":"secret"}`), hash},
	} {
		if _, err := decodeGenerationSnapshot(tc.version, tc.body, tc.hash, 7, 4); !errors.Is(err, ErrInvalidGenerationRequest) {
			t.Fatalf("invalid snapshot accepted: %s %v", tc.body, err)
		}
	}
}
