package novelcreation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"qiantie/backend/internal/documenttext"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type DocumentObject struct {
	Key         string `json:"-"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	SHA256      string `json:"sha256"`
}

// Keys remain in durable private JSON; public jobs never expose them.
type storedDocumentObject struct {
	Key         string
	ContentType string
	Size        int64
	SHA256      string
}
type DocumentObjectStore interface {
	Put(context.Context, DocumentObject, io.Reader) error
	Open(context.Context, DocumentObject) (io.ReadCloser, error)
	Delete(context.Context, DocumentObject) error
}
type DocumentParser interface {
	Parse(context.Context, string, []byte) (documenttext.Result, error)
}
type MaterialChapter struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Number     int    `json:"number"`
	Characters int    `json:"characters"`
}
type DocumentJob struct {
	ID         string             `json:"id"`
	ProjectID  string             `json:"projectId"`
	Filename   string             `json:"filename"`
	Status     string             `json:"status"`
	Error      string             `json:"error,omitempty"`
	Attempt    int                `json:"attempt"`
	Version    int                `json:"version"`
	Object     DocumentObject     `json:"object"`
	Encoding   string             `json:"encoding,omitempty"`
	Chapters   []MaterialChapter  `json:"chapters"`
	EmptyPages []int              `json:"emptyPages,omitempty"`
	Selection  *KnowledgeMaterial `json:"selection,omitempty"`
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
}
type documentStored struct {
	Job    DocumentJob
	Object storedDocumentObject
}
type DocumentSelection struct {
	ExpectedVersion int      `json:"expectedVersion"`
	ChapterIDs      []string `json:"chapterIds"`
}
type DocumentSegments struct {
	ExpectedVersion int               `json:"expectedVersion"`
	Chapters        []MaterialChapter `json:"chapters"`
}

func (s *Service) ConfigureDocuments(objects DocumentObjectStore, parser DocumentParser) {
	s.documents = objects
	s.documentParser = parser
	if parser == nil {
		s.documentParser = documenttext.NewParser()
	}
}

func (s *Service) PasteDocument(ctx context.Context, owner, project, title, text string) (j DocumentJob, err error) {
	parser := documenttext.NewParser()
	result, e := parser.Parse(ctx, "paste.txt", []byte(text))
	if e != nil {
		return j, inputError(e.Error())
	}
	now := time.Now().UTC()
	j = DocumentJob{ID: novelID("document"), ProjectID: project, Filename: strings.TrimSpace(title), Status: "ready", Version: 1, Encoding: "UTF-8", Chapters: splitMaterialChapters(result.Text), CreatedAt: now, UpdatedAt: now}
	if j.Filename == "" {
		j.Filename = "粘贴文本"
	}
	if len(j.Chapters) > 2000 {
		return DocumentJob{}, inputError("文档超过2000个分段，请拆分")
	}
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error { return storeDocument(ctx, tx, owner, j, time.Time{}) })
	return
}
func (s *Service) DocumentAvailability() map[string]any {
	return map[string]any{"uploadAvailable": s.documents != nil, "reason": map[bool]string{true: "", false: "未配置TOS，文件上传与持久化导出暂不可用；仍可粘贴文本和直接下载已保存正文"}[s.documents != nil], "maxFileBytes": documenttext.MaxFileBytes, "maxBatchFiles": 8}
}
func storeDocument(ctx context.Context, st Store, owner string, j DocumentJob, until time.Time) error {
	o := j.Object
	return putContent(ctx, st, owner, j.ProjectID, "document", j.ID, j.Status, until, documentStored{Job: j, Object: storedDocumentObject{o.Key, o.ContentType, o.Size, o.SHA256}})
}
func getDocument(ctx context.Context, st Store, owner, project, id string) (DocumentJob, error) {
	var v documentStored
	ok, e := readContent(ctx, st, owner, project, "document", id, &v)
	if e != nil {
		return v.Job, e
	}
	if !ok || v.Job.Status == "deleted" {
		return v.Job, inputError("资料不存在或已删除")
	}
	v.Job.Object = DocumentObject(v.Object)
	return v.Job, nil
}
func (s *Service) GetDocument(ctx context.Context, owner, project, id string) (DocumentJob, error) {
	return getDocument(ctx, s.store, owner, project, id)
}
func (s *Service) ListDocuments(ctx context.Context, owner, project string) ([]DocumentJob, error) {
	rows, e := s.store.ListContentRecords(ctx, owner, project, "document")
	if e != nil {
		return nil, e
	}
	out := []DocumentJob{}
	for _, r := range rows {
		if r.Status == "deleted" {
			continue
		}
		var v documentStored
		if e = decodeRecord(r.Data, &v); e != nil {
			return nil, e
		}
		v.Job.Object = DocumentObject(v.Object)
		out = append(out, v.Job)
	}
	return out, nil
}
func (s *Service) UploadDocument(ctx context.Context, owner, project, name string, reader io.Reader) (j DocumentJob, err error) {
	if s.documents == nil {
		return j, inputError("未配置TOS，无法上传文件；请先粘贴文本")
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".txt" && ext != ".docx" && ext != ".pdf" {
		return j, inputError("仅支持TXT、DOCX和PDF")
	}
	body, e := io.ReadAll(io.LimitReader(reader, documenttext.MaxFileBytes+1))
	if e != nil {
		return j, e
	}
	if len(body) > documenttext.MaxFileBytes {
		return j, inputError("文件超过20MiB")
	}
	hash := sha256.Sum256(body)
	id := novelID("document")
	ownerHash := sha256.Sum256([]byte(owner))
	now := time.Now().UTC()
	j = DocumentJob{ID: id, ProjectID: project, Filename: filepath.Base(name), Status: "uploading", Version: 1, CreatedAt: now, UpdatedAt: now, Object: DocumentObject{Key: fmt.Sprintf("novel-documents/%x/%s/%s%s", ownerHash[:12], project, id, ext), Size: int64(len(body)), SHA256: hex.EncodeToString(hash[:]), ContentType: map[string]string{".txt": "text/plain; charset=utf-8", ".pdf": "application/pdf", ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}[ext]}}
	// Record the object intent before TOS: even a post-upload database outage
	// retains a discoverable object for retry/cleanup; no untracked orphan.
	if err = s.store.Atomic(ctx, owner, project, func(tx Store) error { return storeDocument(ctx, tx, owner, j, time.Now().UTC().Add(3*time.Minute)) }); err != nil {
		return j, err
	}
	uploadCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	err = s.documents.Put(uploadCtx, j.Object, bytes.NewReader(body))
	if err != nil {
		j.Status = "upload_failed"
		j.Error = "原件上传失败，请重新上传"
	} else {
		j.Status = "queued"
	}
	j.UpdatedAt = time.Now().UTC()
	saveErr := s.store.Atomic(ctx, owner, project, func(tx Store) error {
		latest, e := getDocument(ctx, tx, owner, project, j.ID)
		if e != nil {
			return e
		}
		if latest.Status != "uploading" || latest.Attempt != j.Attempt {
			return conflict()
		}
		return storeDocument(ctx, tx, owner, j, time.Time{})
	})
	if saveErr != nil {
		return j, saveErr
	}
	return j, err
}
func (s *Service) RetryDocument(ctx context.Context, owner, project, id string) (j DocumentJob, err error) {
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		var e error
		j, e = documentForMutation(ctx, tx, owner, project, id)
		if e != nil {
			return e
		}
		if j.Status != "failed" && j.Status != "uploading" {
			return inputError("当前资料状态不可重试")
		}
		j.Status = "queued"
		j.Error = ""
		return storeDocument(ctx, tx, owner, j, time.Time{})
	})
	return
}
func (s *Service) ParseDocument(ctx context.Context, owner, project, id string) (j DocumentJob, err error) {
	if s.documents == nil {
		return j, inputError("未配置TOS，无法读取原件")
	}
	var attempt int
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		r, ok, e := tx.GetContentRecord(ctx, owner, project, "document", id)
		if e != nil {
			return e
		}
		if !ok {
			return inputError("资料不存在")
		}
		now, e := tx.Clock(ctx)
		if e != nil {
			return e
		}
		if (r.Status != "queued" && r.Status != "parsing") || r.LeaseUntil.After(now) {
			return inputError("资料不在待解析状态")
		}
		j, e = getDocument(ctx, tx, owner, project, id)
		if e != nil {
			return e
		}
		j.Status = "parsing"
		j.Attempt++
		attempt = j.Attempt
		return storeDocument(ctx, tx, owner, j, now.Add(2*time.Minute))
	})
	if err != nil {
		return j, err
	}
	parseCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	r, e := s.documents.Open(parseCtx, j.Object)
	var result documenttext.Result
	if e == nil {
		var data []byte
		data, e = io.ReadAll(io.LimitReader(r, documenttext.MaxFileBytes+1))
		r.Close()
		if e == nil {
			sum := sha256.Sum256(data)
			if int64(len(data)) != j.Object.Size || hex.EncodeToString(sum[:]) != j.Object.SHA256 {
				e = inputError("原件大小或摘要不匹配，请重新上传")
			} else {
				parser := s.documentParser
				if parser == nil {
					parser = documenttext.NewParser()
				}
				result, e = parser.Parse(parseCtx, j.Filename, data)
			}
		}
	}
	parseErr := e
	var parsedChapters []MaterialChapter
	if parseErr == nil {
		parsedChapters = splitMaterialChapters(result.Text)
		if len(parsedChapters) > 2000 {
			parseErr = inputError("文档超过2000个分段，请拆分")
			parsedChapters = nil
		}
	}

	// Background bounded commit preserves failed/cancelled status. Attempt CAS
	// prevents an expired parser from replacing a newer attempt or deletion.
	saveCtx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	err = s.store.Atomic(saveCtx, owner, project, func(tx Store) error {
		latest, e := getDocument(saveCtx, tx, owner, project, id)
		if e != nil {
			return e
		}
		if latest.Attempt != attempt || latest.Status != "parsing" {
			return conflict()
		}
		j = latest
		j.UpdatedAt = time.Now().UTC()
		if parseErr != nil {
			j.Status = "failed"
			j.Error = parseErr.Error()
			j.Chapters = nil
		} else {
			j.Status = "ready"
			j.Error = ""
			j.Encoding = result.Encoding
			j.EmptyPages = result.EmptyPages
			j.Chapters = parsedChapters
		}
		return storeDocument(saveCtx, tx, owner, j, time.Time{})
	})
	if err == nil {
		err = parseErr
	}
	return
}

var chapterHeading = regexp.MustCompile(`(?i)^(第[零〇一二三四五六七八九十百千万两0-9]+[章节卷部回].*|chapter\s+[0-9ivxlcdm]+\b.*)$`)

func splitMaterialChapters(text string) []MaterialChapter {
	out := []MaterialChapter{}
	var lines []string
	title := "片段"
	flush := func() {
		content := strings.TrimSpace(strings.Join(lines, "\n"))
		if content != "" {
			out = append(out, MaterialChapter{ID: novelID("segment"), Number: len(out) + 1, Title: title, Content: content, Characters: utf8.RuneCountInString(content)})
		}
		lines = nil
	}
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if len([]rune(strings.TrimSpace(line))) <= 150 && chapterHeading.MatchString(strings.TrimSpace(line)) {
			flush()
			title = strings.TrimSpace(line)
		}
		lines = append(lines, line)
	}
	flush()
	return out
}
func (s *Service) SelectDocument(ctx context.Context, owner, project, id string, in DocumentSelection) (out KnowledgeMaterial, err error) {
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		j, e := getDocument(ctx, tx, owner, project, id)
		if e != nil {
			return e
		}
		out, e = selectedDocument(j, in)
		if e != nil {
			return e
		}
		j.Selection = &out
		return storeDocument(ctx, tx, owner, j, time.Time{})
	})
	return
}
func selectedDocument(j DocumentJob, in DocumentSelection) (out KnowledgeMaterial, err error) {
	if j.Status != "ready" || in.ExpectedVersion != j.Version {
		return out, inputError("资料未就绪或来源已改版，请重新选择")
	}
	if len(in.ChapterIDs) == 0 {
		return out, inputError("请选择章节范围")
	}
	byID := map[string]MaterialChapter{}
	for _, c := range j.Chapters {
		byID[c.ID] = c
	}
	seen := map[string]bool{}
	var text []string
	for _, id := range in.ChapterIDs {
		c, ok := byID[id]
		if !ok || seen[id] {
			return out, inputError("章节范围无效，请重新选择")
		}
		seen[id] = true
		text = append(text, c.Content)
	}
	out = KnowledgeMaterial{ID: "docref-" + j.ID, Title: j.Filename, Content: strings.Join(text, "\n\n"), SourceVersion: j.Version, SelectedChapterIDs: append([]string(nil), in.ChapterIDs...)}
	out.Characters = utf8.RuneCountInString(out.Content)
	if out.Characters > maxKnowledgeMaterialRunes {
		return out, inputError("所选正文超过12000字元，请缩小范围")
	}
	return
}
func (s *Service) ResegmentDocument(ctx context.Context, owner, project, id string, in DocumentSegments) (j DocumentJob, err error) {
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		var e error
		j, e = getDocument(ctx, tx, owner, project, id)
		if e != nil {
			return e
		}
		if j.Status != "ready" || j.Version != in.ExpectedVersion {
			return conflict()
		}
		if len(in.Chapters) == 0 || len(in.Chapters) > 2000 {
			return inputError("分段数量须为1到2000")
		}
		total := 0
		for i := range in.Chapters {
			c := &in.Chapters[i]
			c.Content = strings.TrimSpace(c.Content)
			if c.Content == "" {
				return inputError("章节正文不能为空")
			}
			c.ID = novelID("segment")
			c.Number = i + 1
			c.Characters = utf8.RuneCountInString(c.Content)
			total += c.Characters
		}
		if total > documenttext.MaxCharacters {
			return inputError("正文超过200万字元")
		}
		j.Version++
		j.Selection = nil
		j.Chapters = in.Chapters
		return storeDocument(ctx, tx, owner, j, time.Time{})
	})
	return
}
func (s *Service) DeleteDocument(ctx context.Context, owner, project, id string) error {
	return s.store.Atomic(ctx, owner, project, func(tx Store) error {
		j, e := documentForMutation(ctx, tx, owner, project, id)
		if e != nil {
			return e
		}
		j.Status = "deleted"
		j.Chapters = nil
		j.Selection = nil
		j.Attempt++
		return storeDocument(ctx, tx, owner, j, time.Time{})
	})
}
func StartDocumentWorker(ctx context.Context, s *Service) {
	go func() {
		tick := time.NewTicker(2 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				rows, e := s.store.ListParseDispatch(ctx, 8)
				if e == nil {
					for _, r := range rows {
						_, _ = s.ParseDocument(ctx, r.Owner, r.ProjectID, r.ID)
					}
				}
			}
		}
	}()
}

// Called inside the project transaction. Do not discard an upload intent while
// its bounded external Put may still be writing the original object.
func documentForMutation(ctx context.Context, tx Store, owner, project, id string) (DocumentJob, error) {
	rec, ok, e := tx.GetContentRecord(ctx, owner, project, "document", id)
	if e != nil {
		return DocumentJob{}, e
	}
	if !ok {
		return DocumentJob{}, inputError("资料不存在")
	}
	if rec.Status == "uploading" {
		now, e := tx.Clock(ctx)
		if e != nil {
			return DocumentJob{}, e
		}
		if rec.LeaseUntil.After(now) {
			return DocumentJob{}, inputError("原件正在上传，请等待上传结束后再重试或删除")
		}
	}
	return getDocument(ctx, tx, owner, project, id)
}
