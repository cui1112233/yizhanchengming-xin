package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func synopsisTestClock() time.Time { return time.Date(2026, 11, 19, 3, 0, 0, 0, time.UTC) }

func TestNormalizeSynopsisResultValid(t *testing.T) {
	raw := `这是模型多余的解释。` + "```json\n" + `{"title":"雨夜的一盏灯","synopsis":"林川在一座南方小城长大。他自幼丧父，与母亲相依为命。高中毕业后，他进入一家老旧照相馆工作，意外发现父亲留下的日记里藏着一桩旧案。为了查清真相，他开始追查二十年前的那场雨夜车祸。随着线索一点点浮现，他发现母亲的沉默、照相馆老板的回避，都与那场事故有关。最终，他在一次暴雨中找回了被篡改的底片，也终于明白了父亲当年的选择。故事以真相大白与母子和解收尾。","audience":"男频","genre":"悬疑推理","tags":["悬疑","都市","复仇","成长","校园Extra"]}` + "\n```"
	got, err := NormalizeSynopsisResult(raw, synopsisTestClock())
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if got.Title != "雨夜的一盏灯" || got.Audience != "男频" || got.Genre != "悬疑推理" {
		t.Fatalf("fields wrong: %+v", got)
	}
	if len(got.Tags) != 4 {
		t.Fatalf("tags (3-char 校园Extra dropped) wrong: %v", got.Tags)
	}
	if !got.GeneratedAt.Equal(synopsisTestClock()) {
		t.Fatalf("generated_at wrong: %v", got.GeneratedAt)
	}
	if utf8.RuneCountInString(got.Synopsis) < 100 {
		t.Fatalf("synopsis too short: %d", utf8.RuneCountInString(got.Synopsis))
	}
}

func TestNormalizeSynopsisResultFallbacks(t *testing.T) {
	raw := `{"title":"","synopsis":"概要正文","audience":"少年频","genre":"奇幻修真","tags":["好看","精彩","甜","甜宠","逆袭","重生","穿越"]}`
	got, err := NormalizeSynopsisResult(raw, synopsisTestClock())
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if got.Title != "未命名故事" {
		t.Fatalf("title fallback: %s", got.Title)
	}
	if got.Audience != "女频" || got.Genre != "其他" {
		t.Fatalf("audience/genre fallback: %+v", got)
	}
	// 好看/精彩 各3字但过于笼统？洛水只限制 2-4 字，这里应保留；甜 1 字丢弃；甜宠重复丢弃；最多 6 个。
	for _, tag := range got.Tags {
		if tag == "甜" {
			t.Fatalf("1-char tag should be dropped: %v", got.Tags)
		}
	}
	seen := map[string]bool{}
	for _, tag := range got.Tags {
		if seen[tag] {
			t.Fatalf("duplicate tag: %v", got.Tags)
		}
		seen[tag] = true
	}
	if len(got.Tags) > 6 {
		t.Fatalf("too many tags: %v", got.Tags)
	}
}

func TestNormalizeSynopsisResultRejectsEmptyAndInvalid(t *testing.T) {
	if _, err := NormalizeSynopsisResult("模型没返回 JSON", synopsisTestClock()); err == nil {
		t.Fatalf("expected error on no JSON")
	}
	if _, err := NormalizeSynopsisResult(`{"synopsis":""}`, synopsisTestClock()); err == nil {
		t.Fatalf("expected error on empty synopsis")
	}
	if _, err := NormalizeSynopsisResult(`{"synopsis":`, synopsisTestClock()); err == nil {
		t.Fatalf("expected error on broken JSON")
	}
}

func TestGenerateSynopsisRejectsEmptySourceAndMissingModel(t *testing.T) {
	if _, err := GenerateSynopsis(context.Background(), "   ", func(context.Context, string, string) (string, error) { return "{}", nil }, synopsisTestClock()); err == nil {
		t.Fatalf("expected empty source rejection")
	}
	if _, err := GenerateSynopsis(context.Background(), "有原文", nil, synopsisTestClock()); err == nil {
		t.Fatalf("expected nil model rejection")
	}
}

func TestGenerateSynopsisTruncatesLongSourceAndCallsModel(t *testing.T) {
	long := strings.Repeat("林川走进雨夜的巷口。", 8000) // 远超 60000 runes
	var gotSystem, gotUser string
	got, err := GenerateSynopsis(context.Background(), long, func(_ context.Context, system, user string) (string, error) {
		gotSystem, gotUser = system, user
		return `{"title":"雨夜","synopsis":"林川走进巷口，雨夜里他发现了父亲留下的旧照片。照片背面写着一句话，让他停下了脚步。巷子尽头的灯忽明忽暗，像是在等他做出选择。他深吸一口气，推开了那扇久未开启的门。门后是一间尘封的暗房，墙上挂满了他从未见过的底片。","audience":"女频","genre":"现代言情","tags":["言情","都市","治愈"]}`, nil
	}, synopsisTestClock())
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.Contains(gotSystem, "300-500") || !strings.Contains(gotSystem, "audience") {
		t.Fatalf("system prompt missing luoshui rules")
	}
	if utf8.RuneCountInString(gotUser) > 60000+200 {
		t.Fatalf("source not truncated: %d", utf8.RuneCountInString(gotUser))
	}
	if got.Title != "雨夜" || got.Audience != "女频" {
		t.Fatalf("result wrong: %+v", got)
	}
}

func TestGenerateSynopsisPropagatesModelError(t *testing.T) {
	want := errors.New("模型调用失败")
	_, err := GenerateSynopsis(context.Background(), "原文", func(context.Context, string, string) (string, error) { return "", want }, synopsisTestClock())
	if !errors.Is(err, want) {
		t.Fatalf("expected model error, got %v", err)
	}
}
