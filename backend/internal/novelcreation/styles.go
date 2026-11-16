package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type StyleStatistics struct {
	Characters         int     `json:"characters"`
	Paragraphs         int     `json:"paragraphs"`
	Sentences          int     `json:"sentences"`
	MeanSentenceLength float64 `json:"meanSentenceLength"`
	DialogueMarks      int     `json:"dialogueMarks"`
}
type StyleSource struct {
	Profile         StyleProfile      `json:"profile"`
	Source          KnowledgeMaterial `json:"source"`
	Statistics      StyleStatistics   `json:"statistics"`
	UserDescription string            `json:"userDescription"`
	AnalysisVersion int               `json:"analysisVersion"`
	AnalysisRunID   string            `json:"analysisRunId,omitempty"`
	Status          string            `json:"status"`
}

func styleStats(text string) StyleStatistics {
	v := StyleStatistics{Characters: utf8.RuneCountInString(text)}
	for _, p := range strings.Split(text, "\n") {
		if strings.TrimSpace(p) != "" {
			v.Paragraphs++
		}
	}
	for _, r := range text {
		if strings.ContainsRune("。！？.!?", r) {
			v.Sentences++
		}
		if strings.ContainsRune("“”\"", r) {
			v.DialogueMarks++
		}
	}
	if v.Sentences > 0 {
		v.MeanSentenceLength = float64(v.Characters) / float64(v.Sentences)
	}
	return v
}
func (s *Service) CreateStyleFromDocument(ctx context.Context, owner, project, document, title, description string) (out StyleSource, err error) {
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		j, e := getDocument(ctx, tx, owner, project, document)
		if e != nil {
			return e
		}
		if j.Selection == nil {
			return inputError("请先选择资料章节范围")
		}
		selected, e := selectedDocument(j, DocumentSelection{j.Selection.SourceVersion, j.Selection.SelectedChapterIDs})
		if e != nil {
			return e
		}
		out = StyleSource{Source: selected, Statistics: styleStats(selected.Content), UserDescription: strings.TrimSpace(description), Status: "statistics_ready"}
		content := fmt.Sprintf("文风描述：%s\n所选资料统计：%d字元，%d段，%d句，平均每句%.1f字元。", out.UserDescription, out.Statistics.Characters, out.Statistics.Paragraphs, out.Statistics.Sentences, out.Statistics.MeanSentenceLength)
		out.Profile, e = tx.SaveStyleProfile(ctx, owner, StyleProfile{Title: strings.TrimSpace(title), Content: content})
		if e != nil {
			return e
		}
		return putContent(ctx, tx, owner, project, "style", out.Profile.ID, out.Status, time.Time{}, out)
	})
	return
}
func (s *Service) GetStyleSource(ctx context.Context, owner, project, id string) (out StyleSource, err error) {
	ok, e := readContent(ctx, s.store, owner, project, "style", id, &out)
	if e != nil {
		return out, e
	}
	if !ok {
		return out, inputError("文风来源不存在")
	}
	out.Profile, err = s.store.GetStyleProfile(ctx, owner, id)
	if out.AnalysisRunID != "" {
		if run, e := s.store.GetRun(ctx, owner, out.AnalysisRunID); e == nil {
			out.Status = string(run.Status)
		}
	}
	return
}
func (s *Service) ListStyleSources(ctx context.Context, owner, project string) ([]StyleSource, error) {
	rows, e := s.store.ListContentRecords(ctx, owner, project, "style")
	if e != nil {
		return nil, e
	}
	out := []StyleSource{}
	for _, r := range rows {
		var v StyleSource
		if e = json.Unmarshal(r.Data, &v); e != nil {
			return nil, persistenceError(e)
		}
		v, e = s.GetStyleSource(ctx, owner, project, v.Profile.ID)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, nil
}
