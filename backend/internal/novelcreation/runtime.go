package novelcreation

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/mergeworker"
)

const runtimeCancelTTL = 24 * time.Hour

const (
	defaultRuntimePrefix = "qiantie:novel-creation"
	runtimeCancelSeconds = 24 * 60 * 60
)

var ErrRuntimeQueueEmpty = errors.New("novel creation runtime queue is empty")

// Runtime coordinates short-lived work delivery. Run events remain durable in
// MySQL; Publish is only a wake-up path for connected consumers.
type Runtime interface {
	Enqueue(context.Context, string) error
	Dequeue(context.Context) (string, error)
	AcquireProject(context.Context, string, string, time.Duration) (func(), error)
	Publish(context.Context, RunEvent) error
	Cancel(context.Context, string) error
	IsCancelled(context.Context, string) (bool, error)
}

// RedisDoFunc deliberately has the existing merge worker transport shape so
// this package does not bring in another Redis SDK.
type RedisDoFunc = mergeworker.RedisDoFunc

const acquireProjectLockScript = `
if redis.call('SET', KEYS[1], ARGV[1], 'NX', 'PX', ARGV[2]) then return 1 end
return 0
`

const releaseProjectLockScript = `
if redis.call('GET', KEYS[1]) == ARGV[1] then return redis.call('DEL', KEYS[1]) end
return 0
`

type memoryProjectLock struct {
	token     string
	expiresAt time.Time
}

// MemoryRuntime is a local-only runtime used by tests and explicit development
// configurations. Production must use the Redis runtime.
type MemoryRuntime struct {
	mu        sync.Mutex
	locks     map[string]memoryProjectLock
	cancelled map[string]time.Time
	queue     []string
	events    []RunEvent
}

func NewMemoryRuntime() *MemoryRuntime {
	return &MemoryRuntime{
		locks:     make(map[string]memoryProjectLock),
		cancelled: make(map[string]time.Time),
	}
}

func (r *MemoryRuntime) Enqueue(_ context.Context, runID string) error {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return &DomainError{Code: ErrInputRequired, Message: "runId is required"}
	}
	r.mu.Lock()
	r.queue = append(r.queue, runID)
	r.mu.Unlock()
	return nil
}

func (r *MemoryRuntime) Dequeue(_ context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.queue) == 0 {
		return "", ErrRuntimeQueueEmpty
	}
	runID := r.queue[0]
	r.queue = r.queue[1:]
	return runID, nil
}

func (r *MemoryRuntime) Publish(_ context.Context, event RunEvent) error {
	if strings.TrimSpace(event.RunID) == "" {
		return &DomainError{Code: ErrInputRequired, Message: "run event runId is required"}
	}
	r.mu.Lock()
	r.events = append(r.events, cloneRunEvent(event))
	r.mu.Unlock()
	return nil
}

func (r *MemoryRuntime) AcquireProject(_ context.Context, projectID, _ string, ttl time.Duration) (func(), error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, &DomainError{Code: ErrInputRequired, Message: "projectId is required"}
	}
	if ttl <= 0 {
		return nil, &DomainError{Code: ErrInputRequired, Message: "project lock ttl must be positive"}
	}
	token, err := runtimeToken()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	r.mu.Lock()
	if held, exists := r.locks[projectID]; exists && held.expiresAt.After(now) {
		r.mu.Unlock()
		return nil, &DomainError{Code: ErrProjectLocked, Message: "project is already being written"}
	}
	r.locks[projectID] = memoryProjectLock{token: token, expiresAt: now.Add(ttl)}
	r.mu.Unlock()

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if held, exists := r.locks[projectID]; exists && held.token == token {
			delete(r.locks, projectID)
		}
	}, nil
}

func (r *MemoryRuntime) Cancel(_ context.Context, runID string) error {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return &DomainError{Code: ErrInputRequired, Message: "runId is required"}
	}
	r.mu.Lock()
	r.cancelled[runID] = time.Now().UTC().Add(runtimeCancelTTL)
	r.mu.Unlock()
	return nil
}

func (r *MemoryRuntime) IsCancelled(_ context.Context, runID string) (bool, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return false, &DomainError{Code: ErrInputRequired, Message: "runId is required"}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	expiresAt, cancelled := r.cancelled[runID]
	if cancelled && !expiresAt.After(time.Now().UTC()) {
		delete(r.cancelled, runID)
		cancelled = false
	}
	return cancelled, nil
}

func runtimeToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// RedisRuntime is the production implementation. It never falls back to the
// memory runtime: callers receive a Redis error whenever Redis is unavailable.
type RedisRuntime struct {
	prefix string
	do     RedisDoFunc
}

var _ Runtime = (*MemoryRuntime)(nil)
var _ Runtime = (*RedisRuntime)(nil)

// NewRedisRuntime keeps prefix for call-site compatibility. Novel-creation
// runtime keys are intentionally always rooted at defaultRuntimePrefix so a
// deployment setting cannot split writers, cancellation, and event consumers
// across different Redis namespaces.
func NewRedisRuntime(_ string, do RedisDoFunc) *RedisRuntime {
	return &RedisRuntime{prefix: defaultRuntimePrefix, do: do}
}

func (r *RedisRuntime) queueKey() string { return r.prefix + ":queue" }
func (r *RedisRuntime) projectLockKey(projectID string) string {
	return r.prefix + ":project-lock:" + projectID
}
func (r *RedisRuntime) cancelKey(runID string) string    { return r.prefix + ":cancel:" + runID }
func (r *RedisRuntime) eventChannel(runID string) string { return r.prefix + ":events:" + runID }

func (r *RedisRuntime) command(ctx context.Context, args ...string) (any, error) {
	if r == nil || r.do == nil {
		return nil, errors.New("novel creation Redis runtime is unavailable")
	}
	return r.do(ctx, args...)
}

func (r *RedisRuntime) Enqueue(ctx context.Context, runID string) error {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return &DomainError{Code: ErrInputRequired, Message: "runId is required"}
	}
	_, err := r.command(ctx, "RPUSH", r.queueKey(), runID)
	return err
}

func (r *RedisRuntime) Dequeue(ctx context.Context) (string, error) {
	value, err := r.command(ctx, "LPOP", r.queueKey())
	if err != nil {
		return "", err
	}
	runID := strings.TrimSpace(fmt.Sprint(value))
	if value == nil || runID == "" || runID == "<nil>" {
		return "", ErrRuntimeQueueEmpty
	}
	return runID, nil
}

func (r *RedisRuntime) AcquireProject(ctx context.Context, projectID, _ string, ttl time.Duration) (func(), error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, &DomainError{Code: ErrInputRequired, Message: "projectId is required"}
	}
	if ttl <= 0 {
		return nil, &DomainError{Code: ErrInputRequired, Message: "project lock ttl must be positive"}
	}
	token, err := runtimeToken()
	if err != nil {
		return nil, err
	}
	milliseconds := int(ttl.Round(time.Millisecond) / time.Millisecond)
	if milliseconds < 1 {
		milliseconds = 1
	}
	key := r.projectLockKey(projectID)
	value, err := r.command(ctx, "EVAL", acquireProjectLockScript, "1", key, token, strconv.Itoa(milliseconds))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(fmt.Sprint(value)) != "1" {
		return nil, &DomainError{Code: ErrProjectLocked, Message: "project is already being written"}
	}
	return func() {
		// The agreed callback cannot report an error. A failed release never
		// falls back to memory or an unsafe DEL; Redis lock TTL remains the
		// bounded recovery path until the next worker can acquire it.
		_, _ = r.command(context.Background(), "EVAL", releaseProjectLockScript, "1", key, token)
	}, nil
}

func (r *RedisRuntime) Publish(ctx context.Context, event RunEvent) error {
	event.RunID = strings.TrimSpace(event.RunID)
	if event.RunID == "" {
		return &DomainError{Code: ErrInputRequired, Message: "run event runId is required"}
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = r.command(ctx, "PUBLISH", r.eventChannel(event.RunID), string(payload))
	return err
}

func (r *RedisRuntime) Cancel(ctx context.Context, runID string) error {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return &DomainError{Code: ErrInputRequired, Message: "runId is required"}
	}
	_, err := r.command(ctx, "SET", r.cancelKey(runID), "1", "EX", strconv.Itoa(runtimeCancelSeconds))
	return err
}

func (r *RedisRuntime) IsCancelled(ctx context.Context, runID string) (bool, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return false, &DomainError{Code: ErrInputRequired, Message: "runId is required"}
	}
	value, err := r.command(ctx, "EXISTS", r.cancelKey(runID))
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(fmt.Sprint(value)) == "1", nil
}
