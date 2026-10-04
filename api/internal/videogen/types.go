package videogen

import "context"

const (
	StateQueued    = "queued"
	StateRunning   = "running"
	StateSucceeded = "succeeded"
	StateFailed    = "failed"
)

type Request struct {
	Model         string
	Prompt        string
	Duration      int
	AspectRatio   string
	Resolution    string
	ReferenceURLs []string
}

type Task struct {
	ID         string
	State      string
	MediaURL   string
	DurationMs uint64
}

type Provider interface {
	Submit(context.Context, Request) (Task, error)
	Poll(context.Context, Task) (Task, error)
}
