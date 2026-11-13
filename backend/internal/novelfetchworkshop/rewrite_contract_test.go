package novelfetchworkshop

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func plainRewriteTestStore(owner string) *MemoryStore {
	store := NewMemoryStore()
	_ = store.PutConfig(context.Background(), owner, map[string]any{"appConfig": map[string]any{"rewrite": map[string]any{"default_template_id": "test_plain"}}, "knowledgeLibraries": map[string]any{"rewrite_templates": map[string]any{"profiles": []any{map[string]any{"id": "test_plain", "prompt": "保持原文", "response_mode": "plain_text"}}}}})
	return store
}

type namedRewriteTestProvider struct{ rewriteProviderFunc }

func (namedRewriteTestProvider) ModelIdentity() string { return "actual-model" }

func TestRewritePersistsActualModelAndVersionEvidence(t *testing.T) {
	ctx := context.Background()
	store := plainRewriteTestStore("owner")
	_ = store.PutDocument(ctx, "owner", Document{BookID: "123", Original: "原文"})
	provider := namedRewriteTestProvider{rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) { return "改写正文", nil })}
	_, err := GenerateAIRewrites(ctx, store, "owner", "123", 1, provider)
	if err != nil {
		t.Fatal(err)
	}
	doc, _ := store.GetDocument(ctx, "owner", "123")
	if doc.Meta["rewriteModel"] != "actual-model" || doc.Meta["aiCurrentVersion"] != "ai1" {
		t.Fatalf("missing execution identity: %#v", doc.Meta)
	}
	evidence := doc.Meta["rewriteEvidenceByVersion"].(map[string]any)["ai1"].(map[string]any)
	if evidence["model"] != "actual-model" {
		t.Fatalf("version lost model: %#v", evidence)
	}
}

func TestRewriteTemplateContractAndParsedBody(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	_ = store.PutConfig(ctx, "a", map[string]any{"appConfig": map[string]any{"rewrite": map[string]any{"default_template_id": "tpl", "process_line_count": 1}}, "knowledgeLibraries": map[string]any{"rewrite_templates": map[string]any{"profiles": []any{map[string]any{"id": "tpl", "system_prompt": "模板系统", "user_prompt_template": "{content_id}|{target_lines_text}|{anchor_lines_text}|{extra_instruction}", "response_mode": "changed_lines_json"}}}}})
	_ = store.PutDocument(ctx, "a", Document{BookID: "123", Original: "首行原文\n后续保留"})
	calls := 0
	doc, err := GenerateAIRewrites(ctx, store, "a", "123", 1, rewriteProviderFunc(func(_ context.Context, req RewriteRequest) (string, error) {
		calls++
		if req.SystemPrompt != "模板系统" || !strings.Contains(req.UserPrompt, "123|首行原文|后续保留") || strings.Contains(req.UserPrompt, "{content_id}") {
			t.Fatalf("template not sent: %#v", req)
		}
		return `{"id":"123","changed_lines":[{"line_no":1,"text":"改后首行"}]}`, nil
	}))
	if err != nil || calls != 1 || doc.Versions["ai1"] != "改后首行\n后续保留" {
		t.Fatalf("doc=%#v err=%v calls=%d", doc, err, calls)
	}
}

func TestRewriteRejectsWrongIDAndDoesNotSavePollution(t *testing.T) {
	for _, raw := range []string{`{"id":"other","changed_lines":[{"line_no":1,"text":"改文"}]}`, `{"id":"123","changed_lines":[{"line_no":2,"text":"改文"}]}`, `{"id":"123","changed_lines":[{"line_no":1,"text":"抱歉，我无法协助"}]}`, `{"id":"123","changed_lines":[]} trailing`} {
		if _, err := parseRewriteOutput(raw, "changed_lines_json", "123", 1); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	ctx := context.Background()
	store := NewMemoryStore()
	_ = store.PutDocument(ctx, "a", Document{BookID: "123", Original: "原文"})
	calls := 0
	_, err := GenerateAIRewrites(ctx, store, "a", "123", 1, rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) { calls++; return "抱歉，我无法协助", nil }))
	doc, _ := store.GetDocument(ctx, "a", "123")
	if err == nil || calls != 3 || len(doc.Versions) != 0 {
		t.Fatalf("err=%v calls=%d doc=%#v", err, calls, doc)
	}
	if doc.Meta["rewriteValidationHold"] != true {
		t.Fatal("validation failure must pause automatic generation")
	}
}

func TestImportedRewriteTemplateIsActuallyUsed(t *testing.T) {
	settings := DefaultWorkshopSettings()
	plan := rewritePlan(settings, "ai1", 1)
	if plan.TemplateName == "" || plan.SystemPrompt == "" || plan.UserTemplate == "" || plan.ResponseMode != "changed_lines_json" {
		t.Fatalf("imported template not connected: %q %q", plan.TemplateID, plan.TemplateName)
	}
	prompt, _, chars := buildRewritePrompt(Document{BookID: "book", Gender: "女频"}, "中文原文", 1, plan)
	if chars != 4 || strings.Contains(prompt, "{content_id}") || !strings.Contains(prompt, "中文原文") {
		t.Fatalf("template substitution/chinese size broken: chars=%d", chars)
	}
}

func TestRewriteKeepsSuccessfulVersionWhenNextFails(t *testing.T) {
	ctx := context.Background()
	store := plainRewriteTestStore("a")
	_ = store.PutDocument(ctx, "a", Document{BookID: "123", Original: "原文"})
	calls := 0
	_, err := GenerateAIRewrites(ctx, store, "a", "123", 2, rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		calls++
		if calls == 1 {
			return "已完成改文", nil
		}
		return "", errors.New("upstream down")
	}))
	doc, _ := store.GetDocument(ctx, "a", "123")
	if err == nil || doc.Versions["ai1"] != "已完成改文" || doc.Versions["ai2"] != nil {
		t.Fatalf("lost completed version doc=%#v err=%v", doc, err)
	}
}
