package novelcreation

import (
	"reflect"
	"strings"
	"testing"
)

// This catches definitions disappearing, becoming incomplete, or moving their
// system prompt out of the backend-owned contract.
func TestDefinitionsKeepSystemPromptsInBackend(t *testing.T) {
	definitions := All()
	if len(definitions) != 7 {
		t.Fatalf("definitions = %d, want 7", len(definitions))
	}

	seen := map[string]struct{}{}
	for _, def := range definitions {
		if strings.TrimSpace(def.ID) == "" || def.Version < 1 || strings.TrimSpace(def.SystemPrompt) == "" {
			t.Fatalf("invalid definition: %+v", def)
		}
		if len(def.RequiredInputs) == 0 || len(def.Errors) == 0 {
			t.Fatalf("incomplete definition: %+v", def)
		}
		if !strings.Contains(def.OutputContract, "JSON") {
			t.Fatalf("%s missing JSON contract", def.ID)
		}
		if _, exists := seen[def.ID]; exists {
			t.Fatalf("duplicate definition ID %q", def.ID)
		}
		seen[def.ID] = struct{}{}
	}
}

// This catches a response schema changing without updating the pipeline that
// will parse each model response. It locks identity, pipeline order, inputs,
// stable errors, exact JSON structures, and the no-extra-output rule.
func TestDefinitionsExposeExactStableContracts(t *testing.T) {
	want := []definitionContract{
		{
			id: "novel-creation.foundation-plan", inputs: []string{"project_brief", "user_message"},
			errors: map[string]string{"input_required": "创作要求不能为空", "invalid_output": "基础设定 JSON 格式无效", "upstream_failed": "基础设定生成失败"},
			output: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"title":"string","genre":"string","world":"string","characters":[{"name":"string","role":"string","goal":"string","conflict":"string"}],"coreConflict":"string","volumes":[{"number":"integer","title":"string","goal":"string","chapterRange":["integer","integer"]}]}}`,
		},
		{
			id: "novel-creation.chapter-plan", inputs: []string{"foundation_plan", "chapter_number", "memory_context", "user_message"},
			errors: map[string]string{"context_required": "章节策划上下文不能为空", "invalid_output": "章节策划 JSON 格式无效", "upstream_failed": "章节策划生成失败"},
			output: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"chapterNumber":"integer","title":"string","goal":"string","conflict":"string","turn":"string","hook":"string","targetWords":"integer"}}`,
		},
		{
			id: "novel-creation.chapter-draft", inputs: []string{"foundation_plan", "chapter_plan", "memory_context"},
			errors: map[string]string{"plan_required": "章节计划不能为空", "invalid_output": "章节正文 JSON 格式无效", "upstream_failed": "章节正文生成失败"},
			output: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"content":"string"}}`,
		},
		{
			id: "novel-creation.chapter-audit", inputs: []string{"foundation_plan", "chapter_plan", "chapter_content", "memory_context"},
			errors: map[string]string{"content_required": "章节正文不能为空", "invalid_output": "章节审核 JSON 格式无效", "upstream_failed": "章节审核失败"},
			output: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"score":"number 0..100","issues":[{"id":"string","severity":"low|medium|high|critical","description":"string","category":"string optional","evidence":"string optional"}],"passed":"boolean"}}`,
		},
		{
			id: "novel-creation.chapter-revision", inputs: []string{"chapter_content", "chapter_plan", "audit_issues", "memory_context"},
			errors: map[string]string{"issues_required": "审核问题不能为空", "invalid_output": "章节修订 JSON 格式无效", "upstream_failed": "章节修订失败"},
			output: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"strategy":"string","content":"string","resolvedIssueIDs":["string issue.id"]}}`,
		},
		{
			id: "novel-creation.truth-settlement", inputs: []string{"chapter_content", "previous_truth", "chapter_number"},
			errors: map[string]string{"content_required": "章节正文不能为空", "invalid_output": "事实结算 JSON 格式无效", "upstream_failed": "事实结算失败"},
			output: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"facts":[{"id":"string","statement":"string"}],"characterChanges":[{"character":"string","change":"string"}],"timeline":[{"event":"string","position":"string"}],"foreshadowing":[{"thread":"string","status":"open|resolved"}]}}`,
		},
		{
			id: "novel-creation.memory-index", inputs: []string{"chapter_bundle", "previous_memory", "truth_settlement"},
			errors: map[string]string{"bundle_required": "章节结果不能为空", "invalid_output": "记忆索引 JSON 格式无效", "upstream_failed": "记忆索引生成失败"},
			output: `{"format":"JSON object only","markdownCodeFence":"forbidden","additionalExplanation":"forbidden","structure":{"summary":"string","entities":[{"name":"string","type":"string","state":"string"}],"openThreads":[{"id":"string","thread":"string","nextExpectation":"string"}]}}`,
		},
	}

	got := All()
	if len(got) != len(want) {
		t.Fatalf("definitions = %d, want %d", len(got), len(want))
	}
	for index, expected := range want {
		definition := got[index]
		if definition.ID != expected.id || definition.Version != 1 {
			t.Fatalf("definition[%d] = %s@%d, want %s@1", index, definition.ID, definition.Version, expected.id)
		}
		if !reflect.DeepEqual(definition.RequiredInputs, expected.inputs) {
			t.Fatalf("%s inputs = %#v, want %#v", expected.id, definition.RequiredInputs, expected.inputs)
		}
		if !reflect.DeepEqual(definition.Errors, expected.errors) {
			t.Fatalf("%s errors = %#v, want %#v", expected.id, definition.Errors, expected.errors)
		}
		if definition.OutputContract != expected.output {
			t.Fatalf("%s output contract = %q, want %q", expected.id, definition.OutputContract, expected.output)
		}
	}
}

type definitionContract struct {
	id     string
	inputs []string
	errors map[string]string
	output string
}

// This catches callers accidentally changing future All() results by mutating
// the slices or maps returned by a previous call.
func TestAllReturnsDeepCopies(t *testing.T) {
	first := All()
	first[0].RequiredInputs[0] = "mutated_input"
	first[0].Errors["input_required"] = "mutated error"
	first[0].Errors["new_error"] = "unexpected"

	second := All()
	if got := second[0].RequiredInputs[0]; got != "project_brief" {
		t.Fatalf("required inputs leaked mutation: %q", got)
	}
	if got := second[0].Errors["input_required"]; got != "创作要求不能为空" {
		t.Fatalf("errors leaked mutation: %q", got)
	}
	if _, exists := second[0].Errors["new_error"]; exists {
		t.Fatal("errors leaked unexpected key")
	}
}

func TestChapterDraftRequiresJSONContentWithoutMarkdownFence(t *testing.T) {
	if !strings.Contains(ChapterDraft.SystemPrompt, "content") || !strings.Contains(ChapterDraft.SystemPrompt, "Markdown") {
		t.Fatalf("chapter draft prompt must require content JSON without a Markdown fence: %s", ChapterDraft.SystemPrompt)
	}
}
