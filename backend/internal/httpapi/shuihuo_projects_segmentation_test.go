package httpapi

import (
	"archive/zip"
	"bytes"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestParseSmartShuihuoCandidatesAcceptsCandidatesAndSegments(t *testing.T) {
	for _, raw := range []string{
		`{"candidates":[{"text":"第一段","speaker":"旁白"}]}`,
		"```json\n{\"segments\":[{\"sourceText\":\"第二段\",\"subtitleText\":\"第二段\"}]}\n```",
	} {
		got, err := parseSmartShuihuoCandidates(raw, "book-1")
		if err != nil {
			t.Fatalf("parse candidates: %v", err)
		}
		if len(got) != 1 || got[0]["text"] == "" || got[0]["id"] == "" {
			t.Fatalf("unexpected candidates: %#v", got)
		}
	}
}

func TestParseSmartShuihuoCandidatesRejectsInvalidOutput(t *testing.T) {
	if _, err := parseSmartShuihuoCandidates("not-json", "book-1"); err == nil {
		t.Fatal("expected invalid model output to fail")
	}
}

func TestFixedShuihuoCandidatesGroupsLines(t *testing.T) {
	got := fixedShuihuoCandidates("book-1", "一\n二\n三", 2)
	if len(got) != 2 || got[0]["text"] != "一\n二" || got[1]["text"] != "三" {
		t.Fatalf("unexpected fixed candidates: %#v", got)
	}
}

func TestParseShuihuoPromptCandidates(t *testing.T) {
	got, err := parseShuihuoPromptCandidates("```json\n{\"candidates\":[{\"segmentId\":\"s1\",\"prompt\":\"夜景\"}]}\n```")
	if err != nil || len(got) != 1 || got[0]["segmentId"] != "s1" {
		t.Fatalf("unexpected prompt candidates: %#v, err=%v", got, err)
	}
}

func TestShuihuoImageGenerationEndpointAndSize(t *testing.T) {
	got, err := shuihuoImageGenerationEndpoint("https://images.example/v1")
	if err != nil || got != "https://images.example/v1/images/generations" {
		t.Fatalf("endpoint=%q err=%v", got, err)
	}
	if shuihuoImageSize("9:16") != "1024x1792" || shuihuoImageSize("1:1") != "1024x1024" {
		t.Fatalf("unexpected image sizes")
	}
}

func TestExtractShuihuoDocxText(t *testing.T) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	entry, err := archive.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = entry.Write([]byte(`<w:document xmlns:w="x"><w:body><w:p><w:r><w:t>第一段</w:t></w:r></w:p><w:p><w:r><w:t>第二段</w:t></w:r></w:p></w:body></w:document>`))
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := extractShuihuoDocxText(buffer.Bytes())
	if err != nil || got != "第一段\n第二段" {
		t.Fatalf("text=%q err=%v", got, err)
	}
}

func TestShuihuoSegmentVoiceSettingsPersist(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 4, Store: batchfactoryv11.NewMemoryStore()})
	created := signedJSONRequest(t, api, now, "voice-user", http.MethodPost, "/api/shuihuo-production/projects", map[string]any{"name": "voice settings", "sourceText": "第一句"})
	if created.Code != http.StatusCreated {
		t.Fatalf("create=%d body=%s", created.Code, created.Body.String())
	}
	project := decodeBody[map[string]any](t, created)
	if project["id"] == nil {
		t.Fatalf("missing project: %s", created.Body.String())
	}
	projectID := project["id"].(string)
	candidates := signedJSONRequest(t, api, now, "voice-user", http.MethodPost, "/api/shuihuo-production/projects/"+projectID+"/segmentation/fixed", map[string]any{"text": "第一句", "linesPerSegment": 1})
	if candidates.Code != http.StatusOK {
		t.Fatalf("segment=%d body=%s", candidates.Code, candidates.Body.String())
	}
	segBody := decodeBody[map[string]any](t, candidates)
	items := segBody["candidates"].([]any)
	confirmed := signedJSONRequest(t, api, now, "voice-user", http.MethodPost, "/api/shuihuo-production/projects/"+projectID+"/segmentation/confirm", map[string]any{"candidates": items})
	if confirmed.Code != http.StatusOK {
		t.Fatalf("confirm=%d body=%s", confirmed.Code, confirmed.Body.String())
	}
	confirmedBody := decodeBody[map[string]any](t, confirmed)
	confirmedProject := confirmedBody["project"].(map[string]any)
	segments := confirmedProject["segments"].([]any)
	segmentID := segments[0].(map[string]any)["id"].(string)
	saved := signedJSONRequest(t, api, now, "voice-user", http.MethodPut, "/api/shuihuo-production/segments/"+segmentID+"/voice-settings", map[string]any{"voiceAssetId": "voice-1", "speechRate": 1.2, "pitch": 3})
	if saved.Code != http.StatusOK {
		t.Fatalf("save=%d body=%s", saved.Code, saved.Body.String())
	}
	settings := signedJSONRequest(t, api, now, "voice-user", http.MethodGet, "/api/shuihuo-production/projects/"+projectID+"/voice-settings", nil)
	if settings.Code != http.StatusOK {
		t.Fatalf("read=%d body=%s", settings.Code, settings.Body.String())
	}
	settingsBody := decodeBody[map[string]any](t, settings)
	bySegment := settingsBody["settings"].(map[string]any)
	got := bySegment[segmentID].(map[string]any)
	if got["voiceAssetId"] != "voice-1" || got["speechRate"].(float64) != 1.2 || got["pitch"].(float64) != 3 {
		t.Fatalf("unexpected settings: %#v", got)
	}
}
