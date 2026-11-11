package screenplay

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

const sampleScreenplay = `# 第1集：八位数和我，总有一个要退圈

## 1-1：日 内 蓝雨传媒28楼·经纪人办公室
人物：顾苒苒 吴姐
（顾苒苒：女/23岁/长相甜美·笑起来两个梨涡/倔强毒舌）
（吴姐：女/40+/金牌经纪人/精明、不耐烦）

△ 落地窗外是城市天际线，玻璃冷光。【拍摄注释：高层办公室冷色调。】
△ 特写：合同抬头——"解约协议"。
吴姐（居高临下、不耐）：退圈，是你最后的体面。
顾苒苒（冷笑、呛声）：凭什么？
顾苒苒（OS、翻白眼）：姐要爬，也挑个圈内小鲜肉。
△【集末卡点】她抓起手机一键下单。

---

## 1-2：夜 内 顾苒苒公寓
人物：顾苒苒

△ 顾苒苒瘫在沙发上充电。
顾苒苒（OS）：回家种田。

---

# 第2集：天亮之前

## 2-1：清晨 外 小河村·祖宅门口
人物：顾苒苒 王婶

△ 晨雾未散。
王婶（惊讶）：苒苒回来啦？
`

func TestParseScreenplayMarkdown(t *testing.T) {
	parsed := ParseScreenplayMarkdown(sampleScreenplay)
	if len(parsed.Episodes) != 2 || parsed.Episodes[0] != 1 || parsed.Episodes[1] != 2 {
		t.Fatalf("episodes = %v, want [1 2]", parsed.Episodes)
	}
	if len(parsed.Shots) != 3 {
		t.Fatalf("shots = %d, want 3", len(parsed.Shots))
	}
	if parsed.Shots[0].SceneName != "蓝雨传媒28楼·经纪人办公室" {
		t.Fatalf("scene = %q", parsed.Shots[0].SceneName)
	}
	if len(parsed.Scenes) != 3 {
		t.Fatalf("scenes = %d, want 3", len(parsed.Scenes))
	}
	if !containsString(parsed.Shots[0].CharacterNames, "顾苒苒") {
		t.Fatalf("characters = %v", parsed.Shots[0].CharacterNames)
	}
	if len(parsed.Characters) < 3 {
		t.Fatalf("characters = %d, want >=3", len(parsed.Characters))
	}
	if parsed.Characters[0].Description == "" {
		t.Fatalf("人物卡描述未解析: %+v", parsed.Characters[0])
	}
	if !strings.Contains(parsed.Shots[0].ScriptText, "解约协议") {
		t.Fatalf("script text 丢失: %q", parsed.Shots[0].ScriptText)
	}
}

type fakeGenerator struct {
	mu        sync.Mutex
	responses []string
	err       error
	requests  []TextRequest
}

func (f *fakeGenerator) Generate(_ context.Context, req TextRequest) (TextResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, req)
	if f.err != nil {
		return TextResponse{}, f.err
	}
	index := len(f.requests) - 1
	if len(f.responses) == 0 {
		return TextResponse{}, errors.New("fake generator 未配置返回")
	}
	if index >= len(f.responses) {
		return TextResponse{Content: f.responses[len(f.responses)-1]}, nil
	}
	return TextResponse{Content: f.responses[index]}, nil
}

func (f *fakeGenerator) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

func newTestService(t *testing.T) (*ScreenplayService, *MemoryStore, *fakeGenerator) {
	t.Helper()
	store := NewMemoryStore()
	generator := &fakeGenerator{responses: []string{sampleScreenplay}}
	service := &ScreenplayService{Store: store, Generator: generator}
	if _, err := store.CreateProject(context.Background(), Project{
		ID: "p1", OwnerUsername: "alice", Title: "测试", PermissionMode: PermissionBypass,
	}); err != nil {
		t.Fatal(err)
	}
	return service, store, generator
}

func TestA1EmptyContentRejectedWithoutCallingModel(t *testing.T) {
	service, _, generator := newTestService(t)
	_, err := service.GenerateA1(context.Background(), A1Request{Owner: "alice", ProjectID: "p1", Content: "  "})
	if !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("err = %v, want ErrEmptyContent", err)
	}
	if len(generator.requests) != 0 {
		t.Fatal("空输入不应调用模型")
	}
}

func TestA1AutoPermissionReturnsPlanOnly(t *testing.T) {
	service, store, generator := newTestService(t)
	if _, err := store.CreateProject(context.Background(), Project{
		ID: "p2", OwnerUsername: "alice", PermissionMode: PermissionAuto,
	}); err != nil {
		t.Fatal(err)
	}
	result, err := service.GenerateA1(context.Background(), A1Request{Owner: "alice", ProjectID: "p2", Content: "第1章 正文"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.RequiresConfirmation || len(result.PendingActions) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if len(generator.requests) != 0 {
		t.Fatal("auto 权限未确认时不应调用模型")
	}
	if _, err := store.ActiveDocument(context.Background(), "alice", "p2"); !errors.Is(err, ErrNoActiveScript) {
		t.Fatal("auto 权限未确认时不应写库")
	}
}

func TestA1ConfirmedPersistsAndVersions(t *testing.T) {
	service, store, generator := newTestService(t)
	first, err := service.GenerateA1(context.Background(), A1Request{Owner: "alice", ProjectID: "p1", Content: "第1章"})
	if err != nil {
		t.Fatal(err)
	}
	if first.Document.Version != 1 || first.Document.Status != "active" {
		t.Fatalf("document = %+v", first.Document)
	}
	if first.EpisodeCount != 2 || first.ShotCount != 3 {
		t.Fatalf("episode/shot = %d/%d", first.EpisodeCount, first.ShotCount)
	}
	second, err := service.GenerateA1(context.Background(), A1Request{Owner: "alice", ProjectID: "p1", Content: "第1章 重跑"})
	if err != nil {
		t.Fatal(err)
	}
	if second.Document.Version != 2 {
		t.Fatalf("version = %d, want 2", second.Document.Version)
	}
	documents, err := store.ListDocuments(context.Background(), "alice", "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 2 {
		t.Fatalf("documents = %d, want 2", len(documents))
	}
	if documents[0].Status != "active" || documents[1].Status != "superseded" {
		t.Fatalf("status = %s/%s", documents[0].Status, documents[1].Status)
	}
	if len(generator.requests) != 2 {
		t.Fatalf("calls = %d, want 2", len(generator.requests))
	}
}

func TestA1SegmentsConcatenatedInOrder(t *testing.T) {
	service, _, generator := newTestService(t)
	if _, err := service.GenerateA1(context.Background(), A1Request{
		Owner: "alice", ProjectID: "p1",
		Segments: []string{"第一段", "第二段"},
	}); err != nil {
		t.Fatal(err)
	}
	if len(generator.requests) != 2 {
		t.Fatalf("calls = %d, want 2", len(generator.requests))
	}
	if generator.requests[0].UserPrompt != "第一段" || generator.requests[1].UserPrompt != "第二段" {
		t.Fatalf("分段顺序错误: %q / %q", generator.requests[0].UserPrompt, generator.requests[1].UserPrompt)
	}
}

func TestA1EmptySegmentRejected(t *testing.T) {
	service, _, generator := newTestService(t)
	_, err := service.GenerateA1(context.Background(), A1Request{
		Owner: "alice", ProjectID: "p1", Segments: []string{"有内容", "   "},
	})
	if !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("err = %v, want ErrEmptyContent", err)
	}
	if len(generator.requests) != 0 {
		t.Fatal("空分段不应调用模型")
	}
}

func TestA1ProviderEmptyContentMapsToSpecString(t *testing.T) {
	service, _, generator := newTestService(t)
	generator.responses = []string{"   "}
	_, err := service.GenerateA1(context.Background(), A1Request{Owner: "alice", ProjectID: "p1", Content: "第1章"})
	if !errors.Is(err, ErrAIEmptyContent) {
		t.Fatalf("err = %v, want ErrAIEmptyContent", err)
	}
	if err.Error() != "AI 未返回内容" {
		t.Fatalf("err string = %q", err.Error())
	}
}

func TestA1UnparsableOutputRejected(t *testing.T) {
	service, _, generator := newTestService(t)
	generator.responses = []string{"这是一段没有格式的文本"}
	_, err := service.GenerateA1(context.Background(), A1Request{Owner: "alice", ProjectID: "p1", Content: "第1章"})
	if !errors.Is(err, ErrArtifactInvalid) {
		t.Fatalf("err = %v, want ErrArtifactInvalid", err)
	}
}

func TestA1TooLongRequiresSegments(t *testing.T) {
	service, _, _ := newTestService(t)
	content := strings.Repeat("字", MaxSingleCallRunes+1)
	_, err := service.GenerateA1(context.Background(), A1Request{Owner: "alice", ProjectID: "p1", Content: content})
	if err == nil || !strings.Contains(err.Error(), "60 万字") {
		t.Fatalf("err = %v", err)
	}
}

func TestBEmptyNovelRejected(t *testing.T) {
	service, _, generator := newTestService(t)
	_, err := service.GenerateB(context.Background(), BRequest{Owner: "alice", ProjectID: "p1", Novel: " "})
	if !errors.Is(err, ErrEmptyNovel) {
		t.Fatalf("err = %v, want ErrEmptyNovel", err)
	}
	if len(generator.requests) != 0 {
		t.Fatal("空 novel 不应调用模型")
	}
}

func TestBIncludesOptionsInUserPrompt(t *testing.T) {
	service, _, generator := newTestService(t)
	if _, err := service.GenerateB(context.Background(), BRequest{
		Owner: "alice", ProjectID: "p1", Novel: "第1章 正文",
		ChapterRange: "1-3", EpisodeTarget: "90秒", ExtraNote: "强化钩子",
	}); err != nil {
		t.Fatal(err)
	}
	prompt := generator.requests[0].UserPrompt
	for _, want := range []string{"本次改编章节：1-3", "单集目标时长：90秒", "特殊要求：强化钩子", "# 待改编的小说原文"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("user prompt 缺少 %q", want)
		}
	}
}
