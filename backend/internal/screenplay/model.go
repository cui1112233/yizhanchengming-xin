package screenplay

import (
	"errors"
	"time"
)

// 批次状态口径见规格 §13.3。原版 JSON 使用 awaiting_user；本实现按项目任务口径
// 统一为 waiting_user，并在 NormalizeBatchStatus 中兼容旧别名。
type BatchStatus string

const (
	BatchRunning     BatchStatus = "running"
	BatchWaitingUser BatchStatus = "waiting_user"
	BatchPaused      BatchStatus = "paused"
	BatchCompleted   BatchStatus = "completed"
	BatchFailed      BatchStatus = "failed"
)

// 步态集合见规格 §13.3。
type StepStatus string

const (
	StepPending    StepStatus = "pending"
	StepInProgress StepStatus = "in_progress"
	StepDone       StepStatus = "done"
	StepStale      StepStatus = "stale"
	StepFailed     StepStatus = "failed"
	StepSkipped    StepStatus = "skipped"
	StepBlocked    StepStatus = "blocked"
)

// 权限两档（规格 §15 / 任务口径）：bypass 全自动；auto 只列计划、需人工确认。
type PermissionMode string

const (
	PermissionBypass PermissionMode = "bypass"
	PermissionAuto   PermissionMode = "auto"
)

// 资产提取档位（规格 §1.3）。
type AssetDegree string

const (
	DegreeStandard   AssetDegree = "standard"
	DegreeMoreDiff   AssetDegree = "more_diff"
	DegreeMaster     AssetDegree = "master"
	DegreeMasterReal AssetDegree = "master_real"
)

// 资产类型。
type AssetType string

const (
	AssetCharacter AssetType = "character"
	AssetScene     AssetType = "scene"
	AssetProp      AssetType = "prop"
)

// 分镜引擎。两套结构不同，各自独立落库（规格 §6 / §7）。
type StoryboardEngine string

const (
	EngineXiaoluoV9    StoryboardEngine = "xiaoluo_v9_30s"
	EngineXiaoluoV7    StoryboardEngine = "xiaoluo_v7_15s"
	EngineXiaoyunqueV3 StoryboardEngine = "xiaoyunque_v3"
)

var (
	ErrNotFound         = errors.New("记录不存在")
	ErrRevisionConflict = errors.New("revision 冲突：数据已被其他请求修改")
	ErrLockBusy         = errors.New("资源正被其他任务占用")
	ErrInvalidState     = errors.New("当前状态不允许该操作")
	ErrInvalidArgument  = errors.New("参数无效")
)

// Project 是一个漫剧项目（对应洛水 workspace 级配置）。
type Project struct {
	CreativeConfig *CreativeConfig `json:"creative_config,omitempty"`
	ID             string          `json:"id"`
	OwnerUsername  string          `json:"owner_username"`
	Title          string          `json:"title"`
	ScriptMode     string          `json:"script_mode"` // a1 | b
	AspectRatio    string          `json:"aspect_ratio"`
	ArtStyle       string          `json:"art_style"`
	StyleDesc      string          `json:"-"`
	SelectedStyle  string          `json:"selected_style"`
	Synopsis       string          `json:"synopsis"`
	SourceText     string          `json:"source_text"`
	ModelID        string          `json:"model_id"`
	PromptID       string          `json:"prompt_id"`
	AssetDegree    AssetDegree     `json:"asset_degree"`
	PermissionMode PermissionMode  `json:"permission_mode"`
	Revision       int64           `json:"revision"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// Document 是剧本正文的不可变版本。A1 / B 档都追加新版本，旧版本置 superseded，
// 从而修复原版无锁覆盖同一文件的缺陷。
type Document struct {
	RawResult     string    `json:"raw_result,omitempty"`
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	OwnerUsername string    `json:"owner_username"`
	Source        string    `json:"source"` // a1 | b
	Version       int       `json:"version"`
	Status        string    `json:"status"` // draft | active | superseded
	Content       string    `json:"content"`
	ContentSHA256 string    `json:"content_sha256"`
	ChapterCount  int       `json:"chapter_count"`
	EpisodeCount  int       `json:"episode_count"`
	SegmentCount  int       `json:"segment_count"`
	Revision      int64     `json:"revision"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Shot 是剧本分场。
type Shot struct {
	ID                string   `json:"id"`
	DocumentID        string   `json:"document_id"`
	ProjectID         string   `json:"project_id"`
	OwnerUsername     string   `json:"owner_username"`
	EpisodeNo         int      `json:"episode_no"`
	SortOrder         int      `json:"sort_order"`
	Heading           string   `json:"heading"`
	ScriptText        string   `json:"script_text"`
	CharacterNames    []string `json:"characters"`
	SceneName         string   `json:"scene"`
	PropNames         []string `json:"props"`
	LocationReference string   `json:"location_reference"`
}

// Character / Scene / Prop 是剧本页的结构化资产。
type Character struct {
	ID             string `json:"id"`
	DocumentID     string `json:"document_id"`
	ProjectID      string `json:"project_id"`
	OwnerUsername  string `json:"owner_username"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ImagePrompt    string `json:"image_prompt"`
	ReferenceImage string `json:"reference_image"`
	SortOrder      int    `json:"sort_order"`
	Voice          *Voice `json:"voice,omitempty"`
}

type Scene struct {
	ID             string `json:"id"`
	DocumentID     string `json:"document_id"`
	ProjectID      string `json:"project_id"`
	OwnerUsername  string `json:"owner_username"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ImagePrompt    string `json:"image_prompt"`
	ReferenceImage string `json:"reference_image"`
	SortOrder      int    `json:"sort_order"`
}

type Prop struct {
	ID             string `json:"id"`
	DocumentID     string `json:"document_id"`
	ProjectID      string `json:"project_id"`
	OwnerUsername  string `json:"owner_username"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ImagePrompt    string `json:"image_prompt"`
	ReferenceImage string `json:"reference_image"`
	SortOrder      int    `json:"sort_order"`
}

// Voice 是角色音色（规格 §5：voice_name / voice_id / voice_settings）。
type Voice struct {
	ID            string         `json:"id"`
	DocumentID    string         `json:"document_id"`
	ProjectID     string         `json:"project_id"`
	OwnerUsername string         `json:"owner_username"`
	CharacterID   string         `json:"character_id"`
	VoiceName     string         `json:"voice_name"`
	VoiceID       string         `json:"voice_id"`
	VoiceSettings map[string]any `json:"voice_settings,omitempty"`
}

// Asset 是四档提取合并后的资产。同名资产按名称合并，保留最新档位。
type Asset struct {
	ID            string            `json:"id"`
	ProjectID     string            `json:"project_id"`
	OwnerUsername string            `json:"owner_username"`
	Type          AssetType         `json:"type"`
	Degree        AssetDegree       `json:"degree"`
	Name          string            `json:"name"`
	Identity      string            `json:"identity"`
	Description   string            `json:"description"`
	Prompt        string            `json:"prompt"`
	Voice         string            `json:"voice,omitempty"`
	Episodes      string            `json:"episodes,omitempty"`
	Tag           string            `json:"tag,omitempty"`
	Extra         map[string]string `json:"extra,omitempty"`
	SortOrder     int               `json:"sort_order"`
	SourceBatch   int               `json:"source_batch"`
	Revision      int64             `json:"revision"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// AssetBatch 记录 5 章一批的提取任务，20 路并发时各自独立落库，最后合并。
type AssetBatch struct {
	ID            string      `json:"id"`
	ProjectID     string      `json:"project_id"`
	OwnerUsername string      `json:"owner_username"`
	Type          AssetType   `json:"type"`
	Degree        AssetDegree `json:"degree"`
	BatchIndex    int         `json:"batch_index"`
	ChapterStart  int         `json:"chapter_start"`
	ChapterEnd    int         `json:"chapter_end"`
	Status        string      `json:"status"`
	ItemCount     int         `json:"item_count"`
	ErrorMessage  string      `json:"error_message"`
	Revision      int64       `json:"revision"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// Storyboard 是某一集、某一引擎的分镜词。
type Storyboard struct {
	ID            string           `json:"id"`
	ProjectID     string           `json:"project_id"`
	OwnerUsername string           `json:"owner_username"`
	EpisodeNo     int              `json:"episode_no"`
	Engine        StoryboardEngine `json:"engine"`
	Status        string           `json:"status"`
	SourceSHA256  string           `json:"source_sha256"`
	ShotCount     int              `json:"shot_count"`
	Revision      int64            `json:"revision"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	Shots         []StoryboardShot `json:"shots"`
}

// StoryboardShot 是分镜六字段契约。JSON 的 id 字段按 1 起序号写入 Seq。
type StoryboardShot struct {
	ID            string  `json:"id"`
	StoryboardID  string  `json:"storyboard_id"`
	ProjectID     string  `json:"project_id"`
	OwnerUsername string  `json:"owner_username"`
	Seq           int     `json:"seq"`
	SceneID       int     `json:"scene_id"`
	Duration      float64 `json:"duration_sec"`
	Characters    string  `json:"characters"`
	Scene         string  `json:"scene"`
	Props         string  `json:"props"`
	VideoDesc     string  `json:"video_desc"`
}

// ReviewIssue 对应医生输出的问题条目。
type ReviewIssue struct {
	Level      string `json:"level"` // error | warning | info
	Category   string `json:"category"`
	Target     string `json:"target"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

// Review 是三个医生的统一结果。闸门只看 Pass 与 BlockNextStep 两个布尔值。
type Review struct {
	ID             string        `json:"id"`
	ProjectID      string        `json:"project_id"`
	OwnerUsername  string        `json:"owner_username"`
	StepKey        string        `json:"step_key"`
	TargetKey      string        `json:"target_key"`
	Mode           string        `json:"mode"` // review | rewrite
	Pass           bool          `json:"pass"`
	BlockNextStep  bool          `json:"block_next_step"`
	Summary        string        `json:"summary"`
	Issues         []ReviewIssue `json:"issues"`
	ReportMarkdown string        `json:"report_markdown,omitempty"`
	PromptID       string        `json:"prompt_id"`
	Revision       int64         `json:"revision"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// PendingAction 是 auto 权限下的待确认动作。bypass 不产生待确认动作。
type PendingAction struct {
	ID              string    `json:"id"`
	Kind            string    `json:"kind"` // write | skill | external
	StepKey         string    `json:"step_key"`
	Title           string    `json:"title"`
	Detail          string    `json:"detail"`
	RequiresConfirm bool      `json:"requires_confirm"`
	CreatedAt       time.Time `json:"created_at"`
}

// EpisodeState 记录按集步骤中每一集的状态。
type EpisodeState struct {
	EpisodeNo int        `json:"episode_no"`
	Status    StepStatus `json:"status"`
	Artifact  string     `json:"artifact"`
	Reason    string     `json:"reason"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// StepState 是单个批次步的状态。
type StepState struct {
	Key           string                `json:"key"`
	Status        StepStatus            `json:"status"`
	TotalEpisodes int                   `json:"total_episodes"`
	StartedAt     *time.Time            `json:"started_at,omitempty"`
	FinishedAt    *time.Time            `json:"finished_at,omitempty"`
	Artifacts     []string              `json:"artifacts,omitempty"`
	Error         string                `json:"error,omitempty"`
	Review        *Review               `json:"review,omitempty"`
	Episodes      map[int]*EpisodeState `json:"episodes,omitempty"`
}

// Batch 是批次状态机的持久化形态（对应原版 state.json，但走 MySQL + revision）。
type Batch struct {
	ID             string                `json:"id"`
	ProjectID      string                `json:"project_id"`
	OwnerUsername  string                `json:"owner_username"`
	SessionID      string                `json:"session_id"`
	Status         BatchStatus           `json:"status"`
	CurrentStep    string                `json:"current_step"`
	CurrentEpisode int                   `json:"current_episode"`
	PlanVersion    int                   `json:"plan_version"`
	PermissionMode PermissionMode        `json:"permission_mode"`
	PendingActions []PendingAction       `json:"pending_actions"`
	LastError      string                `json:"last_error"`
	Revision       int64                 `json:"revision"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	LastRunAt      *time.Time            `json:"last_run_at,omitempty"`
	Steps          map[string]*StepState `json:"steps"`
}

// Step 返回某一步的状态，不存在时返回 nil。
func (b *Batch) Step(key string) *StepState {
	if b == nil || b.Steps == nil {
		return nil
	}
	return b.Steps[key]
}

// NormalizeBatchStatus 兼容原版 awaiting_user 别名。
func NormalizeBatchStatus(status BatchStatus) BatchStatus {
	if status == "awaiting_user" {
		return BatchWaitingUser
	}
	return status
}

// ParsePermissionMode 解析权限模式；空值按交互批次默认 auto。
func ParsePermissionMode(raw string) PermissionMode {
	if PermissionMode(raw) == PermissionBypass {
		return PermissionBypass
	}
	return PermissionAuto
}
