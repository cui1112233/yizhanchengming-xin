package screenplay

import (
	"context"
	"errors"
	"fmt"
	"qiantie/backend/internal/prompt"
	"strings"
	"testing"
)

const professionalSelfCheckReport = "\n【转换自检】\n\n原文对话总数：1句 → dialogue元素数：1 → 遗漏：0\naction元素总数：1（atmosphere: 0 / camera: 0 / action: 1）\n场景总数：1（含回切场景：0个）\nelement_id 是否连续无重复：✓\n每集结构：有开场 ✓ / 有收场 ✓\n"

func professionalJSONForEpisode(episode int) string {
	return fmt.Sprintf(`{"episode_index":%d,"title":"第%d章故事","scenes":[{"scene_number":"%d-1","location":"家","time":"日","indoor_outdoor":"内","characters":["甲"],"script_content":[{"element_type":"action","description":"门开了"},{"element_type":"dialogue","character":"甲","content":"第%d集完整尾句。","is_voiceover":false}]}]}`, episode, episode, episode, episode)
}
func TestProfessional144AcceptsDocumentedSelfCheckOnly(t *testing.T) {
	base := professionalJSONForEpisode(1)
	for _, suffix := range []string{"", professionalSelfCheckReport} {
		c := Chapter{}
		raw := " \n" + base + suffix + " \n"
		if err := ApplyChapterGeneration(&c, raw); err != nil {
			t.Fatal(err)
		}
		if c.RawResult != raw || !strings.Contains(c.ScriptText, "第1集完整尾句。") || strings.Contains(c.ScriptText, "转换自检") {
			t.Fatalf("raw/editable result wrong: %+v", c)
		}
	}
}
func TestProfessional144RejectsUnknownTextWithoutOverwriting(t *testing.T) {
	base := professionalJSONForEpisode(1)
	for _, raw := range []string{"未知前部\n" + base, base + "\n未知尾部", base + professionalSelfCheckReport + "未知尾部", base + professionalSelfCheckReport + professionalJSONForEpisode(2), base + "\n【转换自检】\n原文对话总数：1\n{\"garbage\":true}", `{"scenes":[}`, `{"scenes":[]}`} {
		c := Chapter{ScriptText: "上次成功", RawResult: "上次原文"}
		if ApplyChapterGeneration(&c, raw) == nil || c.ScriptText != "上次成功" || c.RawResult != "上次原文" {
			t.Fatalf("invalid output accepted or overwrote chapter: %q", raw)
		}
	}
}
func professionalService(t *testing.T, responses ...string) (*ScreenplayService, *MemoryStore, *fakeGenerator) {
	t.Helper()
	s := NewMemoryStore()
	_, err := s.CreateProject(context.Background(), Project{ID: "p", OwnerUsername: "o", PromptID: prompt.CreativeScript144, PermissionMode: PermissionBypass})
	if err != nil {
		t.Fatal(err)
	}
	g := &fakeGenerator{responses: responses}
	return &ScreenplayService{Store: s, Generator: g}, s, g
}
func TestProfessional144BAndChapterDoNotRequireJSONOnly(t *testing.T) {
	raw := professionalJSONForEpisode(1) + professionalSelfCheckReport
	svc, _, g := professionalService(t, raw)
	out, err := svc.GenerateB(context.Background(), BRequest{Owner: "o", ProjectID: "p", Novel: "原文"})
	if err != nil {
		t.Fatal(err)
	}
	if g.requests[0].JSONResponse || out.Document.RawResult != raw || strings.Contains(out.Document.Content, "转换自检") {
		t.Fatal("B constrained JSON-only or lost raw/self-check separation")
	}
	req, err := BuildChapterTextRequest(Project{PromptID: prompt.CreativeScript144}, Chapter{SourceText: "原文"}, "", "")
	if err != nil || req.JSONResponse {
		t.Fatalf("chapter JSON-only=%v err=%v", req.JSONResponse, err)
	}
}
func TestProfessional144A1ParsesEachSegmentBeforeJoining(t *testing.T) {
	first := professionalJSONForEpisode(1) + professionalSelfCheckReport
	second := professionalJSONForEpisode(2) + professionalSelfCheckReport
	svc, store, g := professionalService(t, first, second)
	out, err := svc.GenerateA1(context.Background(), A1Request{Owner: "o", ProjectID: "p", Segments: []string{"第一章原文", "第二章原文"}})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := store.ActiveDocument(context.Background(), "o", "p")
	if err != nil {
		t.Fatal(err)
	}
	if saved.RawResult != first+"\n\n"+second || len(out.Shots) != 2 || out.EpisodeCount != 2 || !strings.Contains(saved.Content, "第2集完整尾句。") || strings.Contains(saved.Content, "转换自检") {
		t.Fatalf("segments lost/separation wrong: %+v", out)
	}
	for _, req := range g.requests {
		if req.JSONResponse {
			t.Fatal("A1 JSON-only forced")
		}
	}
}
func TestProfessional144FailedSecondSegmentPreservesActiveDocument(t *testing.T) {
	svc, store, _ := professionalService(t, professionalJSONForEpisode(1)+professionalSelfCheckReport, professionalJSONForEpisode(2)+"\n未知尾部")
	old, err := (&ScreenplayService{Store: store}).persist(context.Background(), "o", "p", "b", sampleScreenplay, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.GenerateA1(context.Background(), A1Request{Owner: "o", ProjectID: "p", Segments: []string{"一", "二"}})
	if err == nil {
		t.Fatal("invalid second result accepted")
	}
	saved, err := store.ActiveDocument(context.Background(), "o", "p")
	if err != nil || saved.ID != old.Document.ID {
		t.Fatal("failed generation changed active document")
	}
}
func TestProfessional144CanceledSelfCheckResponseDoesNotSave(t *testing.T) {
	for _, mode := range []string{"a1", "b"} {
		t.Run(mode, func(t *testing.T) {
			svc, store, _ := professionalService(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			svc.Generator = lateCanceledGenerator{cancel: cancel, content: professionalJSONForEpisode(1) + professionalSelfCheckReport}
			var err error
			if mode == "a1" {
				_, err = svc.GenerateA1(ctx, A1Request{Owner: "o", ProjectID: "p", Content: "原文"})
			} else {
				_, err = svc.GenerateB(ctx, BRequest{Owner: "o", ProjectID: "p", Novel: "原文"})
			}
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("error=%v", err)
			}
			docs, _ := store.ListDocuments(context.Background(), "o", "p")
			if len(docs) != 0 {
				t.Fatal("canceled response saved")
			}
		})
	}
}
func TestProfessional144A1RejectsExtraEpisodeInsideOneResponse(t *testing.T) {
	svc, store, g := professionalService(t, professionalJSONForEpisode(1)+professionalSelfCheckReport+professionalJSONForEpisode(2))
	_, err := svc.GenerateA1(context.Background(), A1Request{Owner: "o", ProjectID: "p", Segments: []string{"第一章", "第二章"}})
	if err == nil || g.callCount() != 1 {
		t.Fatalf("extra JSON accepted or subsequent segment submitted: err=%v calls=%d", err, g.callCount())
	}
	docs, _ := store.ListDocuments(context.Background(), "o", "p")
	if len(docs) != 0 {
		t.Fatal("invalid compound response saved")
	}
}
func TestProfessional144ProviderErrorKeepsPreviousResult(t *testing.T) {
	svc, store, g := professionalService(t)
	old, err := svc.persist(context.Background(), "o", "p", "b", sampleScreenplay, 1)
	if err != nil {
		t.Fatal(err)
	}
	g.err = errors.New("mock provider failure")
	_, err = svc.GenerateB(context.Background(), BRequest{Owner: "o", ProjectID: "p", Novel: "原文"})
	if err == nil {
		t.Fatal("provider error accepted")
	}
	saved, err := store.ActiveDocument(context.Background(), "o", "p")
	if err != nil || saved.ID != old.Document.ID {
		t.Fatal("provider failure replaced previous result")
	}
}
