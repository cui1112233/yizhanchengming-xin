package generation

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestWorkbenchExtractUsesSharedProviderWithoutCreatingRun(t *testing.T) {
	service, store, provider := serviceFixture()
	provider.responses = []string{"{\"characters\":[{\"id\":\"c1\",\"name\":\"林夏\",\"description\":\"记者\",\"protagonist\":true}],\"scenes\":[{\"id\":\"s1\",\"name\":\"医院走廊\",\"description\":\"夜间医院走廊\"}]}"}
	result, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID: 3, BookID: 11, Workbench: true, Action: WorkbenchActionExtract,
		SourceText: "林夏在医院走廊停下脚步。", DirectorMode: DirectorNormal,
	})
	if err != nil { t.Fatalf("extract: %v", err) }
	if result.Extraction == nil || len(result.Extraction.Characters) != 1 || len(result.Extraction.Scenes) != 1 {
		t.Fatalf("unexpected extraction: %#v", result.Extraction)
	}
	if !result.Extraction.Characters[0].Protagonist || result.Extraction.Characters[0].Name != "林夏" {
		t.Fatalf("character = %#v", result.Extraction.Characters[0])
	}
	if len(store.bookRuns) != 0 || len(store.stageRuns) != 0 {
		t.Fatalf("extract must not create generation history: runs=%d stages=%d", len(store.bookRuns), len(store.stageRuns))
	}
	if len(provider.calls) != 1 || provider.calls[0].Stage != StageExtract {
		t.Fatalf("provider calls = %#v", provider.calls)
	}
}

func TestWorkbenchGenerateComposesOpeningOutputEntitiesAndConstraints(t *testing.T) {
	service, _, provider := serviceFixture()
	provider.responses = []string{"SCRIPT", "HOOK", "{\"cards\":[{\"shot\":\"推门\"}]}"}
	result, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID: 3, BookID: 11, Workbench: true, Force: true,
		Action: WorkbenchActionGenerate, SourceText: "林夏推开病房门。",
		OpeningMode: OpeningHook, OutputMode: OutputCanvas, DirectorMode: DirectorNormal,
		Characters: []ScriptEntity{{Name: "林夏", Description: "记者", Protagonist: true, ReferenceImages: []string{"https://example.invalid/linxia.jpg"}}},
		Scenes: []ScriptEntity{{Name: "病房", Description: "夜间病房"}},
		Constraints: ScriptConstraints{VisualPrefix: "写实电影感", NegativePrompt: "不要水印"},
		RequestID: "workbench-1",
	})
	if err != nil { t.Fatalf("generate: %v", err) }
	if result.Run.Status != StatusCompleted { t.Fatalf("status = %s", result.Run.Status) }
	if len(provider.calls) != 3 { t.Fatalf("calls = %d", len(provider.calls)) }
	scriptInput := provider.calls[0].UserPrompt
	for _, needle := range []string{"【权威原文】", "林夏推开病房门", "【开头模式】爆款开头", "【输出模式】画布模式", "写实电影感", "example.invalid/linxia.jpg"} {
		if !strings.Contains(scriptInput, needle) { t.Fatalf("script input missing %q: %s", needle, scriptInput) }
	}
	if provider.calls[1].Stage != StageHook { t.Fatalf("second call stage = %s", provider.calls[1].Stage) }
	if !strings.Contains(provider.calls[2].UserPrompt, "\"openingMode\":\"hook\"") || !strings.Contains(provider.calls[2].UserPrompt, "\"outputMode\":\"canvas\"") {
		t.Fatalf("director payload = %s", provider.calls[2].UserPrompt)
	}
	final := result.Latest[StageFinalPrompt]
	if !strings.Contains(final.OutputText, "写实电影感") || !strings.Contains(final.OutputText, "不要水印") {
		t.Fatalf("final constraints missing: %s", final.OutputText)
	}
}

func TestWorkbenchContinuousSkipsHookAndPersistsHistoryVersions(t *testing.T) {
	service, _, provider := serviceFixture()
	provider.responses = []string{
		"SCRIPT-A", "{\"cards\":[]}",
		"SCRIPT-B", "{\"cards\":[]}",
	}
	base := RunBookRequest{
		BatchProjectID: 3, BookID: 11, Workbench: true, Force: true,
		Action: WorkbenchActionGenerate, SourceText: "第一版原文", OpeningMode: OpeningContinuous,
		OutputMode: OutputShotlist, DirectorMode: DirectorNormal,
	}
	if _, err := service.RunBook(context.Background(), base); err != nil { t.Fatalf("first: %v", err) }
	base.SourceText = "第二版原文"
	base.RequestID = "second"
	if _, err := service.RunBook(context.Background(), base); err != nil { t.Fatalf("second: %v", err) }
	summary, err := service.BookSummary(context.Background(), 3, 11)
	if err != nil { t.Fatal(err) }
	if len(summary.History) != 2 { t.Fatalf("history = %d", len(summary.History)) }
	if len(provider.calls) != 4 { t.Fatalf("provider calls = %d", len(provider.calls)) }
	for _, call := range provider.calls {
		if call.Stage == StageHook { t.Fatal("continuous opening must skip Hook provider call") }
	}
}

func TestRetryScriptReusesWorkbenchSnapshot(t *testing.T) {
	service, _, provider := serviceFixture()
	provider.errors[0] = errors.New("temporary")
	_, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID: 3, BookID: 11, Workbench: true, Force: true,
		Action: WorkbenchActionGenerate, SourceText: "编辑后的原文", OpeningMode: OpeningSegmented,
		OutputMode: OutputCanvas, DirectorMode: DirectorNormal, RequestID: "failed",
	})
	if err == nil { t.Fatal("expected first script failure") }
	delete(provider.errors, 0)
	provider.responses = []string{"", "SCRIPT-RETRY"}
	_, err = service.RetryStage(context.Background(), RetryStageRequest{BatchProjectID: 3, BookID: 11, Stage: StageScript, RequestID: "retry"})
	if err != nil { t.Fatalf("retry: %v", err) }
	if len(provider.calls) < 2 || !strings.Contains(provider.calls[1].UserPrompt, "编辑后的原文") || !strings.Contains(provider.calls[1].UserPrompt, "【开头模式】分段开头") {
		t.Fatalf("retry did not preserve workbench input: %#v", provider.calls)
	}
}
