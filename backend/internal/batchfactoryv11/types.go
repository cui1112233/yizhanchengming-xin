package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrConflict    = errors.New("revision conflict")
	ErrInvalid     = errors.New("invalid input")
	ErrUnavailable = errors.New("capability unavailable")
	// ErrUpstream marks a temporary provider, CDN, or object-store failure.
	// It is deliberately distinct from ErrConflict so callers can retry it
	// without telling users that their saved batch revision is stale.
	ErrUpstream = errors.New("temporary upstream failure")
	// ErrConfigRequired marks a user-fixable missing selection (for example a
	// comic asset prompt the book never chose). It is deliberately distinct
	// from ErrConflict so a missing setting is never reported as a stale
	// revision, and a stale revision is never reported as a missing setting.
	ErrConfigRequired = errors.New("configuration required")
	// ErrDirectorRevisionRequired marks an operation that needs an active
	// Director storyboard the book does not have yet. It is deliberately
	// distinct from ErrConflict so a missing storyboard is never reported as a
	// stale revision conflict.
	ErrDirectorRevisionRequired = errors.New("director revision required")
)

// DirectorRevisionRequiredMessage is the actionable text for
// ErrDirectorRevisionRequired. The API layer returns it verbatim.
const DirectorRevisionRequiredMessage = "当前书籍还没有导演分镜，请先生成导演分镜后再试（漫画模式无需视频提示词）"

// ConfigRequiredError carries the plain action text the user can follow. It
// unwraps to ErrConfigRequired without prefixing that sentinel onto the text.
type ConfigRequiredError struct {
	Message string
}

func (e ConfigRequiredError) Error() string {
	return e.Message
}

func (e ConfigRequiredError) Unwrap() error {
	return ErrConfigRequired
}

func NewConfigRequiredError(message string) error {
	return ConfigRequiredError{Message: message}
}

type SettingsPatch map[string]json.RawMessage

type SettingsState struct {
	Patch    SettingsPatch `json:"patch"`
	Revision int64         `json:"revision"`
}

type ScopeKind string

const (
	ScopeBatch ScopeKind = "batch"
	ScopeBook  ScopeKind = "book"
	ScopeVideo ScopeKind = "video"
)

type ScopeRef struct {
	Kind    ScopeKind `json:"kind"`
	BatchID string    `json:"batchId"`
	BookID  string    `json:"bookId,omitempty"`
	VideoID string    `json:"videoId,omitempty"`
}
type SettingsUpdate struct {
	Patch            SettingsPatch `json:"patch"`
	RestoreKeys      []string      `json:"restoreKeys,omitempty"`
	ExpectedRevision int64         `json:"expectedRevision"`
}
type SettingsResult struct {
	Scope    ScopeRef        `json:"scope"`
	Patch    SettingsPatch   `json:"patch"`
	Revision int64           `json:"revision"`
	Snapshot *ConfigSnapshot `json:"snapshot,omitempty"`
}

type Video struct {
	ID      string `json:"id"`
	BatchID string `json:"batchId"`
	BookID  string `json:"bookId"`
	Label   string `json:"label,omitempty"`
	// VideoPrompt is the Director/video-motion instruction compiled for the
	// video model. VisualPrompt is deliberately separate and only drives the
	// optional still-image generation chain.
	VideoPrompt           string            `json:"videoPrompt,omitempty"`
	AIOriginalVideoPrompt string            `json:"aiOriginalVideoPrompt,omitempty"`
	VisualPrompt          string            `json:"visualPrompt,omitempty"`
	ComicPrompt           string            `json:"comicPrompt,omitempty"`
	ComicImagePrompt      string            `json:"comicImagePrompt,omitempty"`
	ComicRoles            []string          `json:"comicRoles,omitempty"`
	ComicRoleStage        map[string]string `json:"comicRoleStage,omitempty"`
	ComicScenePresets     []string          `json:"comicScenePresets,omitempty"`
	DurationSeconds       float64           `json:"durationSeconds,omitempty"`
	// Ordinal 是这个分镜在书内的顺序。历史轮次（orphaned）记录也带 ordinal，
	// 前端据此把上一轮已生成成功的视频挂回同一个分镜。
	Ordinal            int    `json:"ordinal"`
	CompatibilityState string `json:"compatibilityState"`
	// 历史槽位（orphaned）补齐可播放信息：前端无需再依赖生产任务列表就能
	// 预览、播放并把该任务设为主版本。active 槽位不填这些字段。
	MediaURL           string              `json:"mediaUrl,omitempty"`
	ProductionTaskID   string              `json:"productionTaskId,omitempty"`
	Provider           string              `json:"provider,omitempty"`
	Succeeded          bool                `json:"succeeded"`
	GeneratedAt        time.Time           `json:"generatedAt,omitempty"`
	Revision           int64               `json:"revision"`
	SettingsState      SettingsState       `json:"settingsState"`
	ComicSourceSegment *ComicSourceSegment `json:"comicSourceSegment,omitempty"`
}
type Book struct {
	ID                  string            `json:"id"`
	BatchID             string            `json:"batchId"`
	BookID              string            `json:"bookId"`
	Title               string            `json:"title"`
	SourceText          string            `json:"sourceText,omitempty"`
	WorkingFrontContent string            `json:"workingFrontContent,omitempty"`
	SourceTaskID        string            `json:"sourceTaskId,omitempty"`
	Platform            string            `json:"platform,omitempty"`
	TxtText             string            `json:"txtText,omitempty"`
	TxtFileName         string            `json:"txtFileName,omitempty"`
	SourceMetadata      map[string]any    `json:"sourceMetadata,omitempty"`
	Revision            int64             `json:"revision"`
	SettingsState       SettingsState     `json:"settingsState"`
	Mode                string            `json:"mode,omitempty"`
	Hook                *HookRevision     `json:"hook,omitempty"`
	DirectorRevision    *DirectorRevision `json:"directorRevision,omitempty"`
	Assets              DirectorAssets    `json:"assets,omitempty"`
	AssetRecords        []BookAsset       `json:"assetRecords,omitempty"`
	Videos              []Video           `json:"videos"`
	// VideoHistory 保留被轮换成 orphaned 的历史视频记录（轻量字段）。它让前端能把
	// 上一轮已经生成成功的媒体继续展示为「上一版」，而不是随轮换凭空消失。
	VideoHistory []Video `json:"videoHistory,omitempty"`
}
type Batch struct {
	ID                     string `json:"id"`
	Title                  string `json:"title"`
	SourceIntakeID         string `json:"sourceIntakeId,omitempty"`
	ProjectCoverArtifactID string `json:"projectCoverArtifactId,omitempty"`
	ProjectCoverMediaURL   string `json:"projectCoverMediaUrl,omitempty"`
	// ProjectCoverImageURL 是项目卡片在没有合并成片时的图片封面回退
	// （取首个漫画页主图或首个资产主图），仅在摘要索引中填充。
	ProjectCoverImageURL string `json:"-"`
	// ProductionMode 是摘要索引里按设置推断出的项目类型（comic/video）。
	ProductionMode string        `json:"-"`
	Revision       int64         `json:"revision"`
	SettingsState  SettingsState `json:"settingsState"`
	Books          []Book        `json:"books"`
	CreatedAt      time.Time     `json:"createdAt,omitempty"`
	UpdatedAt      time.Time     `json:"updatedAt,omitempty"`
}
type CreateVideoInput struct {
	Label           string  `json:"label"`
	VideoPrompt     string  `json:"videoPrompt,omitempty"`
	VisualPrompt    string  `json:"visualPrompt,omitempty"`
	DurationSeconds float64 `json:"durationSeconds,omitempty"`
}
type CreateBookInput struct {
	ID             string             `json:"id,omitempty"`
	BookID         string             `json:"bookId,omitempty"`
	SourceTaskID   string             `json:"sourceTaskId,omitempty"`
	Title          string             `json:"title"`
	Platform       string             `json:"platform,omitempty"`
	SourceText     string             `json:"sourceText,omitempty"`
	TxtText        string             `json:"txtText,omitempty"`
	TxtFileName    string             `json:"txtFileName,omitempty"`
	SourceMetadata map[string]any     `json:"sourceMetadata,omitempty"`
	Videos         []CreateVideoInput `json:"videos,omitempty"`
}
type CreateBatchInput struct {
	Title string            `json:"title"`
	Books []CreateBookInput `json:"books,omitempty"`
	// InitialBatchSettings is written as the batch-level patch during creation.
	// It exists so a batch cannot be consumed from an intake before its selected
	// automation preset has been frozen into the batch itself.
	InitialBatchSettings SettingsPatch `json:"initialBatchSettings,omitempty"`
	// GiantAutomationPlan is copied into each newly created giant-material book.
	// The executor recovery path then reads the persisted plan instead of browser
	// state that can disappear after the batch creation request returns.
	GiantAutomationPlan map[string]any `json:"giantAutomationPlan,omitempty"`
}
type Intake struct {
	ID         string          `json:"id"`
	Owner      string          `json:"-"`
	Payload    json.RawMessage `json:"payload"`
	ConsumedAt *time.Time      `json:"consumedAt,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}
type NovelFetchIntakeInput struct {
	Books    []CreateBookInput `json:"books"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}
type ConfigVersion struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"createdAt,omitempty"`
}
type RenameConfigVersionInput struct {
	Name string `json:"name"`
}
type ConfigSnapshot struct {
	ID        string        `json:"id"`
	BatchID   string        `json:"batchId"`
	BookID    string        `json:"bookId,omitempty"`
	VideoID   string        `json:"videoId,omitempty"`
	Effective SettingsPatch `json:"effective"`
	CreatedAt time.Time     `json:"createdAt"`
}
type ChangeImpact struct {
	AffectedBooks       int      `json:"affectedBooks"`
	AffectedVideos      int      `json:"affectedVideos"`
	InvalidatesDirector bool     `json:"invalidatesDirector"`
	PreservesOverrides  bool     `json:"preservesOverrides"`
	ChangedKeys         []string `json:"changedKeys"`
}
type Prompt struct {
	ID        string    `json:"id"`
	VersionID string    `json:"versionId"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Content   string    `json:"content"`
	Revision  int64     `json:"revision"`
	CreatedAt time.Time `json:"createdAt"`
}
type Draft struct {
	Key       string    `json:"key"`
	Kind      string    `json:"kind"`
	Scope     string    `json:"scope"`
	Content   string    `json:"content"`
	Revision  int64     `json:"revision"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type HookRevision struct {
	ID           string     `json:"id"`
	BatchID      string     `json:"batchId"`
	BookID       string     `json:"bookId"`
	Revision     int64      `json:"revision"`
	Status       string     `json:"status"`
	Text         string     `json:"text"`
	SourceDigest string     `json:"sourceDigest"`
	CreatedAt    time.Time  `json:"createdAt"`
	ApprovedAt   *time.Time `json:"approvedAt,omitempty"`
}

type DirectorAssets struct {
	Characters []NamedPrompt `json:"characters"`
	Scenes     []NamedPrompt `json:"scenes"`
	Props      []NamedPrompt `json:"props"`
}

// BookAsset is a durable, book-scoped production asset. Director extraction
// can seed an asset, but manual edits stay attached to this book and never
// become a browser-only draft.
type BookAsset struct {
	ID                      string    `json:"id"`
	BatchID                 string    `json:"batchId"`
	BookID                  string    `json:"bookId"`
	Kind                    string    `json:"kind"`
	Name                    string    `json:"name"`
	FormalName              string    `json:"formalName,omitempty"`
	Facts                   string    `json:"facts,omitempty"`
	Appearance              string    `json:"appearance,omitempty"`
	Relationships           string    `json:"relationships,omitempty"`
	Prompt                  string    `json:"prompt"`
	ImageModelID            string    `json:"imageModelId"`
	ImageAspectRatio        string    `json:"imageAspectRatio"`
	Source                  string    `json:"source"`
	ExtractionPresetID      string    `json:"extractionPresetId,omitempty"`
	ExtractionPresetVersion int64     `json:"extractionPresetVersion,omitempty"`
	Revision                int64     `json:"revision"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

type CreateBookAssetInput struct {
	Kind             string `json:"kind"`
	Source           string `json:"source,omitempty"`
	Name             string `json:"name"`
	FormalName       string `json:"formalName,omitempty"`
	Facts            string `json:"facts,omitempty"`
	Appearance       string `json:"appearance,omitempty"`
	Relationships    string `json:"relationships,omitempty"`
	Prompt           string `json:"prompt"`
	ImageModelID     string `json:"imageModelId,omitempty"`
	ImageAspectRatio string `json:"imageAspectRatio,omitempty"`
}

type UpdateBookAssetInput struct {
	Name             string `json:"name"`
	FormalName       string `json:"formalName,omitempty"`
	Facts            string `json:"facts,omitempty"`
	Appearance       string `json:"appearance,omitempty"`
	Relationships    string `json:"relationships,omitempty"`
	Prompt           string `json:"prompt"`
	ImageModelID     string `json:"imageModelId,omitempty"`
	ImageAspectRatio string `json:"imageAspectRatio,omitempty"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

type BookAssetImage struct {
	ID           string    `json:"id"`
	AssetID      string    `json:"assetId"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnailUrl,omitempty"`
	StorageRef   string    `json:"-"`
	MediaType    string    `json:"mediaType"`
	Source       string    `json:"source"`
	IsPrimary    bool      `json:"isPrimary"`
	Revision     int64     `json:"revision"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CreateBookAssetImageInput struct {
	URL        string `json:"url"`
	StorageRef string `json:"-"`
	MediaType  string `json:"mediaType,omitempty"`
	Source     string `json:"source"`
}

type ComicPage struct {
	ID                 string   `json:"id"`
	BatchID            string   `json:"batchId"`
	BookID             string   `json:"bookId"`
	VideoID            string   `json:"videoId"`
	ComicPrompt        string   `json:"comicPrompt"`
	ImagePrompt        string   `json:"imagePrompt"`
	ReferenceImageURLs []string `json:"referenceImageUrls,omitempty"`
	ModelID            string   `json:"modelId,omitempty"`
	AspectRatio        string   `json:"aspectRatio"`
	Status             string   `json:"status"`
	// GenerationError 是最近一次页级生成失败的可展示文案：第一行中文主文案，
	// 原始信息拼在 "\n上游原文：" 之后。成功出图或重新认领时清空。
	GenerationError string `json:"generationError,omitempty"`
	SnapshotHash    string `json:"snapshotHash,omitempty"`
	ConfigStale     bool   `json:"configStale,omitempty"`
	// 以下三个字段由 HTTP 读接口派生，不落库；前端只渲染，不再自行重算。
	PromptStatus    string           `json:"promptStatus,omitempty"`
	ImageStatus     string           `json:"imageStatus,omitempty"`
	PrimaryFallback bool             `json:"primaryFallback,omitempty"`
	GenerationState string           `json:"generationState,omitempty"`
	Revision        int64            `json:"revision"`
	Images          []ComicPageImage `json:"images"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}

type UpsertComicPageInput struct {
	ComicPrompt        string   `json:"comicPrompt"`
	ImagePrompt        string   `json:"imagePrompt"`
	ReferenceImageURLs []string `json:"referenceImageUrls,omitempty"`
	Provider           string   `json:"provider,omitempty"`
	ModelID            string   `json:"modelId,omitempty"`
	AspectRatio        string   `json:"aspectRatio"`
	Status             string   `json:"status,omitempty"`
	SnapshotHash       string   `json:"-"`
	// GenerationError 由服务端在失败写回时填写，客户端不能通过请求体注入。
	GenerationError string `json:"-"`
	// ExpectedRevision 是内部条件写回版本号。大于 0 时只有当前行 revision
	// 与之相同才会更新，用于避免两个生成任务把同一页互相覆盖；HTTP 请求
	// 不读取该字段，因此保持 json:"-"。
	ExpectedRevision int64 `json:"-"`
}

// PatchComicPageInput is the HTTP contract for updating one comic page without
// replacing the whole row. Pointer fields distinguish "not provided" from
// "provided as empty"; an omitted field keeps the stored value. SnapshotHash is
// server-owned and is never accepted from a client.
type PatchComicPageInput struct {
	ComicPrompt        *string   `json:"comicPrompt,omitempty"`
	ImagePrompt        *string   `json:"imagePrompt,omitempty"`
	ReferenceImageURLs *[]string `json:"referenceImageUrls,omitempty"`
	ModelID            *string   `json:"modelId,omitempty"`
	AspectRatio        *string   `json:"aspectRatio,omitempty"`
	Status             *string   `json:"status,omitempty"`
}

type ComicPageImage struct {
	ID          string `json:"id"`
	BatchID     string `json:"-"`
	BookID      string `json:"-"`
	ComicPageID string `json:"comicPageId"`
	URL         string `json:"url,omitempty"`
	// ThumbnailURL 只给列表行/缩略图格用；详情和下载仍走 URL 的原图。
	ThumbnailURL string    `json:"thumbnailUrl,omitempty"`
	StorageRef   string    `json:"-"`
	MediaType    string    `json:"mediaType"`
	Source       string    `json:"source"`
	IsPrimary    bool      `json:"isPrimary"`
	Revision     int64     `json:"revision"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CreateComicPageImageInput struct {
	URL        string `json:"url,omitempty"`
	StorageRef string `json:"-"`
	MediaType  string `json:"mediaType,omitempty"`
	Source     string `json:"source"`
}

type ComicCollectionItem struct {
	Ordinal int    `json:"ordinal"`
	VideoID string `json:"videoId"`
	// ComicPageID/ImageID 供响应层拼缩略图地址，不对外暴露。
	ComicPageID  string `json:"comicPageId"`
	ImageID      string `json:"imageId"`
	URL          string `json:"url,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	StorageRef   string `json:"-"`
	MediaType    string `json:"mediaType"`
}

type ComicCollection struct {
	ID        string                `json:"id"`
	BatchID   string                `json:"batchId"`
	BookID    string                `json:"bookId"`
	Status    string                `json:"status"`
	Items     []ComicCollectionItem `json:"items"`
	CreatedAt time.Time             `json:"createdAt"`
}

// ComicPagePatcher is implemented by stores that can update a comic page with
// patch semantics. The HTTP layer uses it when available and falls back to the
// full-row UpsertComicPage contract otherwise.
type ComicPagePatcher interface {
	PatchComicPage(context.Context, string, string, string, string, PatchComicPageInput) (ComicPage, error)
}

type ComicPageStore interface {
	UpsertComicPage(context.Context, string, string, string, string, UpsertComicPageInput) (ComicPage, error)
	ListComicPages(context.Context, string, string, string) ([]ComicPage, error)
	CreateComicPageImage(context.Context, string, string, string, string, CreateComicPageImageInput) (ComicPageImage, error)
	SetPrimaryComicPageImage(context.Context, string, string, string, string, string) (ComicPageImage, error)
	SetComicPageReferenceImages(context.Context, string, string, string, string, []string) (ComicPage, error)
	ApplyComicPageReferenceImages(context.Context, string, string, string, string) (int, error)
	CreateComicCollection(context.Context, string, string, string, ...[]string) (ComicCollection, []string, error)
	ListComicCollections(context.Context, string, string, string) ([]ComicCollection, error)
}

// StaleComicPage 是“停在 generating 且长时间没有更新”的页定位信息。
// 只用于启动/定时回收，不参与正常读写。
type StaleComicPage struct {
	Owner     string
	BatchID   string
	BookID    string
	VideoID   string
	Revision  int64
	UpdatedAt time.Time
}

// StaleComicPageStore 是可选能力：实现它的存储层支持跨用户扫描并回收
// 过期 generating 页。未实现的存储层会被回收任务跳过。
type StaleComicPageStore interface {
	ListStaleGeneratingComicPages(context.Context, time.Time, int) ([]StaleComicPage, error)
	ExpireStaleGeneratingComicPage(context.Context, StaleComicPage, time.Time) (bool, error)
}

// ComicSourceIndex is the minimal source data needed to decide whether a comic
// page binding is stale. It lets the comic-pages read path avoid hydrating the
// whole batch (all books, videos and drafts) just to compare digests.
type ComicSourceIndex struct {
	SourceText   string
	VideoDigests map[string]string
}

type ComicSourceIndexStore interface {
	GetBookComicSourceIndex(context.Context, string, string, string) (ComicSourceIndex, bool, error)
}

type ComicSourceSegment struct {
	VideoID             string                `json:"videoId"`
	BatchID             string                `json:"batchId"`
	BookID              string                `json:"bookId"`
	Ordinal             int                   `json:"ordinal"`
	StartOffset         int                   `json:"startOffset"`
	EndOffset           int                   `json:"endOffset"`
	SourceText          string                `json:"sourceText"`
	CharacterCount      int                   `json:"characterCount"`
	BoundaryType        string                `json:"boundaryType"`
	Reason              string                `json:"reason"`
	SourceDigest        string                `json:"sourceDigest"`
	SegmentationVersion int64                 `json:"segmentationVersion"`
	PromptStale         bool                  `json:"promptStale"`
	MediaStale          bool                  `json:"mediaStale"`
	Revision            int64                 `json:"revision"`
	SafeBoundaries      []ComicSourceBoundary `json:"safeBoundaries,omitempty"`
}

type ComicSourceSegmentation struct {
	BatchID       string               `json:"batchId"`
	BookID        string               `json:"bookId"`
	SourceDigest  string               `json:"sourceDigest"`
	CurrentDigest string               `json:"currentDigest"`
	Stale         bool                 `json:"stale"`
	Version       int64                `json:"version"`
	Segments      []ComicSourceSegment `json:"segments"`
}

type ComicSourceSplitInput struct {
	Direction        string `json:"direction"`
	BoundaryOffset   int    `json:"boundaryOffset"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

type ComicSourceSegmentStore interface {
	ReplaceComicSourceSegments(context.Context, string, string, string, string, []ComicSourceSegmentDraft) (ComicSourceSegmentation, error)
	BindComicSourceSegments(context.Context, string, string, string, string, []string, []ComicSourceSegmentDraft) (ComicSourceSegmentation, error)
	GetComicSourceSegmentation(context.Context, string, string, string) (ComicSourceSegmentation, error)
	SplitComicSourceSegment(context.Context, string, string, string, string, ComicSourceSplitInput) (ComicSourceSegmentation, error)
}

type OrphanedOverride struct {
	VideoID string        `json:"videoId"`
	Patch   SettingsPatch `json:"patch"`
	State   string        `json:"state"`
}

type DirectorRevision struct {
	ID                string             `json:"id"`
	BatchID           string             `json:"batchId"`
	BookID            string             `json:"bookId"`
	Revision          int64              `json:"revision"`
	Mode              string             `json:"mode"`
	SnapshotID        string             `json:"snapshotId"`
	SourceDigest      string             `json:"sourceDigest"`
	HookRevisionID    string             `json:"hookRevisionId,omitempty"`
	Output            DirectorResult     `json:"output"`
	Videos            []Video            `json:"videos"`
	OrphanedOverrides []OrphanedOverride `json:"orphanedOverrides,omitempty"`
	CreatedAt         time.Time          `json:"createdAt"`
}

type HookRunInput struct {
	BatchID string `json:"batchId"`
	BookID  string `json:"bookId"`
}

type DirectorRunInput struct {
	BatchID string `json:"batchId"`
	BookID  string `json:"bookId"`
}

// DirectorPromptCandidate is a temporary, book-scoped regeneration result.
// Its prompts are keyed by the already-persisted VIDEO IDs so accepting one
// candidate never changes the formal director revision or VIDEO inventory.
type DirectorPromptCandidate struct {
	ID                 string            `json:"id"`
	SourceBookRevision int64             `json:"sourceBookRevision"`
	CreatedAt          time.Time         `json:"createdAt"`
	VideoPrompts       map[string]string `json:"videoPrompts"`
}

type DirectorSnapshot struct {
	Effective            SettingsPatch `json:"effective"`
	Mode                 string        `json:"mode"`
	MaxVideoDuration     int           `json:"maxVideoDuration"`
	AudioDurationSeconds float64       `json:"audioDurationSeconds,omitempty"`
	AudioTargetSeconds   int           `json:"audioTargetSeconds,omitempty"`
	FixedSingleVideo     bool          `json:"fixedSingleVideo"`
	AspectRatio          string        `json:"aspectRatio"`
}

type UpdateBookMetadataInput struct {
	Metadata map[string]any `json:"metadata"`
	// Title 非空时覆盖书名。发布按书名匹配漫画作品，占位书名或抓取书名有误时
	// 需要用户手工纠正，所以元数据更新必须能改标题。
	Title            string `json:"title,omitempty"`
	ExpectedRevision int64  `json:"expectedRevision"`
}

// CaptureBookSourceInput is deliberately fill-only. It repairs legacy/manual
// intake records that retained a book ID and platform but never persisted the
// fetched original; it must never replace an existing source text.
type CaptureBookSourceInput struct {
	SourceText    string `json:"sourceText"`
	RawSourceText string `json:"rawSourceText,omitempty"`
	ReplaceSource bool   `json:"replaceSource,omitempty"`
	SourceBookID  string `json:"sourceBookId,omitempty"`
	SourceTitle   string `json:"sourceTitle,omitempty"`
	// ReplaceTitle 让抓正文成功后的上游书名覆盖当前书名，包括用户粘贴或
	// 手工改过的书名。只有真实抓取路径会置 true；上传 TXT 等路径保持占位名
	// 才替换的旧规则。
	ReplaceTitle     bool           `json:"replaceTitle,omitempty"`
	SourceMetadata   map[string]any `json:"sourceMetadata,omitempty"`
	ExpectedRevision int64          `json:"expectedRevision"`
}

// generatedBookTitlePattern 匹配书城导入时的占位书名，例如「小说 793433」。
// 占位名里的数字是来源书城 ID，和内部 book.BookID 并不相同，所以不能只比较 bookID。
var generatedBookTitlePattern = regexp.MustCompile(`^小说\s*[_-]?\s*\d{3,}$`)

func shouldReplaceGeneratedBookTitle(currentTitle, bookID, sourceTitle string) bool {
	currentTitle = strings.TrimSpace(currentTitle)
	bookID = strings.TrimSpace(bookID)
	sourceTitle = strings.TrimSpace(sourceTitle)
	if sourceTitle == "" {
		return false
	}
	if currentTitle == "" || currentTitle == bookID || currentTitle == "小说 "+bookID {
		return true
	}
	return generatedBookTitlePattern.MatchString(currentTitle)
}

type Store interface {
	CreateIntake(context.Context, string, NovelFetchIntakeInput) (Intake, error)
	GetIntake(context.Context, string, string) (Intake, error)
	CreateBatchFromIntake(context.Context, string, string, CreateBatchInput) (Batch, error)
	AppendBooksFromIntake(context.Context, string, string, string, bool) (Batch, error)
	CreateBatch(context.Context, string, CreateBatchInput) (Batch, error)
	ListBatches(context.Context, string) ([]Batch, error)
	GetBatch(context.Context, string, string) (Batch, error)
	DeleteBook(context.Context, string, string, string) error
	DeleteBatch(context.Context, string, string) error
	UpdateBookMetadata(context.Context, string, string, string, UpdateBookMetadataInput) (Book, error)
	CaptureBookSource(context.Context, string, string, string, CaptureBookSourceInput) (Book, error)
	SaveSettings(context.Context, string, ScopeRef, SettingsUpdate) (SettingsResult, error)
	AcceptDirectorPromptCandidate(context.Context, string, string, string, string, string, int64) error
	ConfigVersions(context.Context, string) ([]ConfigVersion, error)
	CreateConfigVersion(context.Context, string, ConfigVersion) (ConfigVersion, error)
	RenameConfigVersion(context.Context, string, string, string) (ConfigVersion, error)
	ChangeImpact(context.Context, string, string, SettingsUpdate) (ChangeImpact, error)
	ListPrompts(context.Context, string, string) ([]Prompt, error)
	CreatePrompt(context.Context, string, Prompt) (Prompt, error)
	GetDraft(context.Context, string, string, string, string) (Draft, error)
	SaveDraft(context.Context, string, Draft) (Draft, error)
	ListBookAssets(context.Context, string, string, string) ([]BookAsset, error)
	CreateBookAsset(context.Context, string, string, string, CreateBookAssetInput) (BookAsset, error)
	UpdateBookAsset(context.Context, string, string, string, string, UpdateBookAssetInput) (BookAsset, error)
	DeleteBookAsset(context.Context, string, string, string, string) error
	ListBookAssetImages(context.Context, string, string, string, string) ([]BookAssetImage, error)
	CreateBookAssetImage(context.Context, string, string, string, string, CreateBookAssetImageInput) (BookAssetImage, error)
	SetPrimaryBookAssetImage(context.Context, string, string, string, string, string) (BookAssetImage, error)
	PersistExtractedBookAssets(context.Context, string, Book, DirectorSnapshot, DirectorAssets) ([]BookAsset, error)
	PersistComicExtractedBookAssets(context.Context, string, Book, DirectorSnapshot, DirectorAssets) ([]BookAsset, error)
	CreateHookRevision(context.Context, string, string, string, string, string) (HookRevision, error)
	ApproveHookRevision(context.Context, string, string, string, string) (HookRevision, error)
	LatestHookRevision(context.Context, string, string, string) (HookRevision, error)
	PersistDirectorRevision(context.Context, string, Book, DirectorSnapshot, string, string, DirectorResult) (DirectorRevision, error)
}

// ShuihuoConfigStore is optional so the legacy Store contract remains stable
// while the Shuihuo compatibility layer gains durable per-user settings.
type ShuihuoConfigStore interface {
	GetShuihuoConfig(context.Context, string) (map[string]any, error)
	SaveShuihuoConfig(context.Context, string, map[string]any) error
}

// RetentionDeletingStore is optional so existing in-memory stores retain the
// simple deletion contract while the MySQL implementation can durably queue
// physical local-merge cleanup.
type RetentionDeletingStore interface {
	DeleteBookWithRetention(context.Context, string, string, string, int) error
	DeleteBatchWithRetention(context.Context, string, string, int) error
}
