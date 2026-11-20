package novelcreation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	prompt "qiantie/backend/internal/prompt/novelcreation"
)

type requiredShortReview struct {
	Score  *float64  `json:"score"`
	Issues *[]string `json:"issues"`
}

// Each repair attempt decodes into a fresh value. Otherwise omitted fields in
// a repaired reply could inherit values from a previous invalid reply.
func (r *requiredShortReview) UnmarshalJSON(raw []byte) error {
	type plain requiredShortReview
	var value plain
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(&value); e != nil {
		return e
	}
	*r = requiredShortReview(value)
	return nil
}

// Version 1 is frozen when a new short task is created. Old task journals keep
// their original requests and keys; review protocol changes must not replay them.
func shortReviewPasses(in RunContext, review shortReview) bool {
	return in.Agent != nil && review.Score >= in.Agent.PassScore && len(review.Issues) == 0
}

func shortReviewFailure(in RunContext, review shortReview, stage string) error {
	detail := ""
	if len(review.Issues) > 0 {
		detail = "；" + truncateRunes(strings.Join(review.Issues, "；"), 300)
	}
	return inputError(fmt.Sprintf("%s审核未通过（%.1f分，通过分%.1f）；修订次数已用完，未标记为成品%s", stage, review.Score, in.Agent.PassScore, detail))
}

func (p *Pipeline) reviewControlledShortOutline(ctx context.Context, in RunContext, sink EventSink, story *ShortStory, brief map[string]any, value ShortOutline, review shortReview) (ShortOutline, error) {
	remaining := in.Agent.MaxRevisionRounds
	for round := 1; !shortReviewPasses(in, review); round++ {
		if remaining <= 0 {
			return value, shortReviewFailure(in, review, "短篇大纲")
		}
		remaining--
		key := fmt.Sprintf(":controlled:%d", round)
		var revised ShortOutline
		if e := p.shortCall(ctx, in, sink, prompt.ShortOutlineRevision, key, RunPlanning, map[string]any{"creation": brief, "outline": value, "review": review, "pass_score": in.Agent.PassScore}, &revised, func() error { return validateShortOutline(revised, story.Input.TargetChapters) }, 16000); e != nil {
			return value, e
		}
		if story.Input.Title != "" {
			revised.Title = story.Input.Title
		}
		value = revised
		review = shortReview{}
		if e := p.shortCall(ctx, in, sink, prompt.ShortReviewPrompt, ":outline"+key, RunPlanning, map[string]any{"creation": brief, "stage": "outline", "outline": value}, &review, func() error { return validShortReview(review) }, 6000); e != nil {
			return value, e
		}
		story.OutlineAudit, _ = json.Marshal(review)
	}
	return value, nil
}

// All pieces of one actual chapter share its revision budget. Chunking a long
// chapter must not silently multiply the maximum number of revision requests.
func (p *Pipeline) reviewControlledShortBody(ctx context.Context, in RunContext, sink EventSink, story *ShortStory, brief map[string]any, outline *ShortOutline, chapter ShortOutlineChapter, value shortText, review shortReview, target, pieceIndex, pieceTotal int, key, memory string, remaining *int) (shortText, error) {
	for round := 1; ; round++ {
		wrongLength := needsLengthNormalization(value.Content, target)
		if shortReviewPasses(in, review) && (!wrongLength || !in.Agent.EnableLengthNorm) {
			if wrongLength {
				return value, inputError("短篇正文与目标字数偏差仍超过45%；字数整理已关闭，未额外调用修订、未标记为成品")
			}
			return value, nil
		}
		if *remaining <= 0 {
			if !shortReviewPasses(in, review) {
				return value, shortReviewFailure(in, review, "短篇正文")
			}
			return value, inputError("短篇正文与目标字数偏差仍超过45%；修订次数已用完，未标记为成品")
		}
		*remaining = *remaining - 1
		requestKey := fmt.Sprintf("%s:controlled:%d", key, round)
		piece := ShortChapter{Number: chapter.Number, Title: chapter.Title, Content: value.Content}
		var revised shortText
		if e := p.shortCall(ctx, in, sink, prompt.ShortBodyRevision, requestKey, RunWriting, map[string]any{"creation": brief, "outline": outline, "chapter": piece, "review": review, "target_words": target, "piece_index": pieceIndex, "piece_total": pieceTotal, "story_memory": memory, "pass_score": in.Agent.PassScore}, &revised, func() error { return validShortText(revised, 12000) }, target*2+4000); e != nil {
			return value, e
		}
		value = revised
		piece.Content = value.Content
		review = shortReview{}
		if e := p.shortCall(ctx, in, sink, prompt.ShortReviewPrompt, ":chapter"+requestKey, RunWriting, map[string]any{"creation": brief, "outline": outline, "stage": "chapter_piece", "chapter": piece, "piece_index": pieceIndex, "piece_total": pieceTotal, "story_memory": memory}, &review, func() error { return validShortReview(review) }, 6000); e != nil {
			return value, e
		}
		audit, _ := json.Marshal(review)
		story.ChapterAudits = append(story.ChapterAudits, audit)
	}
}
