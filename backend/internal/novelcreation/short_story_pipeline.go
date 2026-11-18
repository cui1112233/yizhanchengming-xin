package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	catalog "qiantie/backend/internal/prompt"
	prompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"time"
)

type shortText struct {
	Content string `json:"content"`
}
type shortReview struct {
	Score  float64  `json:"score"`
	Issues []string `json:"issues"`
}

func (p *Pipeline) shortCall(ctx context.Context, in RunContext, sink EventSink, def catalog.Definition, key string, status RunStatus, payload any, target any, validate func() error, tokens int) error {
	_, e := p.callJSON(ctx, in, sink, nil, modelCallSpec{definition: def, key: key, step: status, userPayload: payload, temperature: .65, maxTokens: tokens, validate: validate}, target)
	return e
}
func validShortText(v shortText, max int) error {
	if strings.TrimSpace(v.Content) == "" || jsLength(v.Content) > max {
		return inputError("短篇输出为空或超出该步骤上限")
	}
	return nil
}
func validShortReview(v shortReview) error {
	if v.Score < 0 || v.Score > 100 || len(v.Issues) > 50 {
		return inputError("短篇审核结果无效")
	}
	return nil
}
func shortBrief(story ShortStory) map[string]any {
	in := story.Input
	return map[string]any{"mode": in.Mode, "title": in.Title, "requirements": in.Requirements, "targetWordCount": in.TargetWordCount, "targetChapters": in.TargetChapters, "readerPreferences": in.ReaderPreferences, "supplementaryInfo": in.SupplementaryInfo, "inspirationTags": in.InspirationTags, "styles": story.Styles, "notes": story.Notes, "linkedContext": story.LinkedContext}
}
func (p *Pipeline) shortSource(ctx context.Context, in RunContext, sink EventSink, story ShortStory) (string, error) {
	if story.Input.Mode != "short_imitation" {
		return "", nil
	}
	analyses := []string{}
	for i, chunk := range imitationChunks(story.Input.ReferenceText) {
		var text string
		if e := json.Unmarshal(chunk, &text); e != nil {
			return "", e
		}
		v := shortText{}
		if e := p.shortCall(ctx, in, sink, prompt.ShortSourceAnalysis, fmt.Sprintf(":chunk:%d", i), RunPlanning, map[string]any{"reference_chunk": text}, &v, func() error { return validShortText(v, 4000) }, 6000); e != nil {
			return "", e
		}
		analyses = append(analyses, v.Content)
	}
	for level := 0; len(analyses) > 1; level++ {
		next := []string{}
		for start := 0; start < len(analyses); start += 4 {
			end := start + 4
			if end > len(analyses) {
				end = len(analyses)
			}
			v := shortText{}
			if e := p.shortCall(ctx, in, sink, prompt.ShortSourceAnalysis, fmt.Sprintf(":merge:%d:%d", level, start), RunPlanning, map[string]any{"source_analyses": analyses[start:end]}, &v, func() error { return validShortText(v, 4000) }, 6000); e != nil {
				return "", e
			}
			next = append(next, v.Content)
		}
		analyses = next
	}
	return analyses[0], nil
}
func (s *Service) runShortStory(ctx context.Context, owner string, run Run) (Run, error) {
	p, ok := s.pipeline.(*Pipeline)
	if !ok {
		return s.finishWithError(ctx, owner, run, ErrModelUnavailable, "独立短篇执行器未配置", nil)
	}
	var snap struct {
		Story        ShortStory      `json:"short_story"`
		Confirmed    bool            `json:"short_outline_confirmed"`
		Settings     ProjectSettings `json:"project_settings"`
		RetryOfRunID string          `json:"retry_of_run_id"`
	}
	if e := json.Unmarshal(run.PromptSnapshot, &snap); e != nil {
		return s.finishWithError(ctx, owner, run, ErrPersistenceFailed, "短篇快照损坏", e)
	}
	project, e := s.store.GetProject(ctx, owner, run.ProjectID)
	if e != nil {
		return run, e
	}
	session, e := s.store.GetSession(ctx, owner, run.SessionID)
	if e != nil {
		return run, e
	}
	story := snap.Story
	story.RunID = run.ID
	if snap.RetryOfRunID != "" {
		g, ok := s.store.(*guardedStore)
		if !ok {
			return run, leaseLost()
		}
		err := guarded(ctx, g.Store, g.lease, func(tx Store, execution *Execution) error {
			var latest ShortStory
			found, e := readContent(ctx, tx, owner, run.ProjectID, shortStoryKind, "story", &latest)
			if e != nil {
				return e
			}
			if !found || (latest.RunID != snap.RetryOfRunID && latest.RunID != run.ID) {
				return conflict()
			}
			latest.RunID = run.ID
			latest.Status = "queued"
			latest.UpdatedAt = time.Now().UTC()
			return putContent(ctx, tx, owner, run.ProjectID, shortStoryKind, "story", latest.Status, time.Time{}, latest)
		})
		if err != nil {
			return s.recoverDurableWriteFailure(ctx, owner, run, err)
		}
	}
	agent := NormalizeAgentSettings(snap.Settings.Agent, 2000)
	in := RunContext{OwnerUsername: owner, Run: run, Project: project, Session: session, Runtime: s.runtime, Agent: &agent}
	sink := serviceEventSink{service: s, owner: owner}
	result, e := p.buildShortStory(ctx, in, sink, story, snap.Confirmed)
	if e != nil {
		if IsCode(e, ErrSecondaryConfirmation) {
			return s.finishModelCallConfirmation(ctx, owner, run)
		}
		if isDurableWriteFailure(e) || IsCode(e, ErrPersistenceFailed) {
			return s.recoverDurableWriteFailure(ctx, owner, run, e)
		}
		code, message := domainFailure(e)
		status := RunFailed
		if code == ErrResultUnknown {
			status = RunPaused
		}
		if code == ErrRunCancelled || errors.Is(e, context.Canceled) {
			status = RunCancelled
		}
		return s.finish(ctx, owner, run, status, code, message, nil, e)
	}
	if e = p.checkCancelled(ctx, in); e != nil {
		return s.finish(ctx, owner, run, RunCancelled, ErrRunCancelled, "短篇已取消", nil, e)
	}
	g, ok := s.store.(*guardedStore)
	if !ok {
		return run, leaseLost()
	}
	status := RunSucceeded
	eventType := "run.completed"
	message := "短篇成品已保存"
	progress := 100
	if result.Status == "awaiting_confirmation" {
		status = RunAwaitingConfirmation
		eventType = "run.checkpoint"
		message = "短篇大纲已保存，请修改或确认后继续"
		progress = 20
	}
	var out Run
	var event RunEvent
	e = guarded(ctx, g.Store, g.lease, func(tx Store, execution *Execution) error {
		old := ShortStory{}
		found, er := readContent(ctx, tx, owner, run.ProjectID, shortStoryKind, "story", &old)
		if er != nil {
			return er
		}
		if !found {
			return persistenceError(errors.New("short story missing"))
		}
		if old.RunID != run.ID {
			return conflict()
		}
		result.Version = old.Version + 1
		raw, _ := json.Marshal(map[string]any{"shortStoryId": result.ID, "shortOutline": result.Outline, "confirmationKind": "short_outline", "storyVersion": result.Version, "targetWordCount": result.Input.TargetWordCount, "targetChapters": result.Input.TargetChapters})
		result.UpdatedAt = time.Now().UTC()
		if er = putContent(ctx, tx, owner, run.ProjectID, shortStoryKind, "story", result.Status, time.Time{}, result); er != nil {
			return er
		}
		if result.Title != "" {
			if _, er = tx.RenameProject(ctx, owner, run.ProjectID, result.Title); er != nil {
				return er
			}
		}
		out, er = tx.UpdateRun(ctx, owner, RunUpdate{ID: run.ID, Status: status, RuntimeSnapshot: raw})
		if er != nil {
			return er
		}
		event, er = tx.AppendRunEvent(ctx, owner, RunEvent{RunID: run.ID, ProjectID: run.ProjectID, SessionID: run.SessionID, Type: eventType, Status: status, Message: message, Progress: progress})
		if er != nil {
			return er
		}
		execution.State = "done"
		if status == RunAwaitingConfirmation {
			execution.State = "awaiting_confirmation"
		}
		return nil
	})
	if e != nil {
		return s.recoverDurableWriteFailure(ctx, owner, run, e)
	}
	_ = s.runtime.Publish(ctx, event)
	return out, nil
}
func (p *Pipeline) buildShortStory(ctx context.Context, in RunContext, sink EventSink, story ShortStory, confirmed bool) (ShortStory, error) {
	brief := shortBrief(story)
	outline := story.Outline
	stage := func(label string, progress int, status RunStatus) error {
		return p.emitStage(ctx, in, sink, status, label, progress)
	}
	if !confirmed {
		if e := stage("拆解参考与生成短篇大纲", 5, RunPlanning); e != nil {
			return story, e
		}
		source, e := p.shortSource(ctx, in, sink, story)
		if e != nil {
			return story, e
		}
		v := ShortOutline{}
		if e = p.shortCall(ctx, in, sink, prompt.ShortOutlinePrompt, "", RunPlanning, map[string]any{"creation": brief, "reference_analysis": source}, &v, func() error { return validateShortOutline(v, story.Input.TargetChapters) }, 16000); e != nil {
			return story, e
		}
		if story.Input.Title != "" {
			v.Title = story.Input.Title
		}
		review := shortReview{}
		if e = stage("审阅短篇大纲", 15, RunPlanning); e != nil {
			return story, e
		}
		if e = p.shortCall(ctx, in, sink, prompt.ShortReviewPrompt, ":outline", RunPlanning, map[string]any{"creation": brief, "stage": "outline", "outline": v}, &review, func() error { return validShortReview(review) }, 6000); e != nil {
			return story, e
		}
		story.OutlineAudit, _ = json.Marshal(review)
		if len(review.Issues) > 0 {
			if e = stage("修订短篇大纲", 20, RunPlanning); e != nil {
				return story, e
			}
			var revised ShortOutline
			if e = p.shortCall(ctx, in, sink, prompt.ShortOutlineRevision, "", RunPlanning, map[string]any{"creation": brief, "outline": v, "review": review}, &revised, func() error { return validateShortOutline(revised, story.Input.TargetChapters) }, 16000); e != nil {
				return story, e
			}
			v = revised
			if story.Input.Title != "" {
				v.Title = story.Input.Title
			}
		}
		outline = &v
		story.Outline = outline
		story.Title = v.Title
		if story.Input.PauseAfterOutline {
			story.Status = "awaiting_confirmation"
			return story, nil
		}
	}
	if outline == nil {
		return story, inputError("缺少已确认短篇大纲")
	}
	if e := validateShortOutline(*outline, story.Input.TargetChapters); e != nil {
		return story, e
	}
	story.Chapters = []ShortChapter{}
	story.Artifacts = []ShortArtifact{}
	story.ChapterAudits = []json.RawMessage{}
	memory, ending := "", ""
	completed := []ShortOutlineChapter{}
	for i, c := range outline.Chapters {
		chapter := ShortChapter{Number: c.Number, Title: c.Title}
		target := story.Input.TargetWordCount / story.Input.TargetChapters
		if i == len(outline.Chapters)-1 {
			target += story.Input.TargetWordCount % story.Input.TargetChapters
		}
		pieces := (target + 3999) / 4000
		if pieces < 1 {
			pieces = 1
		}
		for j := 0; j < pieces; j++ {
			if e := stage(fmt.Sprintf("撰写短篇第%d章 · 第%d/%d段", i+1, j+1, pieces), 25+50*i/len(outline.Chapters), RunWriting); e != nil {
				return story, e
			}
			v := shortText{}
			partTarget := target / pieces
			if j == pieces-1 {
				partTarget += target % pieces
			}
			payload := map[string]any{"creation": brief, "outline": outline, "current_chapter": c, "completed_chapters": completed, "piece_index": j + 1, "piece_total": pieces, "target_piece_words": partTarget, "story_memory": memory, "previous_text_excerpt": ending}
			key := fmt.Sprintf(":%d:%d", i, j)
			if e := p.shortCall(ctx, in, sink, prompt.ShortBodyPrompt, key, RunWriting, payload, &v, func() error { return validShortText(v, 12000) }, partTarget*2+4000); e != nil {
				return story, e
			}
			if e := stage(fmt.Sprintf("审阅短篇第%d章 · 第%d/%d段", i+1, j+1, pieces), 80, RunWriting); e != nil {
				return story, e
			}
			review := shortReview{}
			piece := ShortChapter{Number: c.Number, Title: c.Title, Content: v.Content}
			if e := p.shortCall(ctx, in, sink, prompt.ShortReviewPrompt, ":chapter"+key, RunWriting, map[string]any{"creation": brief, "outline": outline, "stage": "chapter_piece", "chapter": piece, "piece_index": j + 1, "piece_total": pieces, "story_memory": memory}, &review, func() error { return validShortReview(review) }, 6000); e != nil {
				return story, e
			}
			audit, _ := json.Marshal(review)
			story.ChapterAudits = append(story.ChapterAudits, audit)
			if len(review.Issues) > 0 || needsLengthNormalization(v.Content, partTarget) {
				if e := stage(fmt.Sprintf("修订短篇第%d章 · 第%d/%d段", i+1, j+1, pieces), 85, RunWriting); e != nil {
					return story, e
				}
				var revised shortText
				if e := p.shortCall(ctx, in, sink, prompt.ShortBodyRevision, key, RunWriting, map[string]any{"creation": brief, "outline": outline, "chapter": piece, "review": review, "target_words": partTarget, "piece_index": j + 1, "piece_total": pieces, "story_memory": memory}, &revised, func() error { return validShortText(revised, 12000) }, partTarget*2+4000); e != nil {
					return story, e
				}
				v = revised
			}
			if needsLengthNormalization(v.Content, partTarget) {
				return story, inputError("短篇正文与目标字数偏差仍超过45%，已停止，不能标为成品")
			}
			if chapter.Content != "" {
				chapter.Content += "\n\n"
			}
			chapter.Content += v.Content
			runes := []rune(v.Content)
			if len(runes) > 2000 {
				runes = runes[len(runes)-2000:]
			}
			ending = string(runes)
			// Only bounded continuity records cross generation requests. Every accepted
			// piece contributes, including revisions, without resending full manuscripts.
			if j < pieces-1 || i < len(outline.Chapters)-1 {
				var next shortText
				if e := p.shortCall(ctx, in, sink, prompt.ShortContinuity, key, RunWriting, map[string]any{"previous_memory": memory, "new_text": v.Content}, &next, func() error { return validShortText(next, 4000) }, 6000); e != nil {
					return story, e
				}
				memory = next.Content
			}
		}
		story.Chapters = append(story.Chapters, chapter)
		completed = append(completed, c)
	}
	if e := stage("打包短篇成品与推荐", 95, RunWriting); e != nil {
		return story, e
	}
	recommend := shortText{}
	if e := p.shortCall(ctx, in, sink, prompt.ShortRecommendation, "", RunWriting, map[string]any{"title": story.Title, "outline": outline, "completed_chapters": completed, "story_memory": memory, "ending_excerpt": ending}, &recommend, func() error { return validShortText(recommend, 8000) }, 10000); e != nil {
		return story, e
	}
	story.Recommendation = recommend.Content
	raw, _ := json.MarshalIndent(outline, "", "  ")
	story.Artifacts = append(story.Artifacts, ShortArtifact{Key: "outline/story.json", Title: "短篇大纲", Content: string(raw)})
	for _, c := range story.Chapters {
		story.Artifacts = append(story.Artifacts, ShortArtifact{Key: fmt.Sprintf("chapters/%03d.txt", c.Number), Title: c.Title, Content: c.Content})
	}
	story.Artifacts = append(story.Artifacts, ShortArtifact{Key: "recommendation.txt", Title: "推荐文案", Content: story.Recommendation})
	story.Status = "completed"
	return story, nil
}
