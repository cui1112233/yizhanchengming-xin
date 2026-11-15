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
	prompt, _, chars, _ := buildRewritePrompt(Document{BookID: "book", Gender: "女频"}, "中文原文", 1, plan)
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

const rewriteQualitySource = "她推开旧书店的门，风铃响了一声。柜台后面的人抬起头，两人对视了片刻，谁也没有先开口。窗外下着小雨，街上的行人匆匆走过。她把手里的信放在柜台上，指尖还在发抖。他终于伸手拿起那封信，慢慢拆开信封，看完之后沉默了很久。雨声渐密，店里的老式挂钟敲了六下。"

func TestRewriteRejectsPromptLeakAndRepeatedParagraph(t *testing.T) {
	outputs := []string{
		"待处理片段：" + rewriteQualitySource,
		"这是第 1 个改写版本。她推门而入，风铃轻响。",
		strings.Join([]string{
			"雨水顺着屋檐落下，街灯在风里轻轻摇晃。",
			"雨水顺着屋檐落下，街灯在风里轻轻摇晃。",
			"雨水顺着屋檐落下，街灯在风里轻轻摇晃。",
		}, "\n"),
	}
	for _, output := range outputs {
		calls := 0
		_, err := completeValidatedRewrite(context.Background(), rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) {
			calls++
			return output, nil
		}), RewriteRequest{MaxTokens: 2400}, "book", "plain_text", 1, 120, rewriteQualitySource)
		if !errors.Is(err, ErrRewriteValidation) || calls != 3 {
			t.Fatalf("accepted polluted output %q: calls=%d err=%v", output, calls, err)
		}
	}
}

func TestRewriteRejectsUnchangedOriginalAndAcceptsRealRewrite(t *testing.T) {
	ctx := context.Background()
	store := plainRewriteTestStore("a")
	_ = store.PutDocument(ctx, "a", Document{BookID: "123", Original: rewriteQualitySource})
	calls := 0
	_, err := GenerateAIRewrites(ctx, store, "a", "123", 1, rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		calls++
		return rewriteQualitySource, nil
	}))
	doc, _ := store.GetDocument(ctx, "a", "123")
	if !errors.Is(err, ErrRewriteValidation) || calls != 3 || len(doc.Versions) != 0 {
		t.Fatalf("unchanged original accepted: calls=%d err=%v versions=%#v", calls, err, doc.Versions)
	}
	if _, err := GenerateAIRewrites(ctx, store, "a", "123", 1, rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		return "雨点敲打玻璃时，老周从账本后站起，望见门口那个撑伞的身影。两人隔着书架沉默几秒。她把一封折好的信推过去，手背青筋绷紧。老周读完，靠回椅背，很久没有说话。", nil
	})); err != nil {
		t.Fatalf("genuine rewrite rejected: %v", err)
	}
}

func TestRewriteSimilarityIgnoresPunctuationOnlyChanges(t *testing.T) {
	if err := validateRewriteQuality(strings.ReplaceAll(rewriteQualitySource, "。", "！"), rewriteQualitySource); err == nil {
		t.Fatal("punctuation-only rewrite must be rejected when source is substantial")
	}
}
