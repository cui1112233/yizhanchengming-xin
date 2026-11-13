package novelfetchworkshop

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
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

const releaseUniqueBooksScript = `
for i = 1, #KEYS do
  if redis.call('GET', KEYS[i]) == ARGV[1] then redis.call('DEL', KEYS[i]) end
end
return 1
`

const recoverStaleJobScript = `
if not redis.call('GET', KEYS[1]) then return 0 end
redis.call('LREM', KEYS[3], '0', ARGV[1])
redis.call('LREM', KEYS[4], '0', ARGV[1])
redis.call('DEL', KEYS[1], KEYS[2])
for i = 5, #KEYS do
  if redis.call('GET', KEYS[i]) == ARGV[1] then redis.call('DEL', KEYS[i]) end
end
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

func (q *RedisQueue) Claim(ctx context.Context, lease time.Duration) (ProcessJob, error) {
	seconds := int(lease.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	value, err := q.command(ctx, "BRPOPLPUSH", q.pendingKey(), q.processingKey(), strconv.Itoa(seconds))
	if err != nil || value == nil {
		if err == nil {
			err = ErrQueueEmpty
		}
		return ProcessJob{}, err
	}
	id := strings.TrimSpace(fmt.Sprint(value))
	raw, err := q.command(ctx, "GET", q.jobKey(id))
	if err != nil || raw == nil {
		return ProcessJob{}, fmt.Errorf("read queued novel fetch job: %w", err)
	}
	var job ProcessJob
	if err := json.Unmarshal([]byte(fmt.Sprint(raw)), &job); err != nil {
		return ProcessJob{}, fmt.Errorf("decode queued novel fetch job: %w", err)
	}
	if _, err := q.command(ctx, "SET", q.leaseKey(id), time.Now().UTC().Format(time.RFC3339Nano), "NX", "EX", strconv.Itoa(seconds)); err != nil {
		return ProcessJob{}, err
	}
	return processJobWithDedupeScope(job), nil
}

func (q *RedisQueue) RecoverStaleJob(ctx context.Context, jobID string) (bool, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return false, nil
	}
	value, err := q.command(ctx, "GET", q.jobKey(jobID))
	if err != nil || value == nil {
		return false, err
	}
	var job ProcessJob
	if err := json.Unmarshal([]byte(fmt.Sprint(value)), &job); err != nil {
		return false, fmt.Errorf("decode stale novel fetch queue job: %w", err)
	}
	if job.ID != jobID {
		return false, fmt.Errorf("stale novel fetch queue job id mismatch")
	}
	keys := []string{q.jobKey(jobID), q.leaseKey(jobID), q.processingKey(), q.pendingKey()}
	if strings.HasPrefix(job.DedupeKey, "books:") {
		keys = append(keys, q.jobBookDedupeKeys(job)...)
	} else if strings.TrimSpace(job.DedupeKey) != "" {
		keys = append(keys, q.dedupeKey(job.DedupeKey))
	}
	args := []string{"EVAL", recoverStaleJobScript, strconv.Itoa(len(keys))}
	args = append(args, keys...)
	args = append(args, jobID)
	recovered, err := q.command(ctx, args...)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(fmt.Sprint(recovered)) == "1", nil
}

func (q *RedisQueue) Ack(ctx context.Context, job ProcessJob) error {
	if _, err := q.command(ctx, "LREM", q.processingKey(), "0", job.ID); err != nil {
		return err
	}
	if _, err := q.command(ctx, "DEL", q.jobKey(job.ID), q.leaseKey(job.ID)); err != nil {
		return err
	}
	return q.releaseJobDedupe(ctx, job)
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
	if _, err := q.command(ctx, "LREM", q.processingKey(), "0", job.ID); err != nil {
		return err
	}
	if _, err := q.command(ctx, "DEL", q.leaseKey(job.ID)); err != nil {
		return err
	}
	if retry {
		if strings.HasPrefix(job.DedupeKey, "books:") {
			if err := q.releaseJobDedupe(ctx, processCompletedDedupeBooks(job)); err != nil {
				return err
			}
			job.DedupeBookIDs = append([]string{}, job.BookIDs...)
		}
		if err := q.Enqueue(ctx, job); err != nil {
			_, _ = q.command(ctx, "DEL", q.jobKey(job.ID))
			_ = q.releaseJobDedupe(ctx, job)
			return err
		}
		return nil
	}
	if _, err := q.command(ctx, "DEL", q.jobKey(job.ID)); err != nil {
		return err
	}
	return q.releaseJobDedupe(ctx, job)
}

func (q *RedisQueue) releaseJobDedupe(ctx context.Context, job ProcessJob) error {
	if strings.HasPrefix(job.DedupeKey, "books:") {
		keys := q.jobBookDedupeKeys(job)
		if len(keys) == 0 {
			return nil
		}
		args := []string{"EVAL", releaseUniqueBooksScript, strconv.Itoa(len(keys))}
		args = append(args, keys...)
		args = append(args, job.ID)
		_, err := q.command(ctx, args...)
		return err
	}
	if strings.TrimSpace(job.DedupeKey) == "" {
		return nil
	}
	_, err := q.command(ctx, "DEL", q.dedupeKey(job.DedupeKey))
	return err
}
