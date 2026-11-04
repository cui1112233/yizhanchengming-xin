package mergeworker

import "time"

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
)

type Source struct {
	ProductionJobID string `json:"productionJobId"`
	VideoID         string `json:"videoId"`
	MediaURL        string `json:"mediaUrl"`
	Order           int    `json:"order"`
}

type SubmitRequest struct {
	BatchID    string   `json:"batchId"`
	Sources    []Source `json:"sources"`
	TimingMode string   `json:"timingMode,omitempty"`
	Speed      float64  `json:"speed,omitempty"`
	TTSSpeed   float64  `json:"ttsSpeed,omitempty"`
}

type Job struct {
	ID           string    `json:"taskId"`
	BatchID      string    `json:"batchId"`
	Status       State     `json:"status"`
	Sources      []Source  `json:"sources,omitempty"`
	TimingMode   string    `json:"timingMode,omitempty"`
	Speed        float64   `json:"speed,omitempty"`
	TTSSpeed     float64   `json:"ttsSpeed,omitempty"`
	OutputURL    string    `json:"outputUrl,omitempty"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
}
