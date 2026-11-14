package screenplay

import (
	"context"
	"encoding/json"
	"time"
)

// Discovery never writes the live asset library. Only ApplyAssetCheckJob may
// merge a reviewed draft, and existing exact-name entries always win.
type AssetCheckCandidate struct {
	ID       string          `json:"id"`
	Asset    Asset           `json:"asset"`
	Original json.RawMessage `json:"original"`
}

type AssetCheckEpisode struct {
	ChapterIndex int    `json:"chapter_index"`
	EpisodeNo    int    `json:"episode_no"`
	ChapterID    string `json:"chapter_id,omitempty"`
	Title        string `json:"title"`
	Content      string `json:"content"`
}

type AssetCheckStep struct {
	ID            string                `json:"id"`
	ChapterIndex  int                   `json:"chapter_index"`
	EpisodeNo     int                   `json:"episode_no"`
	Type          AssetType             `json:"type"`
	Status        string                `json:"status"`
	Error         string                `json:"error,omitempty"`
	Request       TextRequest           `json:"request"`
	RequestDigest string                `json:"request_digest"`
	RawResponse   string                `json:"raw_response,omitempty"`
	Candidates    []AssetCheckCandidate `json:"candidates,omitempty"`
	StartedAt     *time.Time            `json:"started_at,omitempty"`
	FinishedAt    *time.Time            `json:"finished_at,omitempty"`
}

type AssetCheckJob struct {
	ID                     string                `json:"id"`
	Owner                  string                `json:"owner"`
	ProjectID              string                `json:"project_id"`
	Actor                  string                `json:"actor"`
	RequestKey             string                `json:"request_key"`
	InputDigest            string                `json:"input_digest"`
	SourceMode             string                `json:"source_mode"`
	SelectedEpisodeNumbers []int                 `json:"selected_episode_numbers,omitempty"`
	Model                  string                `json:"model"`
	Degree                 AssetDegree           `json:"degree"`
	Status                 string                `json:"status"`
	Revision               int64                 `json:"revision"`
	CancelRequested        bool                  `json:"cancel_requested"`
	LeaseToken             string                `json:"lease_token,omitempty"`
	Episodes               []AssetCheckEpisode   `json:"episodes"`
	EpisodeCompletionOrder []int                 `json:"episode_completion_order,omitempty"`
	Steps                  []AssetCheckStep      `json:"steps"`
	Existing               []Asset               `json:"existing"`
	Candidates             []AssetCheckCandidate `json:"candidates"`
	Error                  string                `json:"error,omitempty"`
	Applied                bool                  `json:"applied"`
	Discarded              bool                  `json:"discarded"`
	AddedCount             int                   `json:"added_count"`
	SkippedCount           int                   `json:"skipped_count"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
}

type AssetCheckRepository interface {
	FindAssetCheckJobByRequestKey(context.Context, string, string, string) (AssetCheckJob, error)
	CreateAssetCheckJob(context.Context, AssetCheckJob) (AssetCheckJob, error)
	GetAssetCheckJob(context.Context, string, string, string) (AssetCheckJob, error)
	ListAssetCheckJobs(context.Context, string, string, int) ([]AssetCheckJob, error)
	ListRunnableAssetCheckJobs(context.Context) ([]AssetCheckJob, error)
	MutateAssetCheckJob(context.Context, string, string, string, func(*AssetCheckJob) error) (AssetCheckJob, error)
	ApplyAssetCheckJob(context.Context, string, string, string, int64) (AssetCheckJob, error)
}

func cloneAssetCheckJob(job AssetCheckJob) AssetCheckJob {
	raw, _ := json.Marshal(job)
	var out AssetCheckJob
	_ = json.Unmarshal(raw, &out)
	return out
}
