package task9runtime

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
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
	if config.ProcessingRules != "RULES" || config.KnowledgeBase != "project-kb" || config.ModelConfig != "MODEL" {
		t.Fatalf("config=%+v", config)
	}
	encoded, _ := json.Marshal(config)
	for _, forbidden := range []string{"apiKey", "secret-canary", "providerToken", "secret-token", "password", "credentials"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("settings snapshot leaked %q: %s", forbidden, encoded)
		}
	}
	if !strings.Contains(config.ProjectConfig, "CHARACTERS") || !strings.Contains(config.ProjectConfig, "SCENES") || !strings.Contains(config.UserConfig, `"profile_id":42`) || !strings.Contains(config.UserConfig, `"processing_rule_prompt_ref":"rules-v3"`) {
		t.Fatalf("config metadata=%+v", config)
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
