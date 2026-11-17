package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const chapterImportKind = "chapter_import"

type ImportedChapterInput struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	WordCount *int   `json:"wordCount,omitempty"`
}
type ImportChaptersInput struct {
	BookID           string                 `json:"bookId,omitempty"`
	Title            string                 `json:"title,omitempty"`
	Chapters         []ImportedChapterInput `json:"chapters"`
	KnowledgeBookIDs []string               `json:"knowledgeBookIds"`
	IdempotencyKey   string                 `json:"idempotencyKey"`
}
type ImportChaptersResult struct {
	Book     Project   `json:"book"`
	Chapters []Chapter `json:"chapters"`
	Message  string    `json:"message"`
}
type importedStyleReference struct {
	Profile StyleProfile `json:"profile"`
	Digest  string       `json:"digest"`
}
type submissionReceipt struct {
	Fingerprint string
	ProjectID   string
	Result      json.RawMessage
}

// ImportChapters only persists supplied chapters. It never schedules generation
// or manufactures continuity facts, audits, or style analysis.
func (s *Service) ImportChapters(ctx context.Context, owner string, in ImportChaptersInput) (out ImportChaptersResult, err error) {
	if s == nil || s.store == nil {
		return out, &DomainError{Code: ErrPersistenceFailed, Message: "小说存储服务未配置"}
	}
	// Normalization must not mutate caller slices shared by concurrent retries.
	in.Chapters = append([]ImportedChapterInput(nil), in.Chapters...)
	in.KnowledgeBookIDs = append([]string(nil), in.KnowledgeBookIDs...)
	owner, err = requiredOwner(owner)
	if err != nil {
		return out, err
	}
	in.BookID = strings.TrimSpace(in.BookID)
	in.Title = strings.TrimSpace(in.Title)
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	if in.IdempotencyKey == "" || utf8.RuneCountInString(in.IdempotencyKey) > 128 {
		return out, inputError("请提供有效的导入请求编号")
	}
	if in.KnowledgeBookIDs == nil {
		in.KnowledgeBookIDs = []string{}
	}
	for i := range in.KnowledgeBookIDs {
		in.KnowledgeBookIDs[i] = strings.TrimSpace(in.KnowledgeBookIDs[i])
	}
	for i := range in.Chapters {
		in.Chapters[i].Title = strings.TrimSpace(in.Chapters[i].Title)
	}
	digest := fingerprint(in)
	err = s.store.AtomicCreate(ctx, owner, func(tx Store) error {
		receipt, e := tx.ReserveSubmission(ctx, owner, chapterImportKind, in.IdempotencyKey, digest)
		if e != nil {
			return e
		}
		if receipt.Fingerprint != digest {
			return conflict()
		}
		// Read successful receipts before checking references that may since have
		// been archived or removed. Replays return the original immutable result.
		if len(receipt.Result) > 0 {
			return persistenceError(json.Unmarshal(receipt.Result, &out))
		}
		if in.BookID == "" && in.Title == "" {
			return inputError("新建作品时请输入作品名称")
		}
		if utf8.RuneCountInString(in.Title) > 255 || len(in.BookID) > 96 || len(in.Chapters) == 0 || len(in.Chapters) > 500 || len(in.KnowledgeBookIDs) > 100 {
			return inputError("请检查作品名称、章节数量和所选文风")
		}
		for _, c := range in.Chapters {
			if strings.TrimSpace(c.Content) == "" || utf8.RuneCountInString(c.Content) > 200000 || utf8.RuneCountInString(c.Title) > 255 {
				return inputError("章节正文不能为空且不能超过20万字元，标题不能超过255字元")
			}
		}
		if in.BookID != "" {
			out.Book, e = tx.LockImportProject(ctx, owner, in.BookID)
		} else {
			out.Book, e = tx.CreateImportProject(ctx, owner, in.Title)
			if e == nil {
				out.Book, e = tx.LockImportProject(ctx, owner, out.Book.ID)
			}
		}
		if e != nil {
			return e
		}
		now, e := tx.Clock(ctx)
		if e != nil {
			return e
		}
		slot, e := tx.GetSlot(ctx, owner, out.Book.ID)
		if e != nil {
			return e
		}
		if slot.LeaseUntil.After(now) {
			return &DomainError{Code: ErrProjectLocked, Message: "作品正在执行任务，请先停止任务再导入"}
		}
		runs, e := tx.ListRuns(ctx, owner, out.Book.ID)
		if e != nil {
			return e
		}
		for _, r := range runs {
			if !isTerminalRunStatus(r.Status) {
				return &DomainError{Code: ErrRunInProgress, Message: "作品存在未结束的任务，请先处理任务再导入"}
			}
		}
		styles := make([]importedStyleReference, 0, len(in.KnowledgeBookIDs))
		seen := map[string]bool{}
		for _, id := range in.KnowledgeBookIDs {
			if id == "" || len(id) > 96 || seen[id] {
				return inputError("所选文风无效或重复，请重新选择")
			}
			seen[id] = true
			p, e := tx.GetStyleProfile(ctx, owner, id)
			if e != nil {
				return e
			}
			styles = append(styles, importedStyleReference{Profile: p, Digest: fingerprint(p)})
		}
		previous, e := tx.ListChapters(ctx, owner, out.Book.ID)
		if e != nil {
			return e
		}
		next := 1
		for _, c := range previous {
			if c.Number >= next {
				next = c.Number + 1
			}
		}
		first := next
		out.Chapters = make([]Chapter, 0, len(in.Chapters))
		for i, c := range in.Chapters {
			title := c.Title
			if title == "" {
				title = fmt.Sprintf("第%d章", i+1)
			}
			// Match this application's existing manual-edit count. This is deliberately
			// labelled rather than claiming the original JS UTF-16/nonspace formula.
			chapter := Chapter{ID: novelID("chapter"), ProjectID: out.Book.ID, Number: next + i, Title: title, Content: c.Content, WordCount: utf8.RuneCountInString(c.Content), AuditStatus: "imported", Version: 1, CreatedAt: now, UpdatedAt: now}
			meta, _ := json.Marshal(map[string]any{"source": chapterImportKind, "sourceOrder": i + 1, "title": title, "knowledgeBookIds": in.KnowledgeBookIDs, "styles": styles, "wordCountMethod": "unicode_code_points", "idempotencyKey": in.IdempotencyKey})
			version := ChapterVersion{ID: novelID("version"), ChapterID: chapter.ID, Version: 1, Content: chapter.Content, Snapshot: meta, CreatedAt: now}
			if e = tx.InsertImportedChapter(ctx, owner, chapter, version); e != nil {
				return e
			}
			out.Chapters = append(out.Chapters, chapter)
		}
		var state ContinuityState
		if _, e = readContent(ctx, tx, owner, out.Book.ID, "continuity", "current", &state); e != nil {
			return e
		}
		if state.DirtyFrom == 0 || first < state.DirtyFrom {
			state.DirtyFrom = first
		}
		state.Revision++
		if e = putContent(ctx, tx, owner, out.Book.ID, "continuity", "current", "dirty", time.Time{}, state); e != nil {
			return e
		}
		if e = tx.TouchProject(ctx, owner, out.Book.ID); e != nil {
			return e
		}
		if out.Book, e = tx.GetProject(ctx, owner, out.Book.ID); e != nil {
			return e
		}
		out.Message = "章节已导入"
		raw, e := json.Marshal(out)
		if e != nil {
			return persistenceError(e)
		}
		return tx.CompleteSubmission(ctx, owner, chapterImportKind, in.IdempotencyKey, submissionReceipt{Fingerprint: digest, ProjectID: out.Book.ID, Result: raw})
	})
	if err != nil {
		return ImportChaptersResult{}, err
	}
	return out, nil
}
