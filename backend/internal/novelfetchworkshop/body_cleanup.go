package novelfetchworkshop

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"
)

const (
	DefaultBodyRetentionDays = 7
	MinBodyRetentionDays     = 1
	MaxBodyRetentionDays     = 30

	BodyCleanupReasonExpired  BodyCleanupReason = "expired"
	BodyCleanupReasonCapacity BodyCleanupReason = "capacity"
)

type BodyCleanupReason string

type BodyCleanupRequest struct {
	Now    time.Time
	Reason BodyCleanupReason
	Limit  int
}

type BodyCleanupRecord struct {
	BookID    string            `json:"bookId"`
	VersionID string            `json:"versionId"`
	Reason    BodyCleanupReason `json:"reason"`
}

type BodyCleanupResult struct {
	Deleted int                 `json:"deleted"`
	Results []BodyCleanupRecord `json:"results"`
}

func NormalizeBodyRetentionDays(days int) int {
	if days == 0 {
		return DefaultBodyRetentionDays
	}
	if days < MinBodyRetentionDays {
		return MinBodyRetentionDays
	}
	if days > MaxBodyRetentionDays {
		return MaxBodyRetentionDays
	}
	return days
}

func normalizeCleanupNow(now time.Time) time.Time {
	if now.IsZero() {
		return time.Now().UTC()
	}
	return now.UTC()
}

func normalizeCleanupLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	return limit
}

func parseLifecycleTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func (s *MemoryStore) MarkBodyReleasable(_ context.Context, owner, bookID, versionID string, releasableAt time.Time, retentionDays int) (BodyRef, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	versions := s.bodies[owner][bookID]
	if versions == nil {
		return BodyRef{}, ErrNotFound
	}
	body, ok := versions[versionID]
	if !ok {
		return BodyRef{}, ErrNotFound
	}
	at := normalizeCleanupNow(releasableAt)
	body.State = "releasable"
	body.ReleasableAt = at.Format(time.RFC3339Nano)
	body.ExpiresAt = at.Add(time.Duration(NormalizeBodyRetentionDays(retentionDays)) * 24 * time.Hour).Format(time.RFC3339Nano)
	versions[versionID] = cloneBodyRecord(body)

	if document, ok := s.docs[owner][bookID]; ok {
		document = normalizeDocument(document)
		document.BodyRefs[versionID] = body.BodyRef
		s.docs[owner][bookID] = cloneDocument(document)
	}
	return body.BodyRef, nil
}

type memoryCleanupCandidate struct {
	bookID       string
	versionID    string
	releasableAt time.Time
}

// CleanupExpiredBodies 跨账号回收到期正文，供后台定时任务使用。
func (s *MemoryStore) CleanupExpiredBodies(ctx context.Context, now time.Time, limit int) (BodyCleanupResult, error) {
	now = normalizeCleanupNow(now)
	s.mu.RLock()
	owners := make([]string, 0, len(s.bodies))
	for owner := range s.bodies {
		owners = append(owners, owner)
	}
	s.mu.RUnlock()
	combined := BodyCleanupResult{Results: make([]BodyCleanupRecord, 0)}
	for _, owner := range owners {
		result, err := s.CleanupBodies(ctx, owner, BodyCleanupRequest{Now: now, Reason: BodyCleanupReasonExpired, Limit: limit})
		if err != nil {
			return combined, err
		}
		combined.Deleted += result.Deleted
		combined.Results = append(combined.Results, result.Results...)
		if len(combined.Results) >= limit {
			break
		}
	}
	return combined, nil
}

func (s *MemoryStore) CleanupBodies(_ context.Context, owner string, request BodyCleanupRequest) (BodyCleanupResult, error) {
	if request.Reason != BodyCleanupReasonExpired && request.Reason != BodyCleanupReasonCapacity {
		return BodyCleanupResult{}, errors.New("invalid body cleanup reason")
	}
	now := normalizeCleanupNow(request.Now)
	limit := normalizeCleanupLimit(request.Limit)

	s.mu.Lock()
	defer s.mu.Unlock()

	candidates := make([]memoryCleanupCandidate, 0)
	for bookID, versions := range s.bodies[owner] {
		for versionID, body := range versions {
			if body.State != "releasable" {
				continue
			}
			releasableAt, ok := parseLifecycleTime(body.ReleasableAt)
			if !ok {
				continue
			}
			if request.Reason == BodyCleanupReasonExpired {
				expiresAt, ok := parseLifecycleTime(body.ExpiresAt)
				if !ok || expiresAt.After(now) {
					continue
				}
			}
			candidates = append(candidates, memoryCleanupCandidate{bookID: bookID, versionID: versionID, releasableAt: releasableAt})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].releasableAt.Equal(candidates[j].releasableAt) {
			if candidates[i].bookID == candidates[j].bookID {
				return candidates[i].versionID < candidates[j].versionID
			}
			return candidates[i].bookID < candidates[j].bookID
		}
		return candidates[i].releasableAt.Before(candidates[j].releasableAt)
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	result := BodyCleanupResult{Results: make([]BodyCleanupRecord, 0, len(candidates))}
	for _, candidate := range candidates {
		versions := s.bodies[owner][candidate.bookID]
		if versions == nil {
			continue
		}
		if _, ok := versions[candidate.versionID]; !ok {
			continue
		}
		delete(versions, candidate.versionID)
		if len(versions) == 0 {
			delete(s.bodies[owner], candidate.bookID)
		}
		if document, ok := s.docs[owner][candidate.bookID]; ok {
			document = normalizeDocument(document)
			delete(document.BodyRefs, candidate.versionID)
			s.docs[owner][candidate.bookID] = cloneDocument(document)
		}
		result.Deleted++
		result.Results = append(result.Results, BodyCleanupRecord{
			BookID:    candidate.bookID,
			VersionID: candidate.versionID,
			Reason:    request.Reason,
		})
	}
	return result, nil
}
