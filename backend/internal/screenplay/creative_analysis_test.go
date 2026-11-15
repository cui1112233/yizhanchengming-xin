package screenplay

import (
	"context"
	"os"
	"qiantie/backend/internal/prompt"
	"strings"
	"testing"
)

func realProfessionalAnalysisReply(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("testdata/professional144-analysis-reply.txt")
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestProfessional144RealAnalysisReplyPersistsLosslessly(t *testing.T) {
	raw := realProfessionalAnalysisReply(t)
	for _, mode := range []string{"b", "a1"} {
		t.Run(mode, func(t *testing.T) {
			svc, store, g := professionalService(t, raw)
			var out ScreenplayResult
			var err error
			if mode == "b" {
				out, err = svc.GenerateB(context.Background(), BRequest{Owner: "o", ProjectID: "p", Novel: "测试原文"})
			} else {
				out, err = svc.GenerateA1(context.Background(), A1Request{Owner: "o", ProjectID: "p", Content: "测试原文"})
			}
			if err != nil {
				t.Fatal(err)
			}
			saved, err := store.ActiveDocument(context.Background(), "o", "p")
			if err != nil {
				t.Fatal(err)
			}
			if saved.RawResult != raw || !strings.HasPrefix(saved.Content, "# 第1集") || !strings.Contains(saved.Content, "今天，画一个新的开始。") || strings.Contains(saved.Content, "信息清单核对") || strings.Contains(saved.Content, "转换自检") || len(out.Shots) != 1 || len(g.requests) != 1 {
				t.Fatal("analysis/body/raw separation or call count incorrect")
			}
		})
	}
	chapter := Chapter{}
	if err := ApplyChapterGeneration(&chapter, raw); err != nil || chapter.RawResult != raw || !strings.Contains(chapter.ScriptText, "今天，画一个新的开始。") {
		t.Fatalf("chapter: %v", err)
	}
}

func TestProfessionalAnalysisEnvelopeDoesNotRelaxInvalidReplies(t *testing.T) {
	raw := realProfessionalAnalysisReply(t)
	index := strings.Index(raw, "{")
	body := raw[index:]
	prefix := raw[:index]
	for name, content := range map[string]string{
		"unknown introduction":     "以下是结果\n" + body,
		"unknown analysis section": "场景划分结果：\n第1-1场：书房\n其他说明：\n- 不属于协议\n" + body,
		"analysis only":            prefix,
		"second object":            raw + professionalJSONForEpisode(2),
		"object in preface":        "场景划分结果：\n{\"example\":true}\n" + body,
		"incomplete object":        prefix + `{"episode_index":1,"scenes":[`,
		"empty scene":              prefix + `{"scenes":[]}`,
		"unknown tail":             raw + "\n未约定的尾部说明",
	} {
		t.Run(name, func(t *testing.T) {
			c := Chapter{RawResult: "旧原回包", ScriptText: "旧成功剧本"}
			if ApplyChapterGeneration(&c, content) == nil || c.RawResult != "旧原回包" || c.ScriptText != "旧成功剧本" {
				t.Fatal("invalid reply accepted or old draft overwritten")
			}
		})
	}
	fenced := prefix + "```json\n" + professionalJSONForEpisode(1) + "\n```" + professionalSelfCheckReport
	if _, err := selectedScreenplayOutput(fenced, prompt.CreativeScript144); err != nil {
		t.Fatal(err)
	}
}
