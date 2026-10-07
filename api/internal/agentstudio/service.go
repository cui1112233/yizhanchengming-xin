package agentstudio

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Service struct {
	store    Store
	executor Executor
	now      func() time.Time
	objects  ObjectStore
	bucket   string
}

func (s *Service) SetObjects(objects ObjectStore, bucket string) {
	s.objects = objects
	s.bucket = strings.TrimSpace(bucket)
}
func (s *Service) UploadAttachment(ctx context.Context, a Actor, projectID int64, filename, contentType string, body io.Reader) (Attachment, error) {
	if s == nil || s.store == nil || a.UserID <= 0 || projectID <= 0 || body == nil || strings.TrimSpace(filename) == "" {
		return Attachment{}, ErrInvalid
	}
	if s.objects == nil || s.bucket == "" {
		return Attachment{}, ErrStorageUnavailable
	}
	if _, e := s.store.GetProject(ctx, a, projectID); e != nil {
		return Attachment{}, e
	}
	name := filepath.Base(strings.TrimSpace(filename))
	if name == "." || name == string(filepath.Separator) {
		return Attachment{}, ErrInvalid
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	f, e := os.CreateTemp("", "agent-upload-*")
	if e != nil {
		return Attachment{}, e
	}
	path := f.Name()
	defer os.Remove(path)
	h := sha256.New()
	n, e := io.Copy(io.MultiWriter(f, h), io.LimitReader(body, 20<<20+1))
	closeErr := f.Close()
	if e != nil || closeErr != nil || n < 1 || n > 20<<20 {
		return Attachment{}, ErrInvalid
	}
	raw := make([]byte, 16)
	if _, e = rand.Read(raw); e != nil {
		return Attachment{}, e
	}
	key := fmt.Sprintf("agent/project-%d/%s", projectID, hex.EncodeToString(raw))
	if e = s.objects.PutObjectFromFile(ctx, s.bucket, key, path); e != nil {
		return Attachment{}, e
	}
	v, e := s.store.CreateAttachment(ctx, a, Attachment{ProjectID: projectID, OwnerUserID: a.UserID, Bucket: s.bucket, ObjectKey: key, Filename: name, ContentType: contentType, ByteSize: n, SHA256: hex.EncodeToString(h.Sum(nil))})
	if e != nil {
		_ = s.objects.DeleteObject(ctx, s.bucket, key)
		return Attachment{}, e
	}
	return v, nil
}
func (s *Service) OpenAttachment(ctx context.Context, a Actor, p, id int64) (Attachment, io.ReadCloser, error) {
	if s == nil || s.store == nil || s.objects == nil || a.UserID <= 0 || p <= 0 || id <= 0 {
		return Attachment{}, nil, ErrStorageUnavailable
	}
	v, e := s.store.GetAttachment(ctx, a, p, id)
	if e != nil {
		return Attachment{}, nil, e
	}
	r, e := s.objects.GetObject(ctx, v.Bucket, v.ObjectKey)
	return v, r, e
}
func (s *Service) DeleteAttachment(ctx context.Context, a Actor, p, id int64) error {
	if s == nil || s.store == nil || s.objects == nil || a.UserID <= 0 || p <= 0 || id <= 0 {
		return ErrStorageUnavailable
	}
	v, e := s.store.GetAttachment(ctx, a, p, id)
	if e != nil {
		return e
	}
	if e = s.objects.DeleteObject(ctx, v.Bucket, v.ObjectKey); e != nil {
		return e
	}
	return s.store.DeleteAttachment(ctx, a, p, id)
}
func (s *Service) ListAttachments(ctx context.Context, a Actor, p int64) ([]Attachment, error) {
	if s == nil || s.store == nil || a.UserID <= 0 || p <= 0 {
		return nil, ErrInvalid
	}
	return s.store.ListAttachments(ctx, a, p)
}

func NewService(store Store, executor Executor) *Service {
	return &Service{store: store, executor: executor, now: time.Now}
}

func (s *Service) GetCanvas(ctx context.Context, actor Actor, projectID int64) (Canvas, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 || projectID <= 0 {
		return Canvas{}, ErrInvalid
	}
	return s.store.GetCanvas(ctx, actor, projectID)
}

func (s *Service) SaveCanvas(ctx context.Context, actor Actor, canvas Canvas) (Canvas, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 || canvas.ProjectID <= 0 || canvas.Revision < 0 || !json.Valid(canvas.Document) {
		return Canvas{}, ErrInvalid
	}
	return s.store.SaveCanvas(ctx, actor, canvas)
}

func (s *Service) ListCanvasVersions(ctx context.Context, actor Actor, projectID int64) ([]CanvasVersion, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 || projectID <= 0 {
		return nil, ErrInvalid
	}
	return s.store.ListCanvasVersions(ctx, actor, projectID)
}

func (s *Service) RestoreCanvas(ctx context.Context, actor Actor, projectID int64, revision int) (Canvas, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 || projectID <= 0 || revision <= 0 {
		return Canvas{}, ErrInvalid
	}
	return s.store.RestoreCanvas(ctx, actor, projectID, revision)
}

func (s *Service) ListProjects(ctx context.Context, actor Actor) ([]Project, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 {
		return nil, ErrInvalid
	}
	return s.store.ListProjects(ctx, actor)
}
func (s *Service) DeleteProject(ctx context.Context, actor Actor, id int64) error {
	if s == nil || s.store == nil || actor.UserID <= 0 || id <= 0 {
		return ErrInvalid
	}
	return s.store.DeleteProject(ctx, actor, id)
}
func (s *Service) ListMessages(ctx context.Context, actor Actor, id int64) ([]Message, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 || id <= 0 {
		return nil, ErrInvalid
	}
	return s.store.ListMessages(ctx, actor, id)
}
func (s *Service) ListExecutions(ctx context.Context, actor Actor, id int64) ([]Execution, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 || id <= 0 {
		return nil, ErrInvalid
	}
	return s.store.ListExecutions(ctx, actor, id)
}
func (s *Service) CreateSkill(ctx context.Context, actor Actor, in CreateSkillInput) (Skill, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 {
		return Skill{}, ErrInvalid
	}
	return s.store.CreateSkill(ctx, actor, in)
}
func (s *Service) ListSkills(ctx context.Context, actor Actor) ([]Skill, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 {
		return nil, ErrInvalid
	}
	return s.store.ListSkills(ctx, actor)
}

func (s *Service) CreateProject(ctx context.Context, actor Actor, input CreateProjectInput) (Project, error) {
	if s == nil || s.store == nil || actor.UserID <= 0 || strings.TrimSpace(input.Title) == "" {
		return Project{}, ErrInvalid
	}
	return s.store.CreateProject(ctx, actor, input)
}
func (s *Service) Continue(ctx context.Context, actor Actor, projectID int64, input ContinueInput) (ContinueResult, error) {
	if s == nil || s.store == nil || s.executor == nil || actor.UserID <= 0 || projectID <= 0 || strings.TrimSpace(input.Content) == "" {
		return ContinueResult{}, ErrInvalid
	}
	project, err := s.store.GetProject(ctx, actor, projectID)
	if err != nil {
		return ContinueResult{}, err
	}
	if _, err = s.store.GetSkills(ctx, actor, input.SkillIDs); err != nil {
		return ContinueResult{}, err
	}
	now := s.now().UTC()
	user, err := s.store.CreateMessage(ctx, Message{ProjectID: project.ID, Role: RoleUser, Content: strings.TrimSpace(input.Content), CreatedAt: now})
	if err != nil {
		return ContinueResult{}, err
	}
	exec, err := s.store.CreateExecution(ctx, Execution{ProjectID: project.ID, Status: ExecutionRunning, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return ContinueResult{}, err
	}
	answer, runtimeRef, executeErr := s.executor.Execute(ctx, exec, project, user, input.SkillIDs, input.AttachmentIDs)
	if executeErr != nil {
		exec.Status, exec.ErrorCode, exec.ErrorMessage = ExecutionFailed, "execution_failed", "执行失败"
		if executeErr == ErrExecutorUnavailable {
			exec.Status, exec.ErrorCode, exec.ErrorMessage = ExecutionUnavailable, "executor_unavailable", "当前没有可用执行器"
		}
		exec.UpdatedAt = s.now().UTC()
		saved, saveErr := s.store.UpdateExecution(ctx, exec)
		if saveErr != nil {
			return ContinueResult{}, saveErr
		}
		return ContinueResult{Project: project, User: user, Execution: saved}, executeErr
	}
	exec.Status, exec.RuntimeRef, exec.UpdatedAt = ExecutionCompleted, runtimeRef, s.now().UTC()
	saved, err := s.store.UpdateExecution(ctx, exec)
	if err != nil {
		return ContinueResult{}, err
	}
	assistant, err := s.store.CreateMessage(ctx, Message{ProjectID: project.ID, Role: RoleAssistant, Content: strings.TrimSpace(answer), CreatedAt: s.now().UTC()})
	if err != nil {
		return ContinueResult{}, err
	}
	return ContinueResult{Project: project, User: user, Assistant: &assistant, Execution: saved}, nil
}
