package localexecutor

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

var (
	ErrArtifactNotFound = errors.New("local executor artifact not found")
	ErrArtifactConflict = errors.New("local executor artifact conflict")
)

type ArtifactInput struct {
	ID         string
	MediaType  string
	ByteSize   int64
	SHA256     string
	StorageRef string
}

type ArtifactRecord struct {
	ID            string
	JobID         string
	OwnerUsername string
	MediaType     string
	ByteSize      int64
	SHA256        string
	StorageRef    string
	CreatedAt     time.Time
}

type ArtifactView struct {
	ID            string    `json:"artifactId"`
	JobID         string    `json:"jobId"`
	OwnerUsername string    `json:"ownerUsername,omitempty"`
	MediaType     string    `json:"mediaType"`
	ByteSize      int64     `json:"byteSize"`
	SHA256        string    `json:"sha256"`
	StorageRef    string    `json:"-"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ArtifactStore interface {
	CreateArtifact(context.Context, string, string, SecretHash, int64, ArtifactRecord, time.Time) (ArtifactRecord, error)
	ArtifactForOwner(context.Context, string, string) (ArtifactRecord, error)
	ArtifactForJob(context.Context, string) (ArtifactRecord, error)
	ArtifactsForJob(context.Context, string) ([]ArtifactRecord, error)
}

func NewArtifactID() (string, error) { return randomID("lea_", 12) }

func artifactView(record ArtifactRecord) ArtifactView {
	return ArtifactView{
		ID: record.ID, JobID: record.JobID, OwnerUsername: record.OwnerUsername,
		MediaType: record.MediaType, ByteSize: record.ByteSize, SHA256: record.SHA256,
		StorageRef: record.StorageRef, CreatedAt: record.CreatedAt,
	}
}

func (s *Service) RecordArtifact(ctx context.Context, executorToken, jobID string, lease LeaseCredential, input ArtifactInput) (ArtifactView, error) {
	artifacts, ok := s.store.(ArtifactStore)
	if !ok {
		return ArtifactView{}, ErrInvalidInput
	}
	executor, err := s.executorForToken(ctx, executorToken)
	if err != nil {
		return ArtifactView{}, err
	}
	jobID = strings.TrimSpace(jobID)
	input.ID = strings.TrimSpace(input.ID)
	input.MediaType = strings.ToLower(strings.TrimSpace(input.MediaType))
	input.SHA256 = strings.ToLower(strings.TrimSpace(input.SHA256))
	input.StorageRef = strings.TrimSpace(input.StorageRef)
	if jobID == "" || strings.TrimSpace(lease.Token) == "" || lease.Generation < 1 || !validArtifactInput(input) {
		return ArtifactView{}, ErrInvalidInput
	}
	record := ArtifactRecord{
		ID: input.ID, MediaType: input.MediaType, ByteSize: input.ByteSize,
		SHA256: input.SHA256, StorageRef: input.StorageRef,
	}
	created, err := artifacts.CreateArtifact(ctx, executor.ID, jobID, hashSecret(lease.Token), lease.Generation, record, s.now().UTC())
	if err != nil {
		return ArtifactView{}, err
	}
	return artifactView(created), nil
}

func (s *Service) GetArtifact(ctx context.Context, owner, id string) (ArtifactView, error) {
	artifacts, ok := s.store.(ArtifactStore)
	if !ok {
		return ArtifactView{}, ErrInvalidInput
	}
	owner = strings.TrimSpace(owner)
	id = strings.TrimSpace(id)
	if owner == "" || id == "" {
		return ArtifactView{}, ErrInvalidInput
	}
	record, err := artifacts.ArtifactForOwner(ctx, owner, id)
	if err != nil {
		return ArtifactView{}, err
	}
	return artifactView(record), nil
}

func (s *Service) GetJobArtifacts(ctx context.Context, owner, jobID string) ([]ArtifactView, error) {
	artifacts, ok := s.store.(ArtifactStore)
	if !ok {
		return nil, ErrInvalidInput
	}
	job, err := s.GetJob(ctx, owner, jobID)
	if err != nil {
		return nil, err
	}
	records, err := artifacts.ArtifactsForJob(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	views := make([]ArtifactView, 0, len(records))
	for _, record := range records {
		views = append(views, artifactView(record))
	}
	return views, nil
}

func (s *Service) validateResultArtifact(ctx context.Context, jobID, artifactID string) error {
	artifacts, ok := s.store.(ArtifactStore)
	if !ok {
		return ErrInvalidInput
	}
	records, err := artifacts.ArtifactsForJob(ctx, strings.TrimSpace(jobID))
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.ID == strings.TrimSpace(artifactID) {
			return nil
		}
	}
	if len(records) > 0 {
		return ErrArtifactConflict
	}
	return ErrArtifactNotFound
}

func validArtifactInput(input ArtifactInput) bool {
	if input.ID == "" || len(input.ID) > 96 || !validArtifactMediaType(input.MediaType) || input.ByteSize <= 0 || input.StorageRef == "" || len(input.StorageRef) > 512 {
		return false
	}
	if len(input.SHA256) != 64 {
		return false
	}
	decoded, err := hex.DecodeString(input.SHA256)
	return err == nil && len(decoded) == 32
}

func validArtifactMediaType(mediaType string) bool {
	switch mediaType {
	case "video/mp4", "image/png", "image/jpeg", "image/webp":
		return true
	default:
		return false
	}
}
