package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func TestNativePatchRevisionRealMySQL(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("patch-review")
	localAudit := strings.Replace(failedAuditJSON, `"category":"plot"`, `"category":"plot","repair_scope":"local"`, 1)
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, localAudit, nativeFixturePatch, passedAuditJSON, settlementJSON, memoryJSON}}
	store := NewMySQLStore(db)
	s := NewService(store, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	p, e := s.CreateProject(ctx, owner, CreateProjectInput{Title: "原版补丁持久化验证", Genre: "悬疑", ModelID: "fake-only", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	session, e := s.CreateSession(ctx, owner, CreateSessionInput{ProjectID: p.ID, Mode: "long_original", Title: "修订"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.UpdateProjectSettings(ctx, owner, p.ID, ProjectSettings{Agent: AgentSettings{Version: 1, PauseAfterBlueprint: boolPointer(false), PassScore: 80, MaxRevisionRounds: 1}})
	if e != nil {
		t.Fatal(e)
	}
	r, e := s.SendMessage(ctx, owner, SendMessageInput{ProjectID: p.ID, SessionID: session.ID, Content: "写第一章", ModelID: "fake-only", IdempotencyKey: "patch-effect"})
	if e != nil {
		t.Fatal(e)
	}
	r, e = s.Run(ctx, owner, r.ID)
	if e != nil || r.Status != RunSucceeded {
		t.Fatalf("%s %v", r.Status, e)
	}
	chapters, e := store.ListChapters(ctx, owner, p.ID)
	if e != nil || len(chapters) != 1 || !strings.Contains(chapters[0].Content, "指纹") {
		t.Fatal("patch not saved", e)
	}
	var snapshot []byte
	if e = db.QueryRowContext(ctx, "SELECT snapshot_json FROM novel_revisions WHERE owner_username=? AND run_id=? LIMIT 1", owner, r.ID).Scan(&snapshot); e != nil {
		t.Fatal(e)
	}
	var saved struct {
		Native nativePatchResult `json:"native_revision"`
		Raw    string            `json:"raw_output"`
	}
	if e = json.Unmarshal(snapshot, &saved); e != nil || saved.Native.Applied != 1 || saved.Raw != nativeFixturePatch {
		t.Fatal("revision evidence not persisted", e, string(snapshot))
	}
	if saved.Native.Route == nil || saved.Native.Route.Mode != "patch-only" {
		t.Fatalf("explicit local scope was lost: %s", snapshot)
	}
	var savedAudit []byte
	if e = db.QueryRowContext(ctx, "SELECT snapshot_json FROM novel_audits WHERE owner_username=? AND run_id=? ORDER BY created_at,audit_id LIMIT 1", owner, r.ID).Scan(&savedAudit); e != nil {
		t.Fatal(e)
	}
	var storedAudit chapterAuditOutput
	if e = json.Unmarshal(savedAudit, &storedAudit); e != nil || len(storedAudit.Issues) != 1 || string(storedAudit.Issues[0].RepairScopeSnake) != `"local"` {
		t.Fatal("audit scope not durable", e, string(savedAudit))
	}
	if _, e = s.Run(ctx, owner, r.ID); e != nil || len(provider.requests) != 8 {
		t.Fatal("repeated completed run generated again", e, len(provider.requests))
	}
	t.Log("8 fake-provider calls: explicit local scope -> patch-only request -> PATCHES -> re-audit -> durable MySQL chapter/audit scope/revision route -> repeat did not regenerate; 0 paid calls")
}

const nativeFixturePatch = "=== FIXED_ISSUES ===\n补上指纹线索。\n=== PATCHES ===\n--- PATCH 1 ---\nTARGET_TEXT:\n林澜拆开了那封没有寄件人的信。\nREPLACEMENT_TEXT:\n林澜拆开了那封没有寄件人的信，信封上留下失踪者的指纹。\n--- END PATCH ---\n=== UPDATED_STATE ===\n发现指纹线索"

func TestNativeRevisionPipeline(t *testing.T) {
	for _, mode := range []string{"patch", "whole", "invalid", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			runtime := NewMemoryRuntime()
			input := fixturePipelineRunContext(runtime)
			input.ReviewCycleVersion = 2
			reply := nativeFixturePatch
			if mode == "whole" {
				reply = "=== FIXED_ISSUES ===\n补上指纹线索。\n=== REVISED_CONTENT ===\n雨水敲着窗。林澜找到失踪者的指纹。"
			}
			if mode == "invalid" {
				reply = "=== FIXED_ISSUES ===\n失败\n=== PATCHES ===\n无法匹配"
			}
			provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, reply, passedAuditJSON, settlementJSON, memoryJSON}}
			if mode == "cancel" {
				provider.afterCall = func(n int) {
					if n == 4 {
						_ = runtime.Cancel(context.Background(), input.Run.ID)
					}
				}
			}
			bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
			if mode == "invalid" || mode == "cancel" {
				if err == nil || len(provider.requests) != 5 || bundle.Chapter.Content != "" {
					t.Fatalf("failed/cancelled revision escaped: %v calls=%d", err, len(provider.requests))
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(provider.requests) != 8 || len(bundle.Revisions) != 1 || !strings.Contains(bundle.Chapter.Content, "指纹") {
				t.Fatalf("missing output %+v", bundle)
			}
			if !provider.requests[4].DisableJSONResponse || !strings.Contains(provider.requests[4].SystemPrompt, "=== PATCHES ===") {
				t.Fatal("native protocol not sent")
			}
			if !strings.Contains(provider.requests[5].UserPrompt, "指纹") {
				t.Fatal("re-audit did not receive modified content")
			}
			if !json.Valid(bundle.Revisions[0].Snapshot) || !strings.Contains(string(bundle.Revisions[0].Snapshot), "raw_output") {
				t.Fatal("raw revision missing")
			}
		})
	}
}

func TestNativePatches(t *testing.T) {
	cases := []struct {
		name, content, target, replacement, want string
		applied                                  int
	}{
		{"exact", "甲。雨水敲着窗。乙。", "雨水敲着窗。", "风停了。", "甲。风停了。乙。", 1},
		{"repeat", "雨。雨。", "雨。", "雪。", "雨。雨。", 0},
		{"deletion", "甲。乙。", "乙。", "", "甲。", 1},
		{"whitespace", "前文。她走到门口，\n  轻轻敲了三下门。后文。", "她走到门口， 轻轻敲了三下门。", "她转身离开。", "前文。她转身离开。后文。", 1},
		{"short fuzzy refused", "甲\n乙", "甲 乙", "丙", "甲\n乙", 0},
		{"fuzzy repeat refused", "abcdefghij\nkl。abcdefghij kl", "abcdefghij\tkl", "丙", "abcdefghij\nkl。abcdefghij kl", 0},
		{"empty target", "正文", "", "危险", "正文", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := applyNativePatches(c.content, []nativePatch{{c.target, c.replacement}})
			if r.Content != c.want || r.Applied != c.applied {
				t.Fatalf("%+v", r)
			}
		})
	}
	r := applyNativePatches("甲。乙。", []nativePatch{{"甲。", "丙。"}, {"不存在", "丁。"}, {"甲。", "戊。"}})
	if r.Content != "丙。乙。" || r.Applied != 1 || r.Skipped != 2 {
		t.Fatalf("partial/sequential %+v", r)
	}
}

func TestNativeRevisionSections(t *testing.T) {
	raw := "=== FIXED_ISSUES ===\n修正动作。\n=== PATCHES ===\n--- PATCH 1 ---\nTARGET_TEXT:\n雨水敲着窗。\nREPLACEMENT_TEXT:\n风停了。\n--- END PATCH ---\n=== UPDATED_STATE ===\n窗外风停"
	v, s, e := decodeNativeRevision(raw, "雨水敲着窗。林澜拆信。", nil)
	if e != nil || v.Content != "风停了。林澜拆信。" || s.Applied != 1 || s.UpdatedState != "窗外风停" {
		t.Fatalf("%+v %+v %v", v, s, e)
	}
	for _, bad := range []string{"", "=== FIXED_ISSUES ===\n修正\n=== REVISED_CONTENT ===\n", raw + "\n=== REVISED_CONTENT ===\n冲突正文", "=== REVISED_CONTENT ===\n无说明正文", "=== FIXED_ISSUES ===\n修正\n=== PATCHES ===\n--- PATCH 1 ---\nTARGET_TEXT:\n不存在\nREPLACEMENT_TEXT:\n新文\n--- END PATCH ---"} {
		if _, _, err := decodeNativeRevision(bad, "正文", nil); err == nil {
			t.Fatal("accepted invalid", bad)
		}
	}
}
