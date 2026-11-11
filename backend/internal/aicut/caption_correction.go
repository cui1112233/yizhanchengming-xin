package aicut

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidCaptionCorrection = errors.New("invalid caption correction")

type CaptionCorrection struct {
	ClipID string `json:"clipId"`
	Text   string `json:"text"`
	Reason string `json:"reason,omitempty"`
}

func ParseCaptionCorrections(raw string) ([]CaptionCorrection, error) {
	cleaned := strings.TrimSpace(raw)
	if strings.HasPrefix(cleaned, "```") {
		if newline := strings.IndexByte(cleaned, '\n'); newline >= 0 {
			cleaned = cleaned[newline+1:]
		}
		cleaned = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(cleaned), "```"))
	}
	var envelope struct {
		Corrections []CaptionCorrection `json:"corrections"`
	}
	if cleaned == "" || json.Unmarshal([]byte(cleaned), &envelope) != nil || envelope.Corrections == nil {
		return nil, fmt.Errorf("%w: 模型没有返回有效的字幕修正", ErrInvalidCaptionCorrection)
	}
	seen := make(map[string]struct{}, len(envelope.Corrections))
	for index := range envelope.Corrections {
		envelope.Corrections[index].ClipID = strings.TrimSpace(envelope.Corrections[index].ClipID)
		envelope.Corrections[index].Text = strings.TrimSpace(envelope.Corrections[index].Text)
		envelope.Corrections[index].Reason = strings.TrimSpace(envelope.Corrections[index].Reason)
		if envelope.Corrections[index].ClipID == "" || envelope.Corrections[index].Text == "" {
			return nil, fmt.Errorf("%w: 第 %d 条修正缺少字幕编号或正文", ErrInvalidCaptionCorrection, index+1)
		}
		if _, exists := seen[envelope.Corrections[index].ClipID]; exists {
			return nil, fmt.Errorf("%w: 字幕 %s 被重复修正", ErrInvalidCaptionCorrection, envelope.Corrections[index].ClipID)
		}
		seen[envelope.Corrections[index].ClipID] = struct{}{}
	}
	return envelope.Corrections, nil
}

func ApplyCaptionCorrections(timeline Timeline, corrections []CaptionCorrection) (Timeline, int, error) {
	byID := make(map[string]CaptionCorrection, len(corrections))
	for _, correction := range corrections {
		byID[correction.ClipID] = correction
	}
	tracks := make([]Track, len(timeline.Tracks))
	applied := 0
	for trackIndex, track := range timeline.Tracks {
		tracks[trackIndex] = track
		tracks[trackIndex].Clips = append([]Clip(nil), track.Clips...)
		if track.Kind != TrackSubtitle || track.Locked {
			continue
		}
		for clipIndex := range tracks[trackIndex].Clips {
			clip := &tracks[trackIndex].Clips[clipIndex]
			correction, ok := byID[clip.ID]
			if !ok || clip.Kind != ClipSubtitle {
				continue
			}
			if clip.Text != correction.Text {
				clip.Text = correction.Text
				applied++
			}
		}
	}
	if applied == 0 {
		return Timeline{}, 0, fmt.Errorf("%w: 没有匹配到可修改的字幕", ErrInvalidCaptionCorrection)
	}
	result := Timeline{Revision: timeline.Revision, Tracks: tracks}
	if err := validateTimeline(result); err != nil {
		return Timeline{}, 0, err
	}
	return result, applied, nil
}
