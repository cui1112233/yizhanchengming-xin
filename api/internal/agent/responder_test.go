package agent

import (
	"context"
	"encoding/json"
	"testing"
)

func TestResponderOpensNovelFetch(t *testing.T) {
	response, err := NewResponder().Respond(context.Background(), "owner", "thread", "帮我打开小说获取")
	if err != nil { t.Fatal(err) }
	if response.Tool == nil || response.Tool.Name != ToolNavigate {
		t.Fatalf("expected navigate tool, got %#v", response.Tool)
	}
	var args NavigateArgs
	if err := json.Unmarshal(response.Tool.Arguments, &args); err != nil { t.Fatal(err) }
	if args.Path != "/novel-fetch" { t.Fatalf("path=%q", args.Path) }
	if response.Content == "" { t.Fatal("assistant content must explain the action") }
}

func TestResponderOpensBatchFactory(t *testing.T) {
	response, err := NewResponder().Respond(context.Background(), "owner", "thread", "打开批量工厂")
	if err != nil { t.Fatal(err) }
	var args NavigateArgs
	if response.Tool == nil { t.Fatal("expected tool") }
	if err := json.Unmarshal(response.Tool.Arguments, &args); err != nil { t.Fatal(err) }
	if args.Path != "/batch-factory" { t.Fatalf("path=%q", args.Path) }
}

func TestResponderExplainsImageModelSettingAndLinksConfig(t *testing.T) {
	response, err := NewResponder().Respond(context.Background(), "owner", "thread", "图片模型在哪里设置？")
	if err != nil { t.Fatal(err) }
	if response.Tool == nil { t.Fatal("expected config navigation tool") }
	var args NavigateArgs
	if err := json.Unmarshal(response.Tool.Arguments, &args); err != nil { t.Fatal(err) }
	if args.Path != "/api-config" || args.Section != "image-models" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestResponderNeverNavigatesToArbitraryExternalURL(t *testing.T) {
	response, err := NewResponder().Respond(context.Background(), "owner", "thread", "帮我打开 https://evil.example")
	if err != nil { t.Fatal(err) }
	if response.Tool != nil {
		t.Fatalf("unexpected external navigation tool: %#v", response.Tool)
	}
}

func TestResponderCanCreateDurableTaskIntent(t *testing.T) {
	response, err := NewResponder().Respond(context.Background(), "owner", "thread", "创建任务：检查今天的小说获取")
	if err != nil { t.Fatal(err) }
	if response.Task == nil { t.Fatal("expected task proposal") }
	if response.Task.Title != "检查今天的小说获取" || response.Task.Status != TaskInProgress {
		t.Fatalf("unexpected task: %#v", response.Task)
	}
}

func TestResponderUnknownRequestIsTransparent(t *testing.T) {
	response, err := NewResponder().Respond(context.Background(), "owner", "thread", "把人物头发变短然后直接生成四张图")
	if err != nil { t.Fatal(err) }
	if response.Tool != nil { t.Fatalf("must not fake image tool: %#v", response.Tool) }
	if response.Content == "" { t.Fatal("expected explanatory response") }
}
