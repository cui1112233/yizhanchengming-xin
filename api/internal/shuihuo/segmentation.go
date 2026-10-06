package shuihuo

import (
	"encoding/json"
	"regexp"
	"strings"
)

var importedIndexPrefix = regexp.MustCompile(`^\s*\d+\s*\t\s*`)

func paragraphCandidates(text string) []Candidate {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	blocks := make([]string, 0)
	current := make([]string, 0)
	flush := func() {
		if len(current) == 0 {
			return
		}
		joined := strings.TrimSpace(strings.Join(current, "\n"))
		if joined != "" {
			blocks = append(blocks, joined)
		}
		current = current[:0]
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		current = append(current, strings.TrimSpace(line))
	}
	flush()
	return candidateRows(blocks)
}

func fixedCandidates(text string, linesPerSegment int) []Candidate {
	if linesPerSegment <= 0 {
		return nil
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := make([]string, 0)
	for _, line := range strings.Split(text, "\n") {
		if v := strings.TrimSpace(line); v != "" {
			lines = append(lines, v)
		}
	}
	blocks := make([]string, 0, (len(lines)+linesPerSegment-1)/linesPerSegment)
	for i := 0; i < len(lines); i += linesPerSegment {
		end := i + linesPerSegment
		if end > len(lines) {
			end = len(lines)
		}
		blocks = append(blocks, strings.Join(lines[i:end], "\n"))
	}
	return candidateRows(blocks)
}

func importedCandidates(text string) []Candidate {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	rows := make([]string, 0)
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = importedIndexPrefix.ReplaceAllString(line, "")
		if line != "" {
			rows = append(rows, line)
		}
	}
	return candidateRows(rows)
}

func candidateRows(values []string) []Candidate {
	result := make([]Candidate, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, Candidate{Text: value, Speaker: "旁白"})
		}
	}
	return result
}

func normalizeCandidates(values []Candidate) ([]Candidate, error) {
	out := make([]Candidate, 0, len(values))
	for _, c := range values {
		c.Text = strings.TrimSpace(c.Text)
		if c.Text == "" {
			continue
		}
		c.Speaker = normalizeSpeaker(c.Speaker)
		out = append(out, c)
	}
	if len(out) == 0 {
		return nil, ErrInvalid
	}
	return out, nil
}

func ParseSmartCandidates(output string) ([]Candidate, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, ErrInvalid
	}
	if strings.HasPrefix(output, "```") {
		lines := strings.Split(output, "\n")
		if len(lines) >= 2 {
			lines = lines[1:]
			if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
				lines = lines[:len(lines)-1]
			}
			output = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}
	var list []Candidate
	if err := json.Unmarshal([]byte(output), &list); err == nil {
		return normalizeCandidates(list)
	}
	var wrapped struct {
		Candidates []Candidate `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(output), &wrapped); err == nil && len(wrapped.Candidates) > 0 {
		return normalizeCandidates(wrapped.Candidates)
	}
	rows := importedCandidates(output)
	if len(rows) == 0 {
		return nil, ErrInvalid
	}
	return rows, nil
}
