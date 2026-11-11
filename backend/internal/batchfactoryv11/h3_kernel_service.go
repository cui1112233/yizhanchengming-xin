package batchfactoryv11

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type H3KernelCompileRequest struct {
	ExpectedCompilationID     string                            `json:"expected_compilation_id,omitempty"`
	AllowReplaceManualPrompts bool                              `json:"allow_replace_manual_prompts,omitempty"`
	FinalPromptOverrides      map[string]H3EditableCopyRevision `json:"final_prompt_overrides,omitempty"`
	DirectorRevisionID        string                            `json:"director_revision_id"`
	AudioAssetID              string                            `json:"audio_asset_id"`
	AllowSemanticTimeline     bool                              `json:"allow_semantic_timeline,omitempty"`
	Preset                    H3VideoPreset                     `json:"preset"`
	PrefixText                string                            `json:"prefix_text,omitempty"`
	QualityText               string                            `json:"quality_text,omitempty"`
	VisualRestrictionText     string                            `json:"visual_restriction_text,omitempty"`
	NegativeText              string                            `json:"negative_text,omitempty"`
	Switches                  H3PromptSwitches                  `json:"switches"`
	EditableCopyOverrides     map[string]H3EditableCopyRevision `json:"editable_copy_overrides,omitempty"`
}

type H3KernelCompileResult struct {
	Timeline    H3CanonicalTimelineRevision `json:"timeline"`
	Compilation H3VideoCompilationRevision  `json:"compilation"`
	Videos      []Video                     `json:"videos"`
}

type H3CompilationVideoMaterializer interface {
	ApplyH3CompilationVideos(context.Context, string, string, string, H3VideoCompilationRevision) ([]Video, error)
}

type H3KernelService struct {
	Store      Store
	AudioProbe H3AudioDurationProbe
}

func (s *H3KernelService) Compile(ctx context.Context, owner, batchID, bookID string, request H3KernelCompileRequest) (H3KernelCompileResult, error) {
	if s == nil || s.Store == nil {
		return H3KernelCompileResult{}, ErrUnavailable
	}
	repository, ok := s.Store.(H3Repository)
	if !ok {
		return H3KernelCompileResult{}, ErrUnavailable
	}
	materializer, ok := s.Store.(H3CompilationVideoMaterializer)
	if !ok {
		return H3KernelCompileResult{}, ErrUnavailable
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return H3KernelCompileResult{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return H3KernelCompileResult{}, err
	}
	if book.DirectorRevision == nil || book.DirectorRevision.ID != strings.TrimSpace(request.DirectorRevisionID) {
		return H3KernelCompileResult{}, fmt.Errorf("%w: 当前导演分镜与请求不一致，请刷新后重试", ErrConflict)
	}
	if book.DirectorRevision.Output.H3Director == nil {
		return H3KernelCompileResult{}, fmt.Errorf("%w: 当前导演分镜是旧版格式（缺少 H3 分镜文档），请重新生成导演分镜", ErrConflict)
	}
	document := *book.DirectorRevision.Output.H3Director
	previous, previousErr := repository.LatestH3VideoCompilation(ctx, owner, batchID, bookID, book.DirectorRevision.ID)
	if previousErr != nil && !errors.Is(previousErr, ErrNotFound) {
		return H3KernelCompileResult{}, previousErr
	}
	if request.ExpectedCompilationID != "" && (previousErr != nil || previous.ID != request.ExpectedCompilationID) {
		return H3KernelCompileResult{}, fmt.Errorf("%w: 分镜编译版本已变化，请刷新后编辑", ErrConflict)
	}
	if previousErr == nil && !request.AllowReplaceManualPrompts {
		for _, segment := range previous.Compilation.Segments {
			if segment.CompileTrace.EditableCopySource == "user_final_prompt" {
				if _, supplied := request.FinalPromptOverrides[segment.SegmentKey]; !supplied {
					return H3KernelCompileResult{}, fmt.Errorf("%w: 存在手动编辑的最终分镜提示词，重新编译前必须明确确认覆盖", ErrConflict)
				}
			}
		}
	}
	var timeline H3CanonicalTimeline
	if request.AllowSemanticTimeline {
		timeline, err = AllocateH3SemanticTimeline(book.DirectorRevision.ID, document)
	} else {
		audioRevision, audioErr := repository.GetH3AudioMeasurement(ctx, owner, batchID, bookID, strings.TrimSpace(request.AudioAssetID))
		if audioErr != nil {
			return H3KernelCompileResult{}, fmt.Errorf("%w: verified audio measurement is required", audioErr)
		}
		timeline, err = AllocateH3CanonicalTimeline(book.DirectorRevision.ID, document, audioRevision.Measurement)
	}
	if err != nil {
		return H3KernelCompileResult{}, err
	}
	timelineRevision, err := repository.PersistH3CanonicalTimeline(ctx, owner, batchID, bookID, timeline)
	if err != nil {
		return H3KernelCompileResult{}, err
	}
	analysis, err := h3AnalysisFromBook(document, book)
	if err != nil {
		return H3KernelCompileResult{}, err
	}
	compilation, err := CompileH3VideoSegments(H3VideoCompileInput{
		TimelineID:            timelineRevision.ID,
		Document:              document,
		Timeline:              timelineRevision.Timeline,
		Preset:                request.Preset,
		Analysis:              analysis,
		PrefixText:            request.PrefixText,
		QualityText:           request.QualityText,
		VisualRestrictionText: request.VisualRestrictionText,
		NegativeText:          request.NegativeText,
		Switches:              request.Switches,
		EditableCopyOverrides: request.EditableCopyOverrides,
		FinalPromptOverrides:  request.FinalPromptOverrides,
	})
	if err != nil {
		return H3KernelCompileResult{}, err
	}
	compilationRevision, err := repository.PersistH3VideoCompilation(ctx, owner, batchID, bookID, compilation)
	if err != nil {
		return H3KernelCompileResult{}, err
	}
	videos, err := materializer.ApplyH3CompilationVideos(ctx, owner, batchID, bookID, compilationRevision)
	if err != nil {
		return H3KernelCompileResult{}, err
	}
	return H3KernelCompileResult{Timeline: timelineRevision, Compilation: compilationRevision, Videos: videos}, nil
}

func h3AnalysisFromDirector(document H3DirectorDocument) H3AnalysisSnapshot {
	analysis := H3AnalysisSnapshot{
		VisualBaseline:    h3VisualBaselineText(document.VisualBaseline),
		CharacterSettings: make(map[string]string, len(document.CharacterRoster)),
		SceneSettings:     map[string]string{},
	}
	for _, character := range document.CharacterRoster {
		description := strings.TrimSpace(character.Appearance)
		if description == "" {
			description = strings.TrimSpace(character.CanonicalName)
		}
		analysis.CharacterSettings[character.SlotID] = description
	}
	for _, card := range document.DirectorCards {
		sceneID := strings.TrimSpace(card.Continuity.SceneID)
		if _, exists := analysis.SceneSettings[sceneID]; exists {
			continue
		}
		parts := []string{strings.TrimSpace(card.Continuity.Location), strings.TrimSpace(card.VisualContext)}
		analysis.SceneSettings[sceneID] = strings.Join(parts, "；")
	}
	return analysis
}

func (s *MemoryStore) ApplyH3CompilationVideos(_ context.Context, owner, batchID, bookID string, revision H3VideoCompilationRevision) ([]Video, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.h3Compilations[revision.ID]
	if !ok || stored.Owner != owner {
		return nil, ErrNotFound
	}
	revision = stored.Value
	if revision.BatchID != batchID || revision.BookID != bookID {
		return nil, ErrConflict
	}
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return nil, ErrNotFound
	}
	batch := owned.Value
	bookIndex := -1
	for index := range batch.Books {
		if batch.Books[index].ID == bookID {
			bookIndex = index
			break
		}
	}
	if bookIndex < 0 {
		return nil, ErrNotFound
	}
	book := &batch.Books[bookIndex]
	segments := revision.Compilation.Segments
	if len(segments) == 0 {
		return nil, ErrInvalid
	}
	now := time.Now().UTC()
	videos := make([]Video, len(segments))
	if len(book.Videos) == len(segments) {
		for index, segment := range segments {
			video := book.Videos[index]
			video.Label = fmt.Sprintf("VIDEO %02d", index+1)
			video.VideoPrompt = segment.EditableCopy
			video.DurationSeconds = float64(segment.CanonicalDurationMS) / 1000
			video.CompatibilityState = "active"
			video.Revision++
			videos[index] = video
		}
	} else {
		for index := range book.Videos {
			book.Videos[index].CompatibilityState = "orphaned"
		}
		for index, segment := range segments {
			videos[index] = Video{
				ID:                 s.id("video"),
				BatchID:            batchID,
				BookID:             bookID,
				Label:              fmt.Sprintf("VIDEO %02d", index+1),
				VideoPrompt:        segment.EditableCopy,
				DurationSeconds:    float64(segment.CanonicalDurationMS) / 1000,
				CompatibilityState: "active",
				Revision:           1,
			}
		}
	}
	book.Videos = videos
	book.Revision++
	batch.UpdatedAt = now
	s.batches[batchID] = memoryOwned[Batch]{Owner: owner, Value: batch}
	return append([]Video(nil), videos...), nil
}
