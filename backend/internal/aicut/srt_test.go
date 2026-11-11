package aicut

import (
	"strings"
	"testing"
)

func TestParseSRTNormalizesAndPreservesMultilineText(t *testing.T) {
	cues, err := ParseSRT("\ufeff1\r\n00:00:01,250 --> 00:00:03,500\r\n第一行\r\n第二行\r\n\r\n2\r\n00:00:04.000 --> 00:00:05.100\r\n结束")
	if err != nil {
		t.Fatal(err)
	}
	if len(cues) != 2 {
		t.Fatalf("got %d cues, want 2", len(cues))
	}
	if cues[0].StartMS != 1250 || cues[0].EndMS != 3500 || cues[0].Text != "第一行\n第二行" {
		t.Fatalf("unexpected first cue: %#v", cues[0])
	}
	if cues[1].StartMS != 4000 || cues[1].EndMS != 5100 || cues[1].Text != "结束" {
		t.Fatalf("unexpected second cue: %#v", cues[1])
	}
}

func TestParseSRTRejectsEmptyAndMalformedInput(t *testing.T) {
	if _, err := ParseSRT(" \r\n\t"); err == nil || !strings.Contains(err.Error(), "没有读取到有效字幕") {
		t.Fatalf("empty input error = %v", err)
	}
	if _, err := ParseSRT("1\n00:00:03,000 --> 00:00:02,000\n倒序"); err == nil || !strings.Contains(err.Error(), "第 1 条字幕时间无效") {
		t.Fatalf("reversed cue error = %v", err)
	}
	if _, err := ParseSRT("1\nnot-a-time\n坏时间码"); err == nil || !strings.Contains(err.Error(), "第 1 条字幕时间格式无效") {
		t.Fatalf("bad timestamp error = %v", err)
	}
	if _, err := ParseSRT("1\n00:00:01,000 -->   \n坏结束时间"); err == nil || !strings.Contains(err.Error(), "第 1 条字幕时间格式无效") {
		t.Fatalf("missing end timestamp error = %v", err)
	}
}

func TestApplySRTCuesReusesEmptyTrackThenCreatesAnotherForDuplicateImport(t *testing.T) {
	timeline := Timeline{Revision: 4, Tracks: []Track{
		{ID: "video-1", Kind: TrackVideo, Name: "画面轨道", Order: 0},
		{ID: "subtitle-empty", Kind: TrackSubtitle, Name: "字幕轨道", Order: 1},
	}}
	cues := []SRTCue{{StartMS: 1000, EndMS: 2500, Text: "你好"}}
	next := 0
	id := func(prefix string) string { next++; return prefix + "-" + string(rune('0'+next)) }

	first, firstTrack, err := ApplySRTCues(timeline, "demo.srt", cues, id)
	if err != nil {
		t.Fatal(err)
	}
	if firstTrack != "subtitle-empty" || len(first.Tracks) != 2 || len(first.Tracks[1].Clips) != 1 {
		t.Fatalf("empty subtitle track was not reused: track=%q timeline=%#v", firstTrack, first)
	}
	clip := first.Tracks[1].Clips[0]
	if clip.StartMS != 1000 || clip.DurationMS != 1500 || clip.Text != "你好" || clip.Kind != ClipSubtitle {
		t.Fatalf("unexpected subtitle clip: %#v", clip)
	}

	second, secondTrack, err := ApplySRTCues(first, "demo.srt", cues, id)
	if err != nil {
		t.Fatal(err)
	}
	if secondTrack == firstTrack || len(second.Tracks) != 3 || len(second.Tracks[2].Clips) != 1 {
		t.Fatalf("duplicate import should create a new subtitle track: track=%q timeline=%#v", secondTrack, second)
	}
}
