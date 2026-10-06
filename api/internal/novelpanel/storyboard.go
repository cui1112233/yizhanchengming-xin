package novelpanel

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func sentenceFragments(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	result := make([]string, 0)
	start := 0
	for index, r := range line {
		if !strings.ContainsRune("。！？!?；;", r) {
			continue
		}
		end := index + utf8.RuneLen(r)
		fragment := strings.TrimSpace(line[start:end])
		if fragment != "" {
			result = append(result, fragment)
		}
		start = end
	}
	if start < len(line) {
		fragment := strings.TrimSpace(line[start:])
		if fragment != "" {
			result = append(result, fragment)
		}
	}
	if len(result) == 0 {
		return []string{line}
	}
	return result
}

func ScaffoldShots(text string, density Density) []Shot {
	lines := SourceLines(text)
	shots := make([]Shot, 0, len(lines))
	for _, line := range lines {
		fragments := []string{line.Text}
		switch density {
		case DensityDetailed:
			fragments = sentenceFragments(line.Text)
		case DensityStandard:
			if utf8.RuneCountInString(line.Text) > 100 {
				candidate := sentenceFragments(line.Text)
				if len(candidate) > 1 {
					fragments = candidate
				}
			}
		}
		for fragmentIndex, fragment := range fragments {
			shots = append(shots, Shot{
				ID:          fmt.Sprintf("shot_%d_%d", line.Index, fragmentIndex+1),
				SourceIndex: line.Index,
				SourceBasis: fragment,
			})
		}
	}
	return shots
}

func ValidateShots(sourceText string, shots []Shot) error {
	lines := SourceLines(sourceText)
	if len(lines) == 0 {
		if len(shots) == 0 {
			return nil
		}
		return fmt.Errorf("%w: storyboard exists without source text", ErrInvalid)
	}
	if len(shots) == 0 {
		return nil
	}
	byIndex := make(map[int]SourceLine, len(lines))
	coverage := make(map[int]int, len(lines))
	for _, line := range lines {
		byIndex[line.Index] = line
	}
	seenIDs := map[string]struct{}{}
	for index, shot := range shots {
		line, ok := byIndex[shot.SourceIndex]
		if !ok {
			return fmt.Errorf("%w: shot %d references source index %d outside original text", ErrInvalid, index+1, shot.SourceIndex)
		}
		basis := strings.TrimSpace(shot.SourceBasis)
		if basis == "" {
			return fmt.Errorf("%w: shot %d source basis is required", ErrInvalid, index+1)
		}
		if !strings.Contains(line.Text, basis) {
			return fmt.Errorf("%w: shot %d source basis is not present in source line %d", ErrInvalid, index+1, shot.SourceIndex)
		}
		if strings.TrimSpace(shot.Visual) == "" {
			return fmt.Errorf("%w: shot %d visual is required", ErrInvalid, index+1)
		}
		if shot.DurationSec < 0 || shot.DurationSec > 60 {
			return fmt.Errorf("%w: shot %d duration must be between 0 and 60 seconds", ErrInvalid, index+1)
		}
		if id := strings.TrimSpace(shot.ID); id != "" {
			if _, exists := seenIDs[id]; exists {
				return fmt.Errorf("%w: duplicate shot id %s", ErrInvalid, id)
			}
			seenIDs[id] = struct{}{}
		}
		coverage[shot.SourceIndex]++
	}
	for _, line := range lines {
		if coverage[line.Index] == 0 {
			return fmt.Errorf("%w: source line %d has no storyboard shot", ErrInvalid, line.Index)
		}
	}
	return nil
}
