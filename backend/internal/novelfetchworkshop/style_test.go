package novelfetchworkshop

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type styleProviderFunc func(context.Context, RewriteRequest) (string, error)

func (f styleProviderFunc) Complete(ctx context.Context, req RewriteRequest) (string, error) {
	return f(ctx, req)
}

func TestResolveStyleUsesAIOnlyWhenBookInfoIsNotExplicit(t *testing.T) {
	called := 0
	got := ResolveStyle(context.Background(), BookInfo{Genre: "现代悬疑", Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		called++
		return `{"style":"古风虐文","confidence":0.9,"reason":"正文古代背景"}`, nil
	}))
	if got.Style != "现代悬疑" || got.Source != "bookinfo" || called != 0 {
		t.Fatalf("got=%+v called=%d", got, called)
	}
}

func TestResolveStyleUsesPlatformLabelsBeforeAI(t *testing.T) {
	calls := 0
	got := ResolveStyle(context.Background(), BookInfo{Labels: "现代", Themes: "言情、婚恋", Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		calls++
		return `{"style":"玄幻","confidence":0.9,"reason":"正文奇幻设定"}`, nil
	}))
	if got.Style != "现代女主" || got.Source != "bookinfo" || calls != 0 {
		t.Fatalf("label style decision = %#v, calls=%d", got, calls)
	}
}

func TestResolveStyleFallsBackToAIAndKeepsUnknownPending(t *testing.T) {
	good := ResolveStyle(context.Background(), BookInfo{Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		return `{"style":"现代女主","confidence":0.9,"reason":"女主关系线突出"}`, nil
	}))
	if good.Style != "现代女主" || good.Source != "ai" || good.Confidence != 0.9 || good.Reason != "女主关系线突出" {
		t.Fatalf("good=%+v", good)
	}
	unknown := ResolveStyle(context.Background(), BookInfo{Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		return `{"style":"无法确定"}`, nil
	}))
	if unknown.Style != "" {
		t.Fatalf("unknown=%+v", unknown)
	}
}

func TestResolveStyleAcceptsCodeFenceAndPlainStyleResponses(t *testing.T) {
	for _, response := range []string{"```json\n{\"style\":\"现代女主\",\"confidence\":0.9,\"reason\":\"女主关系线突出\"}\n```"} {
		got := ResolveStyle(context.Background(), BookInfo{Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
			return response, nil
		}))
		if got.Style != "现代女主" {
			t.Fatalf("response %q resolved to %q", response, got.Style)
		}
	}
}

func TestResolveStyleRetriesTransientProviderFailure(t *testing.T) {
	calls := 0
	got := ResolveStyle(context.Background(), BookInfo{Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		calls++
		if calls == 1 {
			return "", errors.New("director provider returned HTTP 503")
		}
		return `{"style":"现代悬疑","confidence":0.9,"reason":"案件线索"}`, nil
	}))
	if calls != 2 || got.Style != "现代悬疑" {
		t.Fatalf("calls=%d decision=%+v", calls, got)
	}
}

func TestResolveStyleFailureIsBoundedAndSafe(t *testing.T) {
	for _, tc := range []struct {
		status string
		calls  int
	}{{"503", 3}, {"401", 1}} {
		calls := 0
		got := ResolveStyle(context.Background(), BookInfo{Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
			calls++
			return "", errors.New("director provider returned HTTP " + tc.status + " secret-token")
		}))
		if calls != tc.calls || got.Style != "" || !strings.Contains(got.Reason, tc.status) || strings.Contains(got.Reason, "secret-token") {
			t.Errorf("status=%s calls=%d decision=%+v", tc.status, calls, got)
		}
	}
}

func TestResolveStyleDoesNotRetryCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	got := ResolveStyle(ctx, BookInfo{Text: "正文"}, styleProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		calls++
		cancel()
		return "", context.Canceled
	}))
	if calls != 1 || got.Style != "" || !strings.Contains(got.Reason, "取消") {
		t.Fatalf("calls=%d decision=%+v", calls, got)
	}
}

func TestParseStyleAIResultKeepsValidClassificationAtLowConfidence(t *testing.T) {
	good := parseStyleAIResult(`{"style":"现代悬疑","confidence":0.88,"reason":"正文以案件线索推进"}`)
	if good.Style != "现代悬疑" || good.Confidence != 0.88 || good.Reason == "" {
		t.Fatalf("good=%+v", good)
	}
	low := parseStyleAIResult(`{"style":"现代悬疑","confidence":0.6,"reason":"判断不充分"}`)
	if low.Style != "现代悬疑" || low.Confidence != 0.6 {
		t.Fatalf("valid low-confidence classification was discarded: %+v", low)
	}
	for _, response := range []string{`现代悬疑`} {
		got := parseStyleAIResult(response)
		if got.Style != "现代悬疑" || got.Confidence <= 0 {
			t.Errorf("plain style response %q should be accepted: %+v", response, got)
		}
	}
	withoutScore := parseStyleAIResult(`{"style":"现代悬疑","reason":"案件线索推动主线"}`)
	if withoutScore.Style != "现代悬疑" {
		t.Fatalf("valid label without score was discarded: %+v", withoutScore)
	}
	withoutReason := parseStyleAIResult(`{"style":"现代悬疑"}`)
	if withoutReason.Style != "现代悬疑" || withoutReason.Confidence <= 0 {
		t.Fatalf("valid label without reason was discarded: %+v", withoutReason)
	}
}

func TestMatchNovelStyleMapsCommonPlatformFacets(t *testing.T) {
	for value, want := range map[string]string{
		"现代言情": "现代女主",
		"都市":   "男频都市",
		"悬疑":   "现代悬疑",
		"古言":   "古风通用",
		"年代":   "年代通用",
		"打脸":   "职场打脸",
		"现实情感": "现代女主",
		"玄幻":   "玄幻",
	} {
		if got := matchNovelStyle(value); got != want {
			t.Errorf("matchNovelStyle(%q) = %q, want %q", value, got, want)
		}
	}
}

func TestClassifyBookAlwaysReturnsUsablePair(t *testing.T) {
	result := ClassifyBook(context.Background(), "", BookInfo{}, nil, nil)
	if result.Gender.Gender != "女频" || result.Style.Style != "现代通用" {
		t.Fatalf("empty classification result = %#v", result)
	}
	if result.Gender.Source != GenderFallback || result.Style.Source != "fallback" {
		t.Fatalf("fallback sources = %#v", result)
	}
	genderOnly := ClassifyBook(context.Background(), "", BookInfo{Category: "男频"}, nil, nil)
	if genderOnly.Style.Style != "男频都市" {
		t.Fatalf("style should follow gender: %#v", genderOnly)
	}
	styleOnly := ClassifyBook(context.Background(), "", BookInfo{Genre: "现代悬疑"}, nil, nil)
	if styleOnly.Gender.Gender != "女频" {
		t.Fatalf("gender should follow style: %#v", styleOnly)
	}
}
