package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	catalog "qiantie/backend/internal/prompt"
	prompt "qiantie/backend/internal/prompt/novelcreation"
)

type ShortFullDraftState struct {
	CurrentVersion     string `json:"currentVersion"`
	RevisionStatus     string `json:"revisionStatus"`
	RevisionWarning    string `json:"revisionWarning,omitempty"`
	Review             string `json:"review,omitempty"`
	CompletionAttempts int    `json:"completionAttempts"`
	MissingChapters    []int  `json:"missingChapters"`
}

func upsertShortFullArtifact(story *ShortStory, key, title, content string) {
	for i := range story.Artifacts {
		if story.Artifacts[i].Key == key {
			story.Artifacts[i] = ShortArtifact{Key: key, Title: title, Content: content}
			return
		}
	}
	story.Artifacts = append(story.Artifacts, ShortArtifact{Key: key, Title: title, Content: content})
}

func refreshShortFullCurrentArtifact(story *ShortStory) error {
	if story.FullDraft == nil {
		return nil
	}
	version := story.FullDraft.CurrentVersion
	if version != "v001" && version != "v002" {
		return inputError("当前全篇稿件版本无效，未保存编辑")
	}
	var raw string
	found := false
	for _, artifact := range story.Artifacts {
		if artifact.Key == "drafts/"+version+".md" {
			raw = artifact.Content
			found = true
			break
		}
	}
	if !found {
		return inputError("缺少当前完整原稿，未保存编辑")
	}
	// Hook extraction uses only explicit native blocks. The loose numbered
	// Markdown fallback must never reparse manually edited chapter contents.
	hook := shortFullFirstBlock(raw, "SHORT_FICTION_OPENING_HOOK")
	if hook == "" {
		hook = shortFullFirstBlock(raw, "OPENING_HOOK")
	}
	current := shortFullDraft{Title: story.Title, OpeningHook: hook, Chapters: append([]ShortChapter(nil), story.Chapters...)}
	upsertShortFullArtifact(story, "drafts/final.md", "当前采用的完整正文", renderShortFullDraft(current))
	return nil
}

func shortFullOutlineMarkdown(story ShortStory) string {
	var text strings.Builder
	if story.Imitation != nil {
		text.WriteString(story.Imitation.Blueprint)
		text.WriteString("\n\n")
	}
	outline := story.Outline
	fmt.Fprintf(&text, "# %s\n\n%s\n", outline.Title, outline.Summary)
	for _, chapter := range outline.Chapters {
		fmt.Fprintf(&text, "\n## 第%d章 %s\n\n%s\n", chapter.Number, chapter.Title, chapter.Summary)
	}
	return text.String()
}

func shortFullPrompt(in RunContext, id, fallback string, fields prompt.ShortFullPromptInput) (string, error) {
	template, err := frozenSystemText(in, id, fallback)
	if err != nil {
		return "", err
	}
	block, err := frozenSystemText(in, prompt.ShortFullChapterBlockID, prompt.ShortFullChapterBlockTemplate)
	if err != nil {
		return "", err
	}
	return prompt.RenderShortFullPrompt(template, block, fields)
}

func shortFullSystem(in RunContext, definition catalog.Definition, writer bool) (string, error) {
	frozen, err := frozenSystemDefinition(in, definition, definition.ID)
	if err != nil {
		return "", err
	}
	if writer {
		addon, err := frozenSystemText(in, prompt.ShortImitationWritingID, prompt.ShortImitationWritingSystem)
		if err != nil {
			return "", err
		}
		frozen.SystemPrompt += "\n\n" + addon
	}
	return agentRequirementsSystemPrompt(in, frozen), nil
}

func (p *Pipeline) shortFullCall(ctx context.Context, in RunContext, sink EventSink, definition catalog.Definition, key, system, user string, tokens int, messages []batchfactoryv11.TextCompletionMessage, allowEmpty bool) (string, error) {
	in.allowEmptyCompletion = allowEmpty
	temperature := in.Agent.BaseTemperature
	switch definition.ID {
	case prompt.ShortFullWriter.ID:
		temperature = .58
	case prompt.ShortFullContinuation.ID:
		temperature = .68
	case prompt.ShortFullRevision.ID:
		temperature = .45
	case prompt.ShortFullReview.ID:
		temperature = .3
		tokens = 8192
	case prompt.ShortFullPackage.ID:
		temperature = .45
		tokens = 4096
	}
	request := batchfactoryv11.TextCompletionRequest{SystemPrompt: system, UserPrompt: user, Messages: messages, Temperature: temperature, MaxTokens: tokens, DisableJSONResponse: true}
	raw, err := p.complete(ctx, in, sink, nil, RunWriting, definition.ID+key, request, definition.Version)
	return raw, shortCallFailure(in, definition.ID+key, err)
}

func shortFullInvalid(in RunContext, key string, err error) error {
	return shortCallFailure(in, key, &DomainError{Code: ErrInvalidOutput, Message: err.Error(), Cause: err})
}

// The legacy JSON outline remains a CM planning adapter. From this boundary
// onward the new protocol uses whole Markdown drafts and native chapter blocks.
func (p *Pipeline) buildFullShortStory(ctx context.Context, in RunContext, sink EventSink, story ShortStory) (ShortStory, error) {
	if in.Agent == nil || story.Imitation == nil || story.Outline == nil {
		return story, inputError("全篇短篇缺少已确认的创作资料")
	}
	story.Status = "writing"
	story.FullDraft = &ShortFullDraftState{RevisionStatus: "drafting", MissingChapters: []int{}}
	fields := prompt.ShortFullPromptInput{Direction: story.Imitation.SelectedDirection, OutlineMarkdown: shortFullOutlineMarkdown(story), ChapterCount: story.Input.TargetChapters, CharsPerChapter: story.Input.TargetWordCount / story.Input.TargetChapters}
	creation, err := json.Marshal(shortBrief(story))
	if err != nil {
		return story, err
	}
	fields.Direction += "\n\n" + string(creation)
	writerSystem, err := shortFullSystem(in, prompt.ShortFullWriter, true)
	if err != nil {
		return story, err
	}
	firstUser, err := shortFullPrompt(in, prompt.ShortFullWriterUserID, prompt.ShortFullWriterUserTemplate, fields)
	if err != nil {
		return story, err
	}
	tokens := estimateShortFullTokens(fields.ChapterCount, fields.CharsPerChapter)
	if err = p.emitStage(ctx, in, sink, RunWriting, "生成全篇短篇首稿", 25); err != nil {
		return story, err
	}
	raw, err := p.shortFullCall(ctx, in, sink, prompt.ShortFullWriter, ":v001", writerSystem, firstUser, tokens, nil, false)
	if err != nil {
		return story, err
	}
	draft, err := parseShortFullDraft(raw, fields.ChapterCount)
	if err != nil {
		return story, shortFullInvalid(in, prompt.ShortFullWriter.ID+":v001", err)
	}
	story.FullDraft.MissingChapters = missingShortFullChapters(draft)
	for attempt := 1; len(story.FullDraft.MissingChapters) > 0 && attempt <= 3; attempt++ {
		upsertShortFullArtifact(&story, "drafts/v001-incomplete.md", "首稿补写过程（未完整）", raw)
		if err = saveShortStage(ctx, in, &story); err != nil {
			return story, err
		}
		fields.MissingChapters = append([]int(nil), story.FullDraft.MissingChapters...)
		fields.DraftMarkdown = renderShortFullDraft(draft)
		fields.DraftTitle = draft.Title
		user, e := shortFullPrompt(in, prompt.ShortFullContinuationUserID, prompt.ShortFullContinuationUserTemplate, fields)
		if e != nil {
			return story, e
		}
		if e = p.emitStage(ctx, in, sink, RunWriting, fmt.Sprintf("补齐全篇缺失章节 · 第%d/3次", attempt), 40+attempt*5); e != nil {
			return story, e
		}
		continuationTokens := estimateShortFullTokens(len(fields.MissingChapters), fields.CharsPerChapter)
		continuation, e := p.shortFullCall(ctx, in, sink, prompt.ShortFullContinuation, fmt.Sprintf(":v001:%d", attempt), writerSystem, user, continuationTokens, nil, false)
		if e != nil {
			return story, e
		}
		raw = shortFullTrim(raw) + "\n\n" + shortFullTrim(continuation)
		draft, e = parseShortFullDraft(raw, fields.ChapterCount)
		if e != nil {
			return story, shortFullInvalid(in, prompt.ShortFullContinuation.ID+fmt.Sprintf(":v001:%d", attempt), e)
		}
		story.FullDraft.CompletionAttempts = attempt
		story.FullDraft.MissingChapters = missingShortFullChapters(draft)
	}
	if err = validateShortFullDraft(draft, fields.ChapterCount); err != nil {
		upsertShortFullArtifact(&story, "drafts/v001-incomplete.md", "首稿补写过程（未完整）", raw)
		if saveErr := saveShortStage(ctx, in, &story); saveErr != nil {
			return story, saveErr
		}
		return story, shortFullInvalid(in, prompt.ShortFullContinuation.ID+":v001:3", err)
	}
	story.Title = draft.Title
	story.Chapters = append([]ShortChapter(nil), draft.Chapters...)
	story.FullDraft.CurrentVersion = "v001"
	story.FullDraft.RevisionStatus = "pending"
	fields.MissingChapters = nil
	upsertShortFullArtifact(&story, "drafts/v001.md", "完整首稿 v001", draft.RawContent)
	if err = saveShortStage(ctx, in, &story); err != nil {
		return story, err
	}
	fields.DraftMarkdown = renderShortFullDraft(draft)
	fields.DraftTitle = draft.Title
	reviewSystem, err := shortFullSystem(in, prompt.ShortFullReview, false)
	if err != nil {
		return story, err
	}
	reviewUser, err := shortFullPrompt(in, prompt.ShortFullReviewUserID, prompt.ShortFullReviewUserTemplate, fields)
	if err != nil {
		return story, err
	}
	if err = p.emitStage(ctx, in, sink, RunWriting, "全篇文字审稿", 65); err != nil {
		return story, err
	}
	review, err := p.shortFullCall(ctx, in, sink, prompt.ShortFullReview, ":v001", reviewSystem, reviewUser, tokens, nil, false)
	if err != nil {
		return story, err
	}
	story.FullDraft.Review = review
	fields.Review = review
	upsertShortFullArtifact(&story, "reviews/full-draft.md", "全篇首稿审稿意见", review)
	if err = saveShortStage(ctx, in, &story); err != nil {
		return story, err
	}
	revisionUser, err := shortFullPrompt(in, prompt.ShortFullRevisionUserID, prompt.ShortFullRevisionUserTemplate, fields)
	if err != nil {
		return story, err
	}
	messages := []batchfactoryv11.TextCompletionMessage{{Role: "system", Content: writerSystem}, {Role: "user", Content: firstUser}, {Role: "assistant", Content: shortFullTrim(draft.RawContent)}, {Role: "user", Content: revisionUser}}
	if err = p.emitStage(ctx, in, sink, RunWriting, "生成全篇第二稿", 75); err != nil {
		return story, err
	}
	revisedRaw, err := p.shortFullCall(ctx, in, sink, prompt.ShortFullRevision, ":v002", writerSystem, revisionUser, tokens, messages, true)
	if err != nil {
		return story, err
	}
	revised, parseErr := parseShortFullDraft(revisedRaw, fields.ChapterCount)
	if parseErr == nil {
		parseErr = validateShortFullDraft(revised, fields.ChapterCount)
	}
	if parseErr != nil {
		story.FullDraft.RevisionStatus = "rejected"
		story.FullDraft.RevisionWarning = "第二稿不完整或格式无效，已保留完整首稿 v001：" + parseErr.Error()
		upsertShortFullArtifact(&story, "drafts/v002-rejected.md", "未采用的第二稿", revisedRaw)
	} else {
		draft = revised
		story.Title = draft.Title
		story.Chapters = append([]ShortChapter(nil), draft.Chapters...)
		story.FullDraft.CurrentVersion = "v002"
		story.FullDraft.RevisionStatus = "adopted"
		upsertShortFullArtifact(&story, "drafts/v002.md", "采用的第二稿 v002", revisedRaw)
	}
	if err = saveShortStage(ctx, in, &story); err != nil {
		return story, err
	}
	fields.DraftMarkdown = renderShortFullDraft(draft)
	fields.DraftTitle = draft.Title
	packageSystem, err := shortFullSystem(in, prompt.ShortFullPackage, false)
	if err != nil {
		return story, err
	}
	packageUser, err := shortFullPrompt(in, prompt.ShortFullPackageUserID, prompt.ShortFullPackageUserTemplate, fields)
	if err != nil {
		return story, err
	}
	if err = p.emitStage(ctx, in, sink, RunWriting, "生成全篇包装文案", 90); err != nil {
		return story, err
	}
	packageRaw, err := p.shortFullCall(ctx, in, sink, prompt.ShortFullPackage, ":final", packageSystem, packageUser, tokens, nil, false)
	if err != nil {
		return story, err
	}
	upsertShortFullArtifact(&story, "package/report.md", "全篇包装原始结果", packageRaw)
	packaging := parseShortFullPackage(packageRaw, draft.Title)
	recommendation := []string{}
	if packaging.Intro != "" {
		recommendation = append(recommendation, packaging.Intro)
		upsertShortFullArtifact(&story, "package/intro.txt", "平台简介", packaging.Intro)
	}
	if len(packaging.SellingPoints) > 0 {
		points := "- " + strings.Join(packaging.SellingPoints, "\n- ")
		recommendation = append(recommendation, points)
		upsertShortFullArtifact(&story, "package/selling-points.md", "故事卖点", points)
	}
	if packaging.CoverPrompt != "" {
		upsertShortFullArtifact(&story, "package/cover-prompt.txt", "书封提示词", packaging.CoverPrompt)
	}
	story.Recommendation = strings.Join(recommendation, "\n\n")
	upsertShortFullArtifact(&story, "drafts/final.md", "当前采用的完整正文", renderShortFullDraft(draft))
	outlineRaw, _ := json.MarshalIndent(story.Outline, "", "  ")
	upsertShortFullArtifact(&story, "outline/story.json", "短篇大纲", string(outlineRaw))
	for _, chapter := range story.Chapters {
		upsertShortFullArtifact(&story, fmt.Sprintf("chapters/%03d.txt", chapter.Number), chapter.Title, chapter.Content)
	}
	upsertShortFullArtifact(&story, "recommendation.txt", "推荐文案", story.Recommendation)
	if err = saveShortStage(ctx, in, &story); err != nil {
		return story, err
	}
	if err = p.finishShortImitation(ctx, in, sink, &story); err != nil {
		return story, err
	}
	story.Status = "completed"
	return story, nil
}
