package screenplay

import (
	"context"
	"time"
)

// BatchLog 是批次日志。
type BatchLog struct {
	ID        int64     `json:"id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// Store 是漫剧链路唯一的持久化端口。
//
// 批次写入有两条路径：
//   - UpdateBatch：revision 乐观锁，适用于「读到状态 → 计算 → 写回」且允许重试的场景；
//   - MutateBatch：事务内 SELECT ... FOR UPDATE 显式锁，适用于 run_next / cancel /
//     医生闸门等不能被并发覆盖的关键路径。
type Store interface {
	CreateProject(ctx context.Context, project Project) (Project, error)
	GetProject(ctx context.Context, owner, id string) (Project, error)
	ListProjects(ctx context.Context, owner string) ([]Project, error)
	UpdateProject(ctx context.Context, owner, id string, changes Project, expectedRevision int64) (Project, error)
	// SaveProjectSource 只写原文正文，不参与项目配置的 revision 乐观锁。
	SaveProjectSource(ctx context.Context, owner, id, content string) (Project, error)
	DeleteProject(ctx context.Context, owner, id string) error

	// SaveScreenplay 原子地追加一版剧本并替换该版本的场次/角色/场景。
	// 同一项目的并发写入由项目行显式锁串行化。
	SaveScreenplay(ctx context.Context, doc Document, shots []Shot, characters []Character, scenes []Scene) (Document, error)
	ActiveDocument(ctx context.Context, owner, projectID string) (Document, error)
	GetDocument(ctx context.Context, owner, projectID, documentID string) (Document, error)
	ListDocuments(ctx context.Context, owner, projectID string) ([]Document, error)
	// DeleteDocument 删除一个历史剧本版本（含该版本的场次/人物/场景）；当前活动版本由领域层拦截。
	DeleteDocument(ctx context.Context, owner, projectID, documentID string) error
	ListShots(ctx context.Context, owner, projectID, documentID string) ([]Shot, error)
	CreateShot(ctx context.Context, shot Shot) (Shot, error)
	UpdateShot(ctx context.Context, owner, projectID, shotID string, changes Shot) (Shot, error)
	MoveShot(ctx context.Context, owner, projectID, shotID string, direction int) ([]Shot, error)
	DeleteShot(ctx context.Context, owner, projectID, shotID string) error
	ListCharacters(ctx context.Context, owner, projectID, documentID string) ([]Character, error)
	ListScenes(ctx context.Context, owner, projectID, documentID string) ([]Scene, error)

	SaveAssetBatch(ctx context.Context, batch AssetBatch) error
	MergeAssets(ctx context.Context, owner, projectID string, assets []Asset) ([]Asset, error)
	ListAssets(ctx context.Context, owner, projectID string, assetType AssetType) ([]Asset, error)
	CreateAsset(ctx context.Context, asset Asset) (Asset, error)
	UpdateAsset(ctx context.Context, owner, projectID, assetID string, changes Asset, expectedRevision int64) (Asset, error)
	DeleteAsset(ctx context.Context, owner, projectID, assetID string) error

	// SaveStoryboard 按「项目 × 用户 × 剧集 × 引擎」幂等覆盖：同一组合重复生成会
	// 替换该集该引擎的全部分镜，v7 与 v3 各自独立、互不覆盖。
	SaveStoryboard(ctx context.Context, storyboard Storyboard) (Storyboard, error)
	GetStoryboard(ctx context.Context, owner, projectID string, episodeNo int, engine StoryboardEngine) (Storyboard, error)
	ListStoryboards(ctx context.Context, owner, projectID string, engine StoryboardEngine) ([]Storyboard, error)

	CreateBatch(ctx context.Context, batch *Batch) error
	GetBatch(ctx context.Context, owner, projectID string) (*Batch, error)
	UpdateBatch(ctx context.Context, batch *Batch, expectedRevision int64) error
	MutateBatch(ctx context.Context, owner, projectID string, fn func(*Batch) error) (*Batch, error)

	AppendBatchLog(ctx context.Context, batchID, owner, level, message string) error
	ListBatchLogs(ctx context.Context, batchID, owner string, limit int) ([]BatchLog, error)

	// AcquireLock 是跨批次资源的显式锁（同一步骤 / 同一集的并发执行）。
	AcquireLock(ctx context.Context, name, token string, ttl time.Duration) error
	ReleaseLock(ctx context.Context, name, token string) error
}
