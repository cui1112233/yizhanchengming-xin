package screenplay

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestSplitIntoBatchesByChapters(t *testing.T) {
	var builder strings.Builder
	for i := 1; i <= 12; i++ {
		builder.WriteString(fmt.Sprintf("第%d章 标题\n正文内容\n", i))
	}
	chunks := SplitIntoBatches(builder.String(), 5)
	if len(chunks) != 3 {
		t.Fatalf("chunks = %d, want 3", len(chunks))
	}
	if !strings.Contains(chunks[0], "第1章") || !strings.Contains(chunks[0], "第5章") {
		t.Fatalf("第一批内容错误: %q", chunks[0])
	}
	if strings.Contains(chunks[0], "第6章") {
		t.Fatal("第一批不应包含第 6 章")
	}
	if !strings.Contains(chunks[2], "第12章") {
		t.Fatalf("最后一批内容错误: %q", chunks[2])
	}
}

func TestSplitIntoBatchesByEpisodesFallback(t *testing.T) {
	content := "# 第1集：A\n正文\n# 第2集：B\n正文\n# 第3集：C\n正文\n"
	chunks := SplitIntoBatches(content, 2)
	if len(chunks) != 2 {
		t.Fatalf("chunks = %d, want 2", len(chunks))
	}
	if !strings.Contains(chunks[0], "第1集") || !strings.Contains(chunks[0], "第2集") {
		t.Fatalf("chunk0 = %q", chunks[0])
	}
}

const charactersJSON = `{
  "characters": [
    {"name": "顾苒苒", "颜值定位": "甜美毒舌女主", "出现集数": "第01-03集", "identity": "主角，倔强", "prompt": "甜美毒舌女主的完整提示词", "voice": "清亮少女音【标签：甜妹音】", "tag": "主角"},
    {"name": "吴姐", "颜值定位": "精明经纪人", "出现集数": "第01集", "identity": "配角", "prompt": "精明经纪人的完整提示词", "voice": "低沉御姐音【标签：御姐音】", "tag": ""}
  ]
}`

func newAssetService(t *testing.T, mode PermissionMode) (*AssetService, *MemoryStore, *fakeGenerator) {
	t.Helper()
	store := NewMemoryStore()
	if _, err := store.CreateProject(context.Background(), Project{ID: "p1", OwnerUsername: "alice", PermissionMode: mode}); err != nil {
		t.Fatal(err)
	}
	generator := &fakeGenerator{responses: []string{charactersJSON}}
	return &AssetService{Store: store, Generator: generator}, store, generator
}

func TestExtractCharactersFourDegreesUseExpectedPrompts(t *testing.T) {
	cases := map[AssetDegree]string{
		DegreeStandard:   "asset_character_standard_v5",
		DegreeMoreDiff:   "asset_character_lite",
		DegreeMaster:     "asset_character_v6_3d",
		DegreeMasterReal: "asset_character_real_v7",
	}
	for degree, promptID := range cases {
		service, store, generator := newAssetService(t, PermissionBypass)
		result, err := service.Extract(context.Background(), AssetRequest{
			Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: degree,
			Content: "第1章 正文",
		})
		if err != nil {
			t.Fatalf("%s: %v", degree, err)
		}
		if len(result.Assets) != 2 {
			t.Fatalf("%s assets = %d", degree, len(result.Assets))
		}
		body, _ := service.prompts().Body(promptID)
		if generator.requests[0].SystemPrompt != body {
			t.Fatalf("%s 使用的提示词不是 %s", degree, promptID)
		}
		assets, _ := store.ListAssets(context.Background(), "alice", "p1", AssetCharacter)
		if assets[0].Voice == "" || assets[0].Episodes != "第01-03集" || assets[0].Tag != "主角" {
			t.Fatalf("%s 音色/集数/标签丢失: %+v", degree, assets[0])
		}
	}
}

func TestExtractScenesAndProps(t *testing.T) {
	scenesJSON := `{"scenes":[{"name":"经纪人办公室","description":"高层冷光","prompt":"无人广角","tag":"重要场景"}]}`
	propsJSON := `{"props":[{"name":"解约协议","description":"关键文件","prompt":"纯背景物体","tag":""}]}`
	for _, tc := range []struct {
		kind  AssetType
		raw   string
		avoid string
	}{
		{AssetScene, scenesJSON, "经纪人办公室"},
		{AssetProp, propsJSON, "解约协议"},
	} {
		service, _, generator := newAssetService(t, PermissionBypass)
		generator.responses = []string{tc.raw}
		result, err := service.Extract(context.Background(), AssetRequest{
			Owner: "alice", ProjectID: "p1", Type: tc.kind, Degree: DegreeMaster, Content: "第1章",
		})
		if err != nil {
			t.Fatalf("%s: %v", tc.kind, err)
		}
		if len(result.Assets) != 1 || result.Assets[0].Name != tc.avoid {
			t.Fatalf("%s assets = %+v", tc.kind, result.Assets)
		}
	}
}

func TestExtractMergesSameNameAcrossChunks(t *testing.T) {
	service, store, generator := newAssetService(t, PermissionBypass)
	generator.responses = []string{
		`{"characters":[{"name":"顾苒苒","identity":"主角","prompt":"短提示词","voice":""}]}`,
		`{"characters":[{"name":"顾苒苒","identity":"","prompt":"更长更完整的提示词","voice":"清亮少女音"}]}`,
	}
	chapters := []string{"批次一内容", "批次二内容"}
	result, err := service.Extract(context.Background(), AssetRequest{
		Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: DegreeStandard, Chapters: chapters,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Assets) != 1 {
		t.Fatalf("assets = %d, want 1（同名合并）", len(result.Assets))
	}
	assets, _ := store.ListAssets(context.Background(), "alice", "p1", AssetCharacter)
	if assets[0].Prompt != "更长更完整的提示词" || assets[0].Voice != "清亮少女音" {
		t.Fatalf("合并结果错误: %+v", assets[0])
	}
	if result.ChunkCount != 2 {
		t.Fatalf("chunks = %d", result.ChunkCount)
	}
}

func TestExtractEmptyResultProducesWarning(t *testing.T) {
	service, _, generator := newAssetService(t, PermissionBypass)
	generator.responses = []string{`{"characters":[]}`}
	result, err := service.Extract(context.Background(), AssetRequest{
		Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: DegreeStandard, Content: "第1章",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) == 0 || !strings.Contains(result.Warnings[0], "未识别到人物") {
		t.Fatalf("warnings = %v", result.Warnings)
	}
}

func TestExtractInvalidJSONFailsChunkButKeepsOthers(t *testing.T) {
	service, _, generator := newAssetService(t, PermissionBypass)
	generator.responses = []string{
		`{"characters":[{"name":"顾苒苒","prompt":"正常"}]}`,
		`这不是 JSON`,
	}
	result, err := service.Extract(context.Background(), AssetRequest{
		Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: DegreeStandard,
		Chapters: []string{"一", "二"},
	})
	if !errors.Is(err, ErrPartialFailure) {
		t.Fatalf("部分批次失败必须返回 ErrPartialFailure，err = %v", err)
	}
	if result.FailedChunks != 1 {
		t.Fatalf("failed = %d, want 1", result.FailedChunks)
	}
	if len(result.Assets) != 1 {
		t.Fatalf("assets = %d, want 1", len(result.Assets))
	}
	if !result.PartialFailure || len(result.FailedBatches) != 1 || result.FailedBatches[0].BatchIndex <= 0 {
		t.Fatalf("失败批次信息不完整: %+v", result.FailedBatches)
	}
}

func TestExtractAllChunksFailedReturnsError(t *testing.T) {
	service, _, generator := newAssetService(t, PermissionBypass)
	generator.err = errors.New("boom")
	_, err := service.Extract(context.Background(), AssetRequest{
		Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: DegreeStandard, Content: "第1章",
	})
	if err == nil {
		t.Fatal("全部批次失败应返回错误")
	}
}

func TestExtractNoSourceAndNoDocument(t *testing.T) {
	service, _, _ := newAssetService(t, PermissionBypass)
	_, err := service.Extract(context.Background(), AssetRequest{
		Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: DegreeStandard,
	})
	if !errors.Is(err, ErrNoActiveScript) {
		t.Fatalf("err = %v, want ErrNoActiveScript", err)
	}
}

func TestExtractAutoPermissionPlansOnly(t *testing.T) {
	service, _, generator := newAssetService(t, PermissionAuto)
	result, err := service.Extract(context.Background(), AssetRequest{
		Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: DegreeStandard, Content: "第1章",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.RequiresConfirmation || len(result.PendingActions) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if generator.callCount() != 0 {
		t.Fatal("auto 未确认不应调用模型")
	}
}

func TestExtractRejectsUnknownTypeAndDegree(t *testing.T) {
	service, _, _ := newAssetService(t, PermissionBypass)
	if _, err := service.Extract(context.Background(), AssetRequest{
		Owner: "alice", ProjectID: "p1", Type: "unknown", Degree: DegreeStandard, Content: "x",
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("type err = %v", err)
	}
	if _, err := service.Extract(context.Background(), AssetRequest{
		Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: "unknown", Content: "x",
	}); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("degree err = %v", err)
	}
}

func TestExtractUsesActiveDocumentWhenContentEmpty(t *testing.T) {
	service, store, generator := newAssetService(t, PermissionBypass)
	doc, err := store.SaveScreenplay(context.Background(), Document{
		ID: "doc1", ProjectID: "p1", OwnerUsername: "alice", Source: "b", Content: "# 第1集：A\n第1章正文",
	}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != 1 {
		t.Fatalf("version = %d", doc.Version)
	}
	if _, err := service.Extract(context.Background(), AssetRequest{
		Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Degree: DegreeStandard,
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(generator.requests[0].UserPrompt, "第1章正文") {
		t.Fatalf("未使用激活剧本: %q", generator.requests[0].UserPrompt)
	}
}
