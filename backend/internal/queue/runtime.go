package queue

import (
	"context"
	"net/url"
	"strings"
)

// RuntimeQueueStatus describes only the short-lived Redis runtime dependency.
// Business records remain persisted in MySQL.
type RuntimeQueueStatus struct {
	Configured bool
	Reachable  bool
	Reason     string
}

type Pinger interface {
	Ping(context.Context) error
}

func RuntimeQueueStatusFor(rawURL string, pinger Pinger) RuntimeQueueStatus {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return RuntimeQueueStatus{Reason: "redis not configured"}
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "redis" || parsed.Host == "" {
		return RuntimeQueueStatus{Configured: true, Reason: "invalid redis address"}
	}
	if pinger == nil {
		return RuntimeQueueStatus{Configured: true, Reason: "redis unavailable"}
	}
	if err := pinger.Ping(context.Background()); err != nil {
		return RuntimeQueueStatus{Configured: true, Reason: "redis unavailable"}
	}
	return RuntimeQueueStatus{Configured: true, Reachable: true}
}
