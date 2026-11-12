package aicut

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrInvalidSRT = errors.New("invalid srt subtitles")

type SRTCue struct {
	StartMS int64
	EndMS   int64
	Text    string
}

func ParseSRT(source string) ([]SRTCue, error) {
	normalized := strings.TrimPrefix(strings.ReplaceAll(strings.ReplaceAll(source, "\r\n", "\n"), "\r", "\n"), "\ufeff")
	if strings.TrimSpace(normalized) == "" {
		return nil, fmt.Errorf("%w: 没有读取到有效字幕", ErrInvalidSRT)
	}
	blocks := strings.Split(normalized, "\n\n")
	cues := make([]SRTCue, 0, len(blocks))
	for _, raw := range blocks {
		lines := strings.Split(strings.TrimSpace(raw), "\n")
		if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
			continue
		}
		lineIndex := 0
		if _, err := strconv.Atoi(strings.TrimSpace(lines[0])); err == nil {
			lineIndex = 1
		}
		ordinal := len(cues) + 1
		if lineIndex >= len(lines) {
			return nil, fmt.Errorf("%w: 第 %d 条字幕时间格式无效", ErrInvalidSRT, ordinal)
		}
		parts := strings.SplitN(strings.TrimSpace(lines[lineIndex]), "-->", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: 第 %d 条字幕时间格式无效", ErrInvalidSRT, ordinal)
		}
		endFields := strings.Fields(strings.TrimSpace(parts[1]))
		if len(endFields) == 0 {
			return nil, fmt.Errorf("%w: 第 %d 条字幕时间格式无效", ErrInvalidSRT, ordinal)
		}
		start, startErr := parseSRTTimestamp(parts[0])
		end, endErr := parseSRTTimestamp(endFields[0])
		if startErr != nil || endErr != nil {
			return nil, fmt.Errorf("%w: 第 %d 条字幕时间格式无效", ErrInvalidSRT, ordinal)
		}
		if end <= start {
			return nil, fmt.Errorf("%w: 第 %d 条字幕时间无效", ErrInvalidSRT, ordinal)
		}
		text := strings.TrimSpace(strings.Join(lines[lineIndex+1:], "\n"))
		if text == "" {
			continue
		}
		cues = append(cues, SRTCue{StartMS: start, EndMS: end, Text: text})
	}
	if len(cues) == 0 {
		return nil, fmt.Errorf("%w: 没有读取到有效字幕", ErrInvalidSRT)
	}
	return cues, nil
}

func parseSRTTimestamp(value string) (int64, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, ".", ","))
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0, ErrInvalidSRT
	}
	secondParts := strings.Split(parts[2], ",")
	if len(secondParts) != 2 || len(secondParts[1]) != 3 {
		return 0, ErrInvalidSRT
	}
	hours, errH := strconv.ParseInt(parts[0], 10, 64)
	minutes, errM := strconv.ParseInt(parts[1], 10, 64)
	seconds, errS := strconv.ParseInt(secondParts[0], 10, 64)
	millis, errMS := strconv.ParseInt(secondParts[1], 10, 64)
	if errH != nil || errM != nil || errS != nil || errMS != nil || hours < 0 || minutes < 0 || minutes > 59 || seconds < 0 || seconds > 59 || millis < 0 || millis > 999 {
		return 0, ErrInvalidSRT
	}
	return ((hours*60+minutes)*60+seconds)*1000 + millis, nil
}

func ApplySRTCues(timeline Timeline, filename string, cues []SRTCue, newID func(string) string) (Timeline, string, error) {
	if len(cues) == 0 || newID == nil {
		return Timeline{}, "", ErrInvalidSRT
	}
	tracks := make([]Track, len(timeline.Tracks))
	for index, track := range timeline.Tracks {
		tracks[index] = track
		tracks[index].Clips = append([]Clip(nil), track.Clips...)
	}
	target := -1
	for index := range tracks {
		if tracks[index].Kind == TrackSubtitle && !tracks[index].Locked && len(tracks[index].Clips) == 0 {
			target = index
			break
		}
	}
	name := strings.TrimSpace(strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)))
	if name == "" {
		name = "导入字幕"
	}
	if target < 0 {
		target = len(tracks)
		tracks = append(tracks, Track{ID: newID("track-subtitle"), Kind: TrackSubtitle, Name: name, Order: len(tracks), Clips: []Clip{}})
	}
	for _, cue := range cues {
		tracks[target].Clips = append(tracks[target].Clips, Clip{
			ID: newID("subtitle"), Kind: ClipSubtitle, Name: name, StartMS: cue.StartMS,
			DurationMS: cue.EndMS - cue.StartMS, SourceStartMS: 0, SourceDurationMS: cue.EndMS - cue.StartMS, Text: cue.Text,
		})
	}
	result := timeline
	result.Tracks = tracks
	if err := validateTimeline(result); err != nil {
		return Timeline{}, "", err
	}
	return result, tracks[target].ID, nil
}
