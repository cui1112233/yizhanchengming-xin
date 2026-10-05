package video

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

const LocalExecutorOnlineThreshold = 45 * time.Second

var (
	ErrLocalExecutorUnauthorized = errors.New("video: local executor unauthorized")
	ErrLocalExecutorInvalid      = errors.New("video: local executor invalid input")
	ErrLocalExecutorTaskNotFound = errors.New("video: local executor task not found")
)

type LocalExecutorRegistrationInput struct {
	Name         string   `json:"name"`
	ProviderKey  string   `json:"providerKey"`
	Model        string   `json:"model"`
	Capabilities []string `json:"capabilities"`
}

type LocalExecutorRegistrationResult struct {
	Executor LocalExecutorIdentity `json:"executor"`
	Token    string                `json:"token"`
}

type LocalExecutorHeartbeatInput struct {
	Capabilities []string `json:"capabilities"`
}

type LocalExecutorRecord struct {
	ID           string
	Name         string
	ProviderKey  string
	Model        string
	Capabilities []string
	TokenHash    [32]byte
	LastSeenAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type LocalExecutorIdentity struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	ProviderKey     string    `json:"providerKey"`
	Model           string    `json:"model"`
	Capabilities    []string  `json:"capabilities"`
	Online          bool      `json:"online"`
	LastSeenAt      time.Time `json:"lastSeenAt"`
	TokenConfigured bool      `json:"tokenConfigured"`
}

type LocalExecutorTask struct {
	ID           string
	SourceTaskID string
	ProviderKey  string
	Model        string
	Prompt       string
	RequestID    string
	Status       TaskStatus
	ExecutorID   string
	ArtifactURL  string
	ErrorCode    ErrorCode
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type LocalExecutorCompleteInput struct {
	ArtifactURL string `json:"artifactUrl"`
}

type LocalExecutorFailInput struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type LocalExecutorStore interface {
	CreateLocalExecutor(context.Context, LocalExecutorRecord) error
	GetLocalExecutorByTokenHash(context.Context, [32]byte) (LocalExecutorRecord, error)
	UpdateLocalExecutorHeartbeat(context.Context, string, []string, time.Time) error
	ListLocalExecutors(context.Context) ([]LocalExecutorRecord, error)
	CreateLocalExecutorTask(context.Context, LocalExecutorTask) error
	GetLocalExecutorTask(context.Context, string) (LocalExecutorTask, error)
	CompleteLocalExecutorTask(context.Context, string, string, string, time.Time) error
	FailLocalExecutorTask(context.Context, string, string, ErrorCode, string, time.Time) error
	CancelLocalExecutorTask(context.Context, string, time.Time) (bool, error)
}

// LocalExecutorLeaseCoordinator is deliberately only a Task 14 consumption
// boundary. Task 14 does not provide a Redis/Scheduler implementation. The
// adapter is expected to bind to the shared Task 9.4 queue/lease runtime once
// those public interfaces land on main.
type LocalExecutorLeaseCoordinator interface {
	Claim(context.Context, LocalExecutorIdentity) (LocalExecutorLease, error)
	Renew(context.Context, LocalExecutorLease) (LocalExecutorLease, error)
	Release(context.Context, LocalExecutorLease) error
	RequeueExpired(context.Context, time.Time, int) (int, error)
}

type LocalExecutorLease struct {
	TaskID       string    `json:"taskId"`
	ExecutorID   string    `json:"executorId"`
	Token        string    `json:"leaseToken"`
	Generation   int64     `json:"generation"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type LocalExecutorService struct {
	store LocalExecutorStore
	now   func() time.Time
}

func NewLocalExecutorService(store LocalExecutorStore, now func() time.Time) *LocalExecutorService {
	if now == nil {
		now = time.Now
	}
	return &LocalExecutorService{store: store, now: now}
}

func (s *LocalExecutorService) Register(ctx context.Context, input LocalExecutorRegistrationInput) (LocalExecutorRegistrationResult, error) {
	if s == nil || s.store == nil {
		return LocalExecutorRegistrationResult{}, providerError(ErrorProviderUnavailable, "local executor store unavailable", nil)
	}
	input.Name = strings.TrimSpace(input.Name)
	input.ProviderKey = strings.TrimSpace(input.ProviderKey)
	input.Model = strings.TrimSpace(input.Model)
	if input.Name == "" || input.ProviderKey == "" || input.Model == "" {
		return LocalExecutorRegistrationResult{}, ErrLocalExecutorInvalid
	}
	mapped, ok := ProviderForModel(input.Model)
	if !ok || mapped != ProviderDoubaoLocalExecutor || input.ProviderKey != ProviderDoubaoLocalExecutor {
		return LocalExecutorRegistrationResult{}, ErrLocalExecutorInvalid
	}
	capabilities, err := normalizeCapabilities(input.Capabilities)
	if err != nil {
		return LocalExecutorRegistrationResult{}, err
	}
	id, err := randomLocalExecutorValue("lex_", 12)
	if err != nil {
		return LocalExecutorRegistrationResult{}, err
	}
	token, err := randomLocalExecutorValue("", 32)
	if err != nil {
		return LocalExecutorRegistrationResult{}, err
	}
	now := s.now().UTC()
	record := LocalExecutorRecord{
		ID: id, Name: input.Name, ProviderKey: input.ProviderKey, Model: input.Model,
		Capabilities: capabilities, TokenHash: sha256.Sum256([]byte(token)),
		LastSeenAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.CreateLocalExecutor(ctx, record); err != nil {
		return LocalExecutorRegistrationResult{}, err
	}
	return LocalExecutorRegistrationResult{Executor: localExecutorView(record, now), Token: token}, nil
}

func (s *LocalExecutorService) Identity(ctx context.Context, token string) (LocalExecutorIdentity, error) {
	record, err := s.executorForToken(ctx, token)
	if err != nil {
		return LocalExecutorIdentity{}, err
	}
	return localExecutorView(record, s.now().UTC()), nil
}

func (s *LocalExecutorService) Heartbeat(ctx context.Context, token string, input LocalExecutorHeartbeatInput) error {
	record, err := s.executorForToken(ctx, token)
	if err != nil {
		return err
	}
	capabilities, err := normalizeCapabilities(input.Capabilities)
	if err != nil {
		return err
	}
	if len(capabilities) == 0 {
		capabilities = append([]string(nil), record.Capabilities...)
	}
	return s.store.UpdateLocalExecutorHeartbeat(ctx, record.ID, capabilities, s.now().UTC())
}

func (s *LocalExecutorService) List(ctx context.Context) ([]LocalExecutorIdentity, error) {
	if s == nil || s.store == nil {
		return nil, providerError(ErrorProviderUnavailable, "local executor store unavailable", nil)
	}
	records, err := s.store.ListLocalExecutors(ctx)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	out := make([]LocalExecutorIdentity, 0, len(records))
	for _, record := range records {
		out = append(out, localExecutorView(record, now))
	}
	return out, nil
}

func (s *LocalExecutorService) CompleteTask(ctx context.Context, token, taskID string, input LocalExecutorCompleteInput) error {
	executor, err := s.executorForToken(ctx, token)
	if err != nil {
		return err
	}
	task, err := s.store.GetLocalExecutorTask(ctx, strings.TrimSpace(taskID))
	if err != nil {
		return err
	}
	if !localExecutorSupports(executor, task.ProviderKey, task.Model) || task.ExecutorID == "" || task.ExecutorID != executor.ID {
		return ErrLocalExecutorUnauthorized
	}
	artifactURL := strings.TrimSpace(input.ArtifactURL)
	parsed, err := url.Parse(artifactURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return ErrLocalExecutorInvalid
	}
	return s.store.CompleteLocalExecutorTask(ctx, task.ID, executor.ID, artifactURL, s.now().UTC())
}

func (s *LocalExecutorService) FailTask(ctx context.Context, token, taskID string, input LocalExecutorFailInput) error {
	executor, err := s.executorForToken(ctx, token)
	if err != nil {
		return err
	}
	task, err := s.store.GetLocalExecutorTask(ctx, strings.TrimSpace(taskID))
	if err != nil {
		return err
	}
	if !localExecutorSupports(executor, task.ProviderKey, task.Model) || task.ExecutorID == "" || task.ExecutorID != executor.ID {
		return ErrLocalExecutorUnauthorized
	}
	code := ErrorCode(strings.TrimSpace(input.Code))
	if code == "" {
		code = ErrorProviderRequestFailed
	}
	return s.store.FailLocalExecutorTask(ctx, task.ID, executor.ID, code, safeExecutorErrorMessage(input.Message), s.now().UTC())
}

func (s *LocalExecutorService) executorForToken(ctx context.Context, token string) (LocalExecutorRecord, error) {
	if s == nil || s.store == nil || strings.TrimSpace(token) == "" {
		return LocalExecutorRecord{}, ErrLocalExecutorUnauthorized
	}
	record, err := s.store.GetLocalExecutorByTokenHash(ctx, sha256.Sum256([]byte(strings.TrimSpace(token))))
	if err != nil {
		return LocalExecutorRecord{}, ErrLocalExecutorUnauthorized
	}
	return record, nil
}

func localExecutorView(record LocalExecutorRecord, now time.Time) LocalExecutorIdentity {
	return LocalExecutorIdentity{
		ID: record.ID, Name: record.Name, ProviderKey: record.ProviderKey, Model: record.Model,
		Capabilities: append([]string(nil), record.Capabilities...),
		Online: record.LastSeenAt.Add(LocalExecutorOnlineThreshold).After(now),
		LastSeenAt: record.LastSeenAt, TokenConfigured: record.TokenHash != [32]byte{},
	}
}

func localExecutorSupports(record LocalExecutorRecord, providerKey, model string) bool {
	return record.ProviderKey == providerKey && record.Model == model
}

func normalizeCapabilities(values []string) ([]string, error) {
	set := map[string]struct{}{}
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if len(value) > 64 {
			return nil, ErrLocalExecutorInvalid
		}
		set[value] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out, nil
}

func randomLocalExecutorValue(prefix string, size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + base64.RawURLEncoding.EncodeToString(raw), nil
}

type doubaoLocalExecutorProvider struct {
	store LocalExecutorStore
	now   func() time.Time
}

func NewDoubaoLocalExecutorProvider(store LocalExecutorStore) Provider {
	return &doubaoLocalExecutorProvider{store: store, now: time.Now}
}

func (p *doubaoLocalExecutorProvider) Submit(ctx context.Context, input SubmitRequest) (SubmitResult, error) {
	if p == nil || p.store == nil {
		return SubmitResult{}, providerError(ErrorProviderUnavailable, "local executor store unavailable", nil)
	}
	model := strings.TrimSpace(input.Model)
	mapped, ok := ProviderForModel(model)
	if !ok || mapped != ProviderDoubaoLocalExecutor {
		return SubmitResult{}, providerError(ErrorProviderUnavailable, "local executor model mismatch", nil)
	}
	if strings.TrimSpace(input.Prompt) == "" || strings.TrimSpace(input.SourceTaskID) == "" {
		return SubmitResult{}, providerError(ErrorProviderInvalidResponse, "local executor prompt/source task is required", nil)
	}
	id, err := randomLocalExecutorValue("let_", 12)
	if err != nil {
		return SubmitResult{}, providerError(ErrorProviderRequestFailed, "create local executor task id", err)
	}
	now := p.now().UTC()
	job := LocalExecutorTask{
		ID: id, SourceTaskID: strings.TrimSpace(input.SourceTaskID), ProviderKey: ProviderDoubaoLocalExecutor,
		Model: model, Prompt: input.Prompt, RequestID: strings.TrimSpace(input.RequestID),
		Status: TaskQueued, CreatedAt: now, UpdatedAt: now,
	}
	if err := p.store.CreateLocalExecutorTask(ctx, job); err != nil {
		return SubmitResult{}, providerError(ErrorProviderRequestFailed, "persist local executor task", err)
	}
	return SubmitResult{ProviderJobID: id, Status: TaskQueued}, nil
}

func (p *doubaoLocalExecutorProvider) Poll(ctx context.Context, providerJobID string) (PollResult, error) {
	job, err := p.store.GetLocalExecutorTask(ctx, strings.TrimSpace(providerJobID))
	if err != nil {
		return PollResult{}, providerError(ErrorProviderRequestFailed, "read local executor task", err)
	}
	return PollResult{Status: job.Status, ArtifactURL: job.ArtifactURL, ErrorCode: job.ErrorCode, ErrorMessage: job.ErrorMessage}, nil
}

func (p *doubaoLocalExecutorProvider) Cancel(ctx context.Context, providerJobID string) (CancelResult, error) {
	accepted, err := p.store.CancelLocalExecutorTask(ctx, strings.TrimSpace(providerJobID), p.now().UTC())
	if err != nil {
		return CancelResult{}, providerError(ErrorProviderRequestFailed, "cancel local executor task", err)
	}
	if !accepted {
		return CancelResult{Accepted: false}, providerError(ErrorProviderCancelUnsupported, "local executor task is terminal", nil)
	}
	return CancelResult{Accepted: true, Status: TaskCancelled}, nil
}

func (p *doubaoLocalExecutorProvider) Probe(ctx context.Context) error {
	records, err := p.store.ListLocalExecutors(ctx)
	if err != nil {
		return providerError(ErrorProviderUnavailable, "local executor registry unavailable", err)
	}
	now := p.now().UTC()
	for _, record := range records {
		if record.ProviderKey == ProviderDoubaoLocalExecutor && record.LastSeenAt.Add(LocalExecutorOnlineThreshold).After(now) {
			return nil
		}
	}
	return providerError(ErrorProviderUnavailable, "no online local executor", nil)
}

func localExecutorTokenFromAuthorization(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return value
}

func providerConfigNeedsSecret(providerKey string) bool {
	return strings.TrimSpace(providerKey) != ProviderDoubaoLocalExecutor
}

func providerConfigConfigured(config ProviderConfig) bool {
	if !config.Enabled {
		return false
	}
	if !providerConfigNeedsSecret(config.ProviderKey) {
		return true
	}
	return len(config.EncryptedSecret) > 0 && len(config.SecretNonce) > 0
}

func localExecutorStoreFromVideoStore(store Store) (LocalExecutorStore, error) {
	local, ok := store.(LocalExecutorStore)
	if !ok {
		return nil, fmt.Errorf("video: local executor store unavailable")
	}
	return local, nil
}
