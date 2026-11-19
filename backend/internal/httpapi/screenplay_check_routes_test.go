package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
)

func TestScreenplayRoutesStoryboardCheck(t *testing.T) {
	now := time.Date(2026, 11, 19, 0, 30, 0, 0, time.UTC)
	store := screenplay.NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, screenplay.Project{ID: "check-project", OwnerUsername: "alice", Title: "分镜检查"}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveStoryboard(ctx, screenplay.Storyboard{
		ProjectID: "check-project", OwnerUsername: "alice", EpisodeNo: 1, Engine: screenplay.EngineXiaoluoV7,
		Shots: []screenplay.StoryboardShot{
			{ID: "sc-1", Seq: 1, VideoDesc: "开场"},
			{ID: "sc-2", Seq: 2, VideoDesc: "推进"},
			{ID: "sc-3", Seq: 4, VideoDesc: "对峙", Duration: 6},
		},
	}); err != nil {
		t.Fatal(err)
	}
	provider := &batchTranslateProvider{result: `{"characters":"林川","scene":"街道","video_desc":"衔接镜头","image_prompt":"夜街","line":"跟上","duration_sec":5}`}
	resolver := func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, resolver, func() time.Time { return now })
	base := "/api/comic-screenplays/projects/check-project/storyboards/xiaoluo_v7_15s/1/check"

	checked := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base, map[string]any{})
	if checked.Code != http.StatusOK {
		t.Fatalf("check=%d %s", checked.Code, checked.Body.String())
	}
	var payload struct {
		Report screenplay.StoryboardCheckReport `json:"report"`
	}
	if err := json.Unmarshal(checked.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Report.Total != 3 || payload.Report.MaxSeq != 4 || payload.Report.ExpectedTotal != 4 {
		t.Fatalf("报告统计错误: %+v", payload.Report)
	}
	if strings.Join(intToStrings(payload.Report.Missing), ",") != "3" || !strings.Contains(checked.Body.String(), `"passed":false`) {
		t.Fatalf("缺失应为 3: %+v %s", payload.Report, checked.Body.String())
	}
	if strings.Contains(checked.Body.String(), `"storyboard"`) {
		t.Fatalf("check 只读，不应返回整集分镜: %s", checked.Body.String())
	}

	filled := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/fill", map[string]any{
		"seq": 3, "model": "text-main",
	})
	if filled.Code != http.StatusOK {
		t.Fatalf("fill=%d %s", filled.Code, filled.Body.String())
	}
	if provider.request.SystemPrompt == "" || !strings.Contains(provider.request.UserPrompt, "分镜序号 3 缺失") {
		t.Fatalf("模型输入不完整: system=%q user=%q", provider.request.SystemPrompt, provider.request.UserPrompt)
	}
	stored, err := store.GetStoryboard(ctx, "alice", "check-project", 1, screenplay.EngineXiaoluoV7)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Shots) != 4 {
		t.Fatalf("补全后应有 4 条分镜: %+v", stored.Shots)
	}
	for index, want := range []int{1, 2, 3, 4} {
		if stored.Shots[index].Seq != want {
			t.Fatalf("顺序错误 index=%d seq=%d want=%d", index, stored.Shots[index].Seq, want)
		}
	}
	if stored.Shots[2].VideoDesc != "衔接镜头" || stored.Shots[2].ID == "sc-1" {
		t.Fatalf("补全分镜未正确插入: %+v", stored.Shots[2])
	}
	report, err := screenplay.CheckStoryboardShots(stored.Shots)
	if err != nil || !report.Passed {
		t.Fatalf("补全后应通过: %+v err=%v", report, err)
	}

	existing := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/fill", map[string]any{"seq": 4, "model": "text-main"})
	if existing.Code == http.StatusOK || !strings.Contains(existing.Body.String(), "不是缺失分镜") {
		t.Fatalf("已存在序号应拒绝: %d %s", existing.Code, existing.Body.String())
	}
	noSeq := signedJSONRequest(t, mux, now, "alice", http.MethodPost, base+"/fill", map[string]any{"model": "text-main"})
	if noSeq.Code == http.StatusOK {
		t.Fatalf("缺少序号应拒绝: %s", noSeq.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodPost, base, map[string]any{})
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign check=%d %s", foreign.Code, foreign.Body.String())
	}
}

func intToStrings(values []int) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, jsonNumber(value))
	}
	return result
}

func jsonNumber(value int) string {
	data, _ := json.Marshal(value)
	return string(data)
}
