package batchfactoryexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// ObservabilityEvent is deliberately a fixed, payload-free log contract. It
// provides enough identifiers for tracing an execution without accepting
// prompts, source text, credentials, cookies, or provider response bodies.
type ObservabilityEvent struct {
	EventType        string `json:"event_type"`
	ExecutionID      string `json:"execution_id"`
	OwnerHash        string `json:"owner_hash"`
	BatchID          string `json:"batch_id"`
	BookID           string `json:"book_id"`
	StageKey         string `json:"stage_key"`
	AttemptNo        uint32 `json:"attempt_no"`
	InputFingerprint string `json:"input_fingerprint"`
	LegacyJobID      string `json:"legacy_job_id"`
	Status           string `json:"status"`
	DurationMS       int64  `json:"duration_ms"`
}

func NewObservabilityEvent(eventType, owner string) ObservabilityEvent {
	sum := sha256.Sum256([]byte(owner))
	return ObservabilityEvent{
		EventType: eventType,
		OwnerHash: hex.EncodeToString(sum[:8]),
	}
}

func (e ObservabilityEvent) JSON() string {
	encoded, err := json.Marshal(e)
	if err != nil {
		return `{"event_type":"observability_encode_failed"}`
	}
	return string(encoded)
}

// CanCompleteBook protects the strongest target from local-artifact false
// positives. A publication is complete only after a remote identifier and a
// confirmed platform readback have both been persisted.
func CanCompleteBook(target string, receipt *PublicationReceipt) bool {
	if target != "published_confirmed" {
		return true
	}
	return receipt != nil && receipt.RemoteWorkID != "" && receipt.ConfirmedAt != nil && receipt.RemoteStatus == "published"
}
