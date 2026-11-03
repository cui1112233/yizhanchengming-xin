package giantmaterialexecutor

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrInvalidPlatform      = errors.New("invalid giant material executor platform")
	ErrInvalidInput         = errors.New("invalid giant material executor input")
	ErrPairingInvalid       = errors.New("giant material pairing code is invalid or expired")
	ErrExecutorUnauthorized = errors.New("giant material executor is unauthorized")
	ErrExecutorOffline      = errors.New("巨量素材执行器未启动或当前离线，请先启动一台执行器再试")
	ErrNoClaimableJob       = errors.New("no claimable giant material job")
	ErrJobNotFound          = errors.New("giant material job not found")
	ErrStaleLease           = errors.New("stale giant material job lease")
	ErrJobCancelled         = errors.New("giant material job cancelled")
	ErrInvalidJobState      = errors.New("invalid giant material job state")
	ErrJobConflict          = errors.New("giant material job conflict")
	ErrResultTooLarge       = errors.New("giant material result is too large")
	ErrPreferenceNotFound   = errors.New("giant material executor preference not found")
	ErrExecutorNotFound     = errors.New("giant material executor not found")
)

const (
	PlatformGiantMaterial    = "giant_material"
	PairingTTL               = 10 * time.Minute
	JobLeaseTTL              = 60 * time.Second
	OnlineThreshold          = 45 * time.Second
	FailureCooldown          = 5 * time.Minute
	StuckProgressLimit       = 10 * time.Minute
	HeartbeatIntervalSeconds = 15
	MaxResultBytes           = 2 << 20
	DefaultPreferredOS       = "windows"
	// pairingCleanupRetention 是过期配对记录的保留时长。过期配对码本身已不可用，
	// 超过保留期后从现有 giant_executor_pairings 表清理，不新增表。
	pairingCleanupRetention = time.Hour
)

type SecretHash [32]byte

type PairingRecord struct {
	ID            string
	OwnerUsername string
	Platform      string
	CodeHash      SecretHash
	ExpiresAt     time.Time
	ConsumedAt    *time.Time
	CreatedAt     time.Time
}

type PairingSecret struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// PendingPairing 是"已生成配对码但还没有设备消费"的只读视图，
// 只暴露指纹与时间，不回传明文配对码或哈希。
type PendingPairing struct {
	Platform    string    `json:"platform"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Expired     bool      `json:"expired"`
}

type PairInput struct {
	Code       string `json:"code"`
	DeviceName string `json:"deviceName"`
	Platform   string `json:"platform"`
	OS         string `json:"os"`
	Version    string `json:"version"`
}

type PairResult struct {
	ExecutorID               string `json:"executorId"`
	Token                    string `json:"token"`
	HeartbeatIntervalSeconds int    `json:"heartbeatIntervalSeconds"`
}

type HeartbeatInput struct {
	DeviceName string `json:"deviceName"`
	OS         string `json:"os"`
	Version    string `json:"version"`
}

type ExecutorRecord struct {
	ID            string
	OwnerUsername string
	Platform      string
	TokenHash     SecretHash
	DeviceName    string
	OS            string
	Version       string
	LastSeenAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ExecutorView struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Platform        string     `json:"platform"`
	OS              string     `json:"os"`
	Version         string     `json:"version"`
	Online          bool       `json:"online"`
	LastSeenAt      *time.Time `json:"lastSeenAt,omitempty"`
	RecentFailureAt *time.Time `json:"recentFailureAt,omitempty"`
}

type JobState string

const (
	JobQueued    JobState = "queued"
	JobLeased    JobState = "leased"
	JobRunning   JobState = "running"
	JobCleaning  JobState = "cleaning"
	JobUploading JobState = "uploading"
	JobSucceeded JobState = "succeeded"
	JobFailed    JobState = "failed"
	JobCancelled JobState = "cancelled"
)

type CreateJobInput struct {
	Platform          string     `json:"platform"`
	MaterialID        string     `json:"materialId"`
	PlatformBookID    string     `json:"platformBookId"`
	Title             string     `json:"title"`
	VideoURL          string     `json:"videoUrl"`
	VideoExpiresAt    *time.Time `json:"videoExpiresAt,omitempty"`
	DurationSeconds   float64    `json:"durationSeconds"`
	ModelVersion      string     `json:"modelVersion"`
	ContentRangeLines string     `json:"contentRangeLines,omitempty"`
}

type ResultInput struct {
	Text      string `json:"text"`
	WordCount int    `json:"wordCount"`
}

type ProgressInput struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
	Percent   int `json:"percent"`
}

type FailureInput struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ResultRecord struct {
	Text      string    `json:"text"`
	WordCount int       `json:"wordCount"`
	CreatedAt time.Time `json:"createdAt"`
}

type JobRecord struct {
	ID                string
	OwnerUsername     string
	Platform          string
	MaterialID        string
	PlatformBookID    string
	Title             string
	VideoURL          string
	VideoExpiresAt    *time.Time
	DurationSeconds   float64
	ModelVersion      string
	ContentRangeLines string
	TargetExecutorID  string
	State             JobState
	CancelRequested   bool
	Progress          ProgressInput
	LeaseExecutorID   string
	LeaseTokenHash    SecretHash
	LeaseGeneration   int64
	LeaseExpiresAt    *time.Time
	Result            ResultRecord
	ErrorCode         string
	ErrorMessage      string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type JobView struct {
	ID                string        `json:"id"`
	Platform          string        `json:"platform"`
	MaterialID        string        `json:"materialId"`
	PlatformBookID    string        `json:"platformBookId"`
	Title             string        `json:"title"`
	ModelVersion      string        `json:"modelVersion"`
	ContentRangeLines string        `json:"contentRangeLines,omitempty"`
	TargetExecutorID  string        `json:"targetExecutorId,omitempty"`
	State             JobState      `json:"state"`
	WaitingForExecutor bool          `json:"waitingForExecutor"`
	CancelRequested   bool          `json:"cancelRequested"`
	LeaseExecutorID   string        `json:"leaseExecutorId,omitempty"`
	LeaseGeneration   int64         `json:"leaseGeneration"`
	LeaseExpiresAt    *time.Time    `json:"leaseExpiresAt,omitempty"`
	Progress          ProgressInput `json:"progress"`
	Result            ResultRecord  `json:"result"`
	ErrorCode         string        `json:"errorCode,omitempty"`
	ErrorMessage      string        `json:"errorMessage,omitempty"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type ExecutorJobView struct {
	ID                string     `json:"id"`
	Platform          string     `json:"platform"`
	MaterialID        string     `json:"materialId"`
	PlatformBookID    string     `json:"platformBookId"`
	Title             string     `json:"title"`
	VideoURL          string     `json:"videoUrl"`
	VideoExpiresAt    *time.Time `json:"videoExpiresAt,omitempty"`
	DurationSeconds   float64    `json:"durationSeconds"`
	ModelVersion      string     `json:"modelVersion"`
	ContentRangeLines string     `json:"contentRangeLines,omitempty"`
	State             JobState   `json:"state"`
}

type LeaseCredential struct {
	Token      string `json:"leaseToken"`
	Generation int64  `json:"leaseGeneration"`
}

type ClaimResult struct {
	Job             ExecutorJobView `json:"job"`
	LeaseToken      string          `json:"leaseToken"`
	LeaseGeneration int64           `json:"leaseGeneration"`
	LeaseExpiresAt  time.Time       `json:"leaseExpiresAt"`
}

type LeaseView struct {
	LeaseGeneration int64     `json:"leaseGeneration"`
	LeaseExpiresAt  time.Time `json:"leaseExpiresAt"`
}

func (r JobRecord) MarshalPayload() ([]byte, error) {
	return json.Marshal(struct {
		MaterialID      string     `json:"materialId"`
		PlatformBookID  string     `json:"platformBookId"`
		Title           string     `json:"title"`
		VideoURL        string     `json:"videoUrl"`
		VideoExpiresAt  *time.Time `json:"videoExpiresAt,omitempty"`
		DurationSeconds float64    `json:"durationSeconds"`
		ModelVersion    string     `json:"modelVersion"`
	}{r.MaterialID, r.PlatformBookID, r.Title, r.VideoURL, r.VideoExpiresAt, r.DurationSeconds, r.ModelVersion})
}

type PreferenceRecord struct {
	OwnerUsername string
	PreferredOS   string
	UpdatedAt     time.Time
}

type PreferenceView struct {
	PreferredOS string    `json:"preferredOs"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}
