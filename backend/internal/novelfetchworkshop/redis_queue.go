package novelfetchworkshop

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/mergeworker"
)

type RedisDoFunc func(context.Context, ...string) (any, error)

type RedisQueue struct {
	Prefix string
	Do     RedisDoFunc
}

const enqueueUniqueBooksScript = `
local count = tonumber(ARGV[4])
for i = 1, count do
  local existing = redis.call('GET', KEYS[i])
  if existing then return {0, existing} end
end
for i = 1, count do redis.call('SET', KEYS[i], ARGV[1], 'EX', ARGV[2]) end
redis.call('SET', KEYS[count + 1], ARGV[3])
redis.call('RPUSH', KEYS[count + 2], ARGV[1])
return {1, ARGV[1]}
`

// The blocking pop and SET NX cannot share a Redis script. If recovery ran
// between them, confirm both the processing entry and the exact payload
// before allowing this claimant to work on it.
const confirmClaimScript = `
if redis.call('GET', KEYS[2]) ~= ARGV[2] then return false end
if redis.call('GET', KEYS[1]) == ARGV[3] then
  for _, id in ipairs(redis.call('LRANGE', KEYS[3], 0, -1)) do
    if id == ARGV[1] then return 'OK' end
  end
end
redis.call('DEL', KEYS[2])
return false
`

const renewLeaseScript = `
if redis.call('GET', KEYS[1]) ~= ARGV[1] then return 0 end
return redis.call('EXPIRE', KEYS[1], ARGV[2])
`

const requireProcessingLeaseScript = `
if redis.call('GET', KEYS[2]) ~= ARGV[2] then return 0 end
`

const ackProcessingJobScript = requireProcessingLeaseScript + `
local release = {}
for i = 4, #KEYS do
  if redis.call('GET', KEYS[i]) == ARGV[1] then release[#release + 1] = KEYS[i] end
end
if redis.call('LREM', KEYS[3], '0', ARGV[1]) == 0 then return 0 end
redis.call('DEL', KEYS[1], KEYS[2])
for _, key in ipairs(release) do redis.call('DEL', key) end
return 1
`

// Redis scripts are atomic to other clients but do not roll back runtime errors.
// Check the two list keys and any values read by the transition before writing.
const validateQueueListsScript = `
for i = 3, 4 do
  local kind = redis.call('TYPE', KEYS[i]).ok
  if kind ~= 'list' and kind ~= 'none' then
    return redis.error_reply('WRONGTYPE novel fetch queue index must be a list')
  end
end
`

const retryProcessingJobScript = requireProcessingLeaseScript + `
local current = redis.call('GET', KEYS[1])
if not current then return 0 end
` + validateQueueListsScript + `
local release = {}
for i = 5, #KEYS do
  if redis.call('GET', KEYS[i]) == ARGV[1] then release[#release + 1] = KEYS[i] end
end
if redis.call('LREM', KEYS[3], '0', ARGV[1]) == 0 then return 0 end
redis.call('LREM', KEYS[4], '0', ARGV[1])
redis.call('SET', KEYS[1], ARGV[3])
redis.call('LPUSH', KEYS[4], ARGV[1])
redis.call('DEL', KEYS[2])
for _, key in ipairs(release) do redis.call('DEL', key) end
return 1
`

const recoverStaleJobScript = `
if redis.call('GET', KEYS[2]) then return 0 end
local job = redis.call('GET', KEYS[1])
` + validateQueueListsScript + `
if redis.call('LREM', KEYS[3], '0', ARGV[1]) == 0 then return 0 end
redis.call('LREM', KEYS[4], '0', ARGV[1])
if job then redis.call('LPUSH', KEYS[4], ARGV[1]) end
return 1
`

func NewRedisQueueWithDo(prefix string, do RedisDoFunc) *RedisQueue {
	prefix = strings.Trim(strings.TrimSpace(prefix), ":")
	if prefix == "" {
		prefix = "qiantie:novel-fetch"
	}
	return &RedisQueue{Prefix: prefix, Do: do}
}

func (q *RedisQueue) pendingKey() string      { return q.Prefix + ":pending" }
func (q *RedisQueue) processingKey() string   { return q.Prefix + ":processing" }
func (q *RedisQueue) jobKey(id string) string { return q.Prefix + ":job:" + id }

func (q *RedisQueue) ProcessJobExists(ctx context.Context, id string) (bool, error) {
	value, err := q.command(ctx, "EXISTS", q.jobKey(strings.TrimSpace(id)))
	if err != nil {
		return false, err
	}
	count, err := strconv.Atoi(fmt.Sprint(value))
	return count > 0, err
}
func (q *RedisQueue) leaseKey(id string) string { return q.Prefix + ":lease:" + id }
func (q *RedisQueue) dedupeKey(key string) string {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(key))))
	return q.Prefix + ":dedupe:" + hash[:24]
}
func (q *RedisQueue) bookDedupeKey(owner, platformID, bookID string) string {
	identity := strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(platformID) + "\x00" + strings.TrimSpace(bookID)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(identity)))
	return q.Prefix + ":book-dedupe:" + hash[:32]
}
func (q *RedisQueue) cancelKey(owner, bookID string) string {
	ownerHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(owner))))
	return q.Prefix + ":cancel:" + ownerHash[:16] + ":" + strings.TrimSpace(bookID)
}

func (q *RedisQueue) CancelBooks(ctx context.Context, owner string, bookIDs []string) error {
	for _, bookID := range bookIDs {
		if _, err := q.command(ctx, "SET", q.cancelKey(owner, bookID), "1", "EX", "604800"); err != nil {
			return err
		}
	}
	return nil
}

func (q *RedisQueue) ClearCancelledBooks(ctx context.Context, owner string, bookIDs []string) error {
	if len(bookIDs) == 0 {
		return nil
	}
	args := []string{"DEL"}
	for _, bookID := range bookIDs {
		args = append(args, q.cancelKey(owner, bookID))
	}
	_, err := q.command(ctx, args...)
	return err
}

func (q *RedisQueue) IsBookCancelled(ctx context.Context, owner, bookID string) (bool, error) {
	value, err := q.command(ctx, "GET", q.cancelKey(owner, bookID))
	return value != nil && strings.TrimSpace(fmt.Sprint(value)) != "", err
}

func (q *RedisQueue) command(ctx context.Context, args ...string) (any, error) {
	if q == nil || q.Do == nil {
		return nil, fmt.Errorf("redis queue unavailable")
	}
	return q.Do(ctx, args...)
}

func (q *RedisQueue) Enqueue(ctx context.Context, job ProcessJob) error {
	if err := validateProcessJob(job); err != nil {
		return err
	}
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	if _, err := q.command(ctx, "SET", q.jobKey(job.ID), string(body)); err != nil {
		return err
	}
	_, err = q.command(ctx, "RPUSH", q.pendingKey(), job.ID)
	return err
}

func (q *RedisQueue) EnqueueUnique(ctx context.Context, key string, job ProcessJob, ttl time.Duration) (bool, error) {
	seconds := int(ttl.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 3600
	}
	dedupeKey := q.dedupeKey(key)
	value, err := q.command(ctx, "SET", dedupeKey, job.ID, "NX", "EX", strconv.Itoa(seconds))
	if err != nil || value == nil {
		return false, err
	}
	job.DedupeKey = key
	if err := q.Enqueue(ctx, job); err != nil {
		_, _ = q.command(ctx, "DEL", dedupeKey)
		return false, err
	}
	return true, nil
}

func (q *RedisQueue) EnqueueUniqueBooks(ctx context.Context, job ProcessJob, ttl time.Duration) (bool, string, error) {
	if err := validateProcessJob(job); err != nil {
		return false, "", err
	}
	if len(job.BookIDs) == 0 || strings.TrimSpace(job.PlatformID) == "" {
		return false, "", fmt.Errorf("platform and source book ids are required for unique enqueue")
	}
	seconds := int(ttl.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 3600
	}
	seen := map[string]bool{}
	keys := make([]string, 0, len(job.BookIDs)+2)
	for _, bookID := range job.BookIDs {
		bookID = strings.TrimSpace(bookID)
		if bookID == "" || seen[bookID] {
			continue
		}
		seen[bookID] = true
		keys = append(keys, q.bookDedupeKey(job.Owner, job.PlatformID, bookID))
	}
	if len(keys) == 0 {
		return false, "", fmt.Errorf("source book ids are required for unique enqueue")
	}
	job.DedupeKey = processBookDedupeMarker(job)
	job.DedupeBookIDs = append([]string{}, job.BookIDs...)
	body, err := json.Marshal(job)
	if err != nil {
		return false, "", err
	}
	bookKeyCount := len(keys)
	keys = append(keys, q.jobKey(job.ID), q.pendingKey())
	args := []string{"EVAL", enqueueUniqueBooksScript, strconv.Itoa(len(keys))}
	args = append(args, keys...)
	args = append(args, job.ID, strconv.Itoa(seconds), string(body), strconv.Itoa(bookKeyCount))
	value, err := q.command(ctx, args...)
	if err != nil {
		return false, "", err
	}
	result, ok := value.([]any)
	if !ok || len(result) < 2 {
		return false, "", fmt.Errorf("unexpected redis response for unique novel fetch enqueue: %T", value)
	}
	queued := strings.TrimSpace(fmt.Sprint(result[0])) == "1"
	existingID := strings.TrimSpace(fmt.Sprint(result[1]))
	return queued, existingID, nil
}

func (q *RedisQueue) ConflictingBookIDs(ctx context.Context, job ProcessJob) (map[string]string, error) {
	seen := map[string]bool{}
	bookIDs := make([]string, 0, len(job.BookIDs))
	keys := make([]string, 0, len(job.BookIDs)+1)
	for _, bookID := range job.BookIDs {
		bookID = strings.TrimSpace(bookID)
		if bookID == "" || seen[bookID] {
			continue
		}
		seen[bookID] = true
		bookIDs = append(bookIDs, bookID)
		keys = append(keys, q.bookDedupeKey(job.Owner, job.PlatformID, bookID))
	}
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	value, err := q.command(ctx, append([]string{"MGET"}, keys...)...)
	if err != nil {
		return nil, err
	}
	values, ok := value.([]any)
	if !ok || len(values) != len(keys) {
		return nil, fmt.Errorf("unexpected redis response for novel fetch conflicts: %T", value)
	}
	conflicts := map[string]string{}
	for i, raw := range values {
		if raw == nil {
			continue
		}
		if existingID := strings.TrimSpace(fmt.Sprint(raw)); existingID != "" {
			conflicts[bookIDs[i]] = existingID
		}
	}
	return conflicts, nil
}

// The production RESP transport reports Redis nil as a typed error.
// Queue and lease absence are ordinary return values for these commands.
func (q *RedisQueue) leaseCommand(ctx context.Context, args ...string) (any, error) {
	value, err := q.command(ctx, args...)
	if mergeworker.IsRedisNil(err) {
		return nil, nil
	}
	return value, err
}

func (q *RedisQueue) Claim(ctx context.Context, wait time.Duration) (ProcessJob, error) {
	seconds := int(wait.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	value, err := q.leaseCommand(ctx, "BRPOPLPUSH", q.pendingKey(), q.processingKey(), strconv.Itoa(seconds))
	if err != nil || value == nil {
		if err == nil {
			err = ErrQueueEmpty
		}
		return ProcessJob{}, err
	}
	id := strings.TrimSpace(fmt.Sprint(value))
	raw, err := q.leaseCommand(ctx, "GET", q.jobKey(id))
	if err != nil {
		return ProcessJob{}, fmt.Errorf("read queued novel fetch job: %w", err)
	}
	if raw == nil {
		return ProcessJob{}, ErrProcessLeaseLost
	}
	var job ProcessJob
	if err := json.Unmarshal([]byte(fmt.Sprint(raw)), &job); err != nil {
		return ProcessJob{}, fmt.Errorf("decode queued novel fetch job: %w", err)
	}
	if job.ID != id {
		return ProcessJob{}, fmt.Errorf("queued novel fetch job id mismatch for %q", id)
	}
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return ProcessJob{}, err
	}
	job.LeaseToken = hex.EncodeToString(token[:])
	value, err = q.leaseCommand(ctx, "SET", q.leaseKey(id), job.LeaseToken, "NX", "EX", strconv.Itoa(int(DefaultProcessLeaseTTL/time.Second)))
	if err != nil {
		return ProcessJob{}, err
	}
	if value != "OK" {
		return ProcessJob{}, ErrProcessLeaseLost
	}
	value, err = q.leaseCommand(ctx, "EVAL", confirmClaimScript, "3", q.jobKey(id), q.leaseKey(id), q.processingKey(), id, job.LeaseToken, fmt.Sprint(raw))
	if err != nil {
		return ProcessJob{}, err
	}
	if value != "OK" {
		return ProcessJob{}, ErrProcessLeaseLost
	}
	return processJobWithDedupeScope(job), nil
}

func (q *RedisQueue) RecoverStaleJob(ctx context.Context, jobID string) (bool, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return false, nil
	}
	result, err := q.command(ctx, "EVAL", recoverStaleJobScript, "4",
		q.jobKey(jobID), q.leaseKey(jobID), q.processingKey(), q.pendingKey(), jobID)
	if err != nil {
		return false, err
	}
	switch fmt.Sprint(result) {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("unexpected redis stale recovery response: %v", result)
	}
}

// ReapStaleProcessing requeues the original payload for jobs with an expired
// lease. Redis removes only the list references when the payload is missing.
func (q *RedisQueue) ReapStaleProcessing(ctx context.Context) (int, error) {
	value, err := q.command(ctx, "LRANGE", q.processingKey(), "0", "-1")
	if err != nil {
		return 0, err
	}
	reaped := 0
	for _, id := range redisStringSlice(value) {
		recovered, recoverErr := q.RecoverStaleJob(ctx, id)
		if recoverErr != nil {
			return reaped, recoverErr
		}
		if recovered {
			reaped++
		}
	}
	return reaped, nil
}

func redisStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, fmt.Sprint(item))
		}
		return out
	default:
		return nil
	}
}

func (q *RedisQueue) Ack(ctx context.Context, job ProcessJob) error {
	keys := []string{q.jobKey(job.ID), q.leaseKey(job.ID), q.processingKey()}
	if strings.HasPrefix(job.DedupeKey, "books:") {
		keys = append(keys, q.jobBookDedupeKeys(job)...)
	} else if job.DedupeKey != "" {
		keys = append(keys, q.dedupeKey(job.DedupeKey))
	}
	args := append([]string{"EVAL", ackProcessingJobScript, strconv.Itoa(len(keys))}, keys...)
	return q.ownedTransition(ctx, append(args, job.ID, job.LeaseToken)...)
}

func (q *RedisQueue) jobBookDedupeKeys(job ProcessJob) []string {
	if len(job.DedupeBookIDs) > 0 {
		job.BookIDs = job.DedupeBookIDs
	}
	seen := map[string]bool{}
	keys := make([]string, 0, len(job.BookIDs))
	for _, bookID := range job.BookIDs {
		bookID = strings.TrimSpace(bookID)
		if bookID == "" || seen[bookID] {
			continue
		}
		seen[bookID] = true
		keys = append(keys, q.bookDedupeKey(job.Owner, job.PlatformID, bookID))
	}
	return keys
}

func (q *RedisQueue) Fail(ctx context.Context, job ProcessJob, retry bool) error {
	if !retry {
		return q.Ack(ctx, job)
	}
	if err := validateProcessJob(job); err != nil {
		return err
	}
	keys := []string{q.jobKey(job.ID), q.leaseKey(job.ID), q.processingKey(), q.pendingKey()}
	if strings.HasPrefix(job.DedupeKey, "books:") {
		keys = append(keys, q.jobBookDedupeKeys(processCompletedDedupeBooks(job))...)
		job.DedupeBookIDs = append([]string{}, job.BookIDs...)
	}
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	args := []string{"EVAL", retryProcessingJobScript, strconv.Itoa(len(keys))}
	args = append(args, keys...)
	return q.ownedTransition(ctx, append(args, job.ID, job.LeaseToken, string(body))...)
}

func (q *RedisQueue) ownedTransition(ctx context.Context, args ...string) error {
	value, err := q.command(ctx, args...)
	if err != nil {
		return err
	}
	switch fmt.Sprint(value) {
	case "1":
		return nil
	case "0":
		return ErrProcessLeaseLost
	default:
		return fmt.Errorf("unexpected redis lease response: %v", value)
	}
}

func (q *RedisQueue) RenewLease(ctx context.Context, job ProcessJob, ttl time.Duration) error {
	if ttl < time.Second {
		return fmt.Errorf("process lease TTL must be at least one second")
	}
	return q.ownedTransition(ctx, "EVAL", renewLeaseScript, "1", q.leaseKey(job.ID), job.LeaseToken, strconv.FormatInt(int64(ttl/time.Second), 10))
}

func (q *RedisQueue) OwnsLease(ctx context.Context, job ProcessJob) (bool, error) {
	value, err := q.leaseCommand(ctx, "GET", q.leaseKey(job.ID))
	return err == nil && job.LeaseToken != "" && value == job.LeaseToken, err
}
