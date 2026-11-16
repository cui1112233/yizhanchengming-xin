package novelcreation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"time"
)

type exportRecord struct {
	ID          string               `json:"id"`
	Filename    string               `json:"filename"`
	ContentType string               `json:"contentType"`
	Object      storedDocumentObject `json:"object"`
	Sources     []ChapterSource      `json:"sources"`
	Status      string               `json:"status"`
	Error       string               `json:"error,omitempty"`
	CreatedAt   time.Time            `json:"createdAt"`
}
type ExportReceipt struct {
	ID          string          `json:"id"`
	Filename    string          `json:"filename"`
	Status      string          `json:"status"`
	DownloadURL string          `json:"downloadUrl"`
	Sources     []ChapterSource `json:"sources"`
}

func (s *Service) PersistExport(ctx context.Context, owner, project, format string) (out ExportReceipt, err error) {
	if s.documents == nil {
		return out, inputError("未配置TOS，无法保存导出文件；可直接下载已保存正文")
	}
	var doc ExportDocument
	var record exportRecord
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		child := *s
		child.store = tx
		var e error
		doc, e = child.ExportProject(ctx, owner, project, format)
		if e != nil {
			return e
		}
		ch, e := tx.ListChapters(ctx, owner, project)
		if e != nil {
			return e
		}
		sum := sha256.Sum256([]byte(doc.Content))
		id := novelID("export")
		ownerHash := sha256.Sum256([]byte(owner))
		record = exportRecord{ID: id, Filename: doc.Filename, ContentType: doc.ContentType, Status: "uploading", CreatedAt: time.Now().UTC(), Sources: chapterSources(ch), Object: storedDocumentObject{Key: "novel-exports/" + hex.EncodeToString(ownerHash[:12]) + "/" + project + "/" + id, ContentType: doc.ContentType, Size: int64(len(doc.Content)), SHA256: hex.EncodeToString(sum[:])}}
		return putContent(ctx, tx, owner, project, "export", id, record.Status, time.Now().UTC().Add(3*time.Minute), record)
	})
	if err != nil {
		return out, err
	}
	uploadCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	err = s.documents.Put(uploadCtx, DocumentObject(record.Object), strings.NewReader(doc.Content))
	if err != nil {
		record.Status = "failed"
		record.Error = "导出文件上传失败"
	} else {
		record.Status = "ready"
	}
	saveErr := s.store.Atomic(ctx, owner, project, func(tx Store) error {
		return putContent(ctx, tx, owner, project, "export", record.ID, record.Status, time.Time{}, record)
	})
	if saveErr != nil {
		return out, saveErr
	}
	return ExportReceipt{ID: record.ID, Filename: record.Filename, Status: record.Status, Sources: record.Sources, DownloadURL: "/api/novel-creation/projects/" + project + "/exports/" + record.ID}, err
}
func (s *Service) DownloadExport(ctx context.Context, owner, project, id string) (ExportDocument, error) {
	var v exportRecord
	ok, e := readContent(ctx, s.store, owner, project, "export", id, &v)
	if e != nil {
		return ExportDocument{}, e
	}
	if !ok || v.Status != "ready" {
		return ExportDocument{}, inputError("导出文件尚未就绪")
	}
	if s.documents == nil {
		return ExportDocument{}, inputError("未配置TOS，无法读取导出文件")
	}
	r, e := s.documents.Open(ctx, DocumentObject(v.Object))
	if e != nil {
		return ExportDocument{}, e
	}
	defer r.Close()
	b, e := io.ReadAll(io.LimitReader(r, v.Object.Size+1))
	if e != nil {
		return ExportDocument{}, e
	}
	sum := sha256.Sum256(b)
	if int64(len(b)) != v.Object.Size || hex.EncodeToString(sum[:]) != v.Object.SHA256 {
		return ExportDocument{}, inputError("导出文件校验失败")
	}
	return ExportDocument{Filename: v.Filename, ContentType: v.ContentType, Content: string(b)}, nil
}
func (s *Service) GetShortResult(ctx context.Context, owner, runID string) (ChapterVersion, error) {
	run, e := s.store.GetRun(ctx, owner, runID)
	if e != nil {
		return ChapterVersion{}, e
	}
	session, e := s.store.GetSession(ctx, owner, run.SessionID)
	if e != nil {
		return ChapterVersion{}, e
	}
	if session.Mode != "short_original" && session.Mode != "short_imitation" {
		return ChapterVersion{}, inputError("该任务不是短篇创作")
	}
	chapters, e := s.store.ListChapters(ctx, owner, run.ProjectID)
	if e != nil {
		return ChapterVersion{}, e
	}
	for _, c := range chapters {
		versions, e := s.store.ListChapterVersions(ctx, owner, run.ProjectID, c.ID)
		if e != nil {
			return ChapterVersion{}, e
		}
		for _, v := range versions {
			if v.RunID == runID && strings.TrimSpace(v.Content) != "" {
				return v, nil
			}
		}
	}
	return ChapterVersion{}, inputError("暂无已保存的短篇正文")
}
func (s *Service) ExportShortResult(ctx context.Context, owner, runID, format string) (ExportDocument, error) {
	v, e := s.GetShortResult(ctx, owner, runID)
	if e != nil {
		return ExportDocument{}, e
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format != "txt" && format != "md" && format != "markdown" {
		return ExportDocument{}, inputError("导出格式仅支持markdown或txt")
	}
	if format == "txt" {
		return ExportDocument{Filename: "短篇.txt", ContentType: "text/plain; charset=utf-8", Content: v.Content + "\n"}, nil
	}
	return ExportDocument{Filename: "短篇.md", ContentType: "text/markdown; charset=utf-8", Content: "# 短篇\n\n" + v.Content + "\n"}, nil
}
