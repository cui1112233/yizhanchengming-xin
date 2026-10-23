package agentworkspace

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrBatchFactoryBindingInvalid    = errors.New("invalid batch factory binding")
	ErrBatchFactoryBindingNotFound   = errors.New("batch factory binding not found")
	ErrBatchFactoryBindingConflict   = errors.New("batch factory binding conflict")
	ErrBatchFactoryNodeOriginInvalid = errors.New("invalid batch factory node origin")
)

type BatchFactoryBinding struct {
	Owner     string
	BatchID   string
	ProjectID string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BatchFactoryNodeOrigin struct {
	ProjectID    string
	NodeID       string
	BatchID      string
	BookID       string
	ResourceType string
	ResourceID   string
	VersionID    string
}

type BatchFactoryBindingStore interface {
	BindBatchFactoryProject(context.Context, BatchFactoryBinding) (BatchFactoryBinding, error)
	RebindBatchFactoryProject(context.Context, string, string, string, string, time.Time) (BatchFactoryBinding, error)
	BatchFactoryBinding(context.Context, string, string) (BatchFactoryBinding, error)
	BindBatchFactoryNode(context.Context, string, BatchFactoryNodeOrigin) error
	BatchFactoryNodeOrigins(context.Context, string, string, string) ([]BatchFactoryNodeOrigin, error)
	RemoveBatchFactoryNodeBinding(context.Context, string, string, string) error
}

func ValidateBatchFactoryBinding(binding BatchFactoryBinding) error {
	if strings.TrimSpace(binding.Owner) == "" || strings.TrimSpace(binding.BatchID) == "" || strings.TrimSpace(binding.ProjectID) == "" {
		return ErrBatchFactoryBindingInvalid
	}
	return nil
}

func ValidateBatchFactoryNodeOrigin(origin BatchFactoryNodeOrigin) error {
	if strings.TrimSpace(origin.BatchID) == "" || strings.TrimSpace(origin.BookID) == "" || strings.TrimSpace(origin.ResourceID) == "" {
		return ErrBatchFactoryNodeOriginInvalid
	}
	switch strings.TrimSpace(origin.ResourceType) {
	case "book_script", "book_asset", "audio", "video", "merge":
		return nil
	default:
		return ErrBatchFactoryNodeOriginInvalid
	}
}

func normalizeBatchFactoryBinding(binding BatchFactoryBinding, now time.Time) (BatchFactoryBinding, error) {
	binding.Owner = strings.TrimSpace(binding.Owner)
	binding.BatchID = strings.TrimSpace(binding.BatchID)
	binding.ProjectID = strings.TrimSpace(binding.ProjectID)
	if err := ValidateBatchFactoryBinding(binding); err != nil {
		return BatchFactoryBinding{}, err
	}
	if binding.CreatedAt.IsZero() {
		binding.CreatedAt = now
	}
	if binding.UpdatedAt.IsZero() {
		binding.UpdatedAt = binding.CreatedAt
	}
	binding.CreatedAt = binding.CreatedAt.UTC().Truncate(time.Microsecond)
	binding.UpdatedAt = binding.UpdatedAt.UTC().Truncate(time.Microsecond)
	return binding, nil
}

func normalizeBatchFactoryNodeOrigin(owner string, origin BatchFactoryNodeOrigin) (BatchFactoryNodeOrigin, error) {
	owner = strings.TrimSpace(owner)
	origin.ProjectID = strings.TrimSpace(origin.ProjectID)
	origin.NodeID = strings.TrimSpace(origin.NodeID)
	origin.BatchID = strings.TrimSpace(origin.BatchID)
	origin.BookID = strings.TrimSpace(origin.BookID)
	origin.ResourceType = strings.TrimSpace(origin.ResourceType)
	origin.ResourceID = strings.TrimSpace(origin.ResourceID)
	origin.VersionID = strings.TrimSpace(origin.VersionID)
	if owner == "" || origin.ProjectID == "" || origin.NodeID == "" {
		return BatchFactoryNodeOrigin{}, ErrBatchFactoryNodeOriginInvalid
	}
	if err := ValidateBatchFactoryNodeOrigin(origin); err != nil {
		return BatchFactoryNodeOrigin{}, err
	}
	return origin, nil
}
