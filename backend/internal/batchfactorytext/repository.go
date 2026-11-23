package batchfactorytext

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("找不到该书或对应文本记录")
	ErrConflict = errors.New("内容版本已变化，请刷新后重试")
	ErrLocked   = errors.New("已上传的锁定页无法自动修改，请核对页面或恢复系统生成")
	ErrNotReady = errors.New("请先生成封面和所选正文页")
	ErrInvalid  = errors.New("文本生产数据无效")
)

type ConflictError struct{ Expected, Actual int64 }

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%v（期望 %d，当前 %d）", ErrConflict, e.Expected, e.Actual)
}
func (e *ConflictError) Unwrap() error { return ErrConflict }

type Repository interface {
	SaveOriginal(context.Context, Scope, string) (Document, error)
	GetDocument(context.Context, Scope) (Document, error)
	CreateVersion(context.Context, Scope, VersionInput) (Version, error)
	ListVersions(context.Context, Scope) ([]Version, error)
	ReplaceSystemPages(context.Context, Scope, PaginationUpdate) (int64, error)
	ListPages(context.Context, Scope) ([]Page, error)
	PatchPage(context.Context, Scope, string, int64, PagePatch) (Page, error)
	LockPageArtifact(context.Context, Scope, string, int64, ArtifactFile) (PageArtifact, error)
	SaveSystemPageArtifact(context.Context, Scope, string, int64, ArtifactFile) (PageArtifact, error)
	RestoreSystemPage(context.Context, Scope, string, int64) (Page, error)
	SaveCover(context.Context, Scope, int64, ArtifactFile, bool) (PageArtifact, error)
	RestoreCover(context.Context, Scope, int64) error
	ListArtifacts(context.Context, Scope) ([]PageArtifact, error)
	CreateCollection(context.Context, Scope, CollectionInput) (Collection, error)
	GetCollection(context.Context, Scope, string) (Collection, error)
	ListCollections(context.Context, Scope) ([]Collection, error)
}
