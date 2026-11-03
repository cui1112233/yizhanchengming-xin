package localexecutor

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Store interface {
	CreatePairing(context.Context, PairingRecord) error
	PairExecutor(context.Context, SecretHash, string, ExecutorRecord, time.Time) (ExecutorRecord, error)
	ExecutorByTokenHash(context.Context, SecretHash) (ExecutorRecord, error)
	UpdateHeartbeat(context.Context, string, HeartbeatInput, time.Time) error
	ListExecutors(context.Context, string) ([]ExecutorRecord, error)
}

type Service struct {
	store       Store
	jobs        JobStore
	now         func() time.Time
	jobSignalMu sync.Mutex
	jobSignal   chan struct{}
}

// PendingPairingStore 是 Store 的可选只读扩展。未实现的 Store 仍可正常工作，
// 状态接口只会省略待配对字段，不影响既有配对、心跳和领任务协议。
type PendingPairingStore interface {
	ListPendingPairings(ctx context.Context, owner, platform string, now time.Time) ([]PairingRecord, error)
}

func NewService(store Store, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	jobs, _ := store.(JobStore)
	return &Service{store: store, jobs: jobs, now: now, jobSignal: make(chan struct{})}
}

func (s *Service) currentJobSignal() <-chan struct{} {
	s.jobSignalMu.Lock()
	defer s.jobSignalMu.Unlock()
	return s.jobSignal
}

func (s *Service) notifyJobCreated() {
	s.jobSignalMu.Lock()
	close(s.jobSignal)
	s.jobSignal = make(chan struct{})
	s.jobSignalMu.Unlock()
}

func (s *Service) CreatePairing(ctx context.Context, owner, platform string) (PairingSecret, error) {
	owner = strings.TrimSpace(owner)
	platform = strings.ToLower(strings.TrimSpace(platform))
	if owner == "" {
		return PairingSecret{}, ErrInvalidInput
	}
	if platform != PlatformDoubao {
		return PairingSecret{}, ErrInvalidPlatform
	}
	code, err := randomPairingCode()
	if err != nil {
		return PairingSecret{}, err
	}
	id, err := randomID("lep_", 12)
	if err != nil {
		return PairingSecret{}, err
	}
	now := s.now().UTC()
	expires := now.Add(PairingTTL)
	record := PairingRecord{
		ID: id, OwnerUsername: owner, Platform: platform,
		CodeHash: hashPairingCode(code), ExpiresAt: expires, CreatedAt: now,
	}
	if err := s.store.CreatePairing(ctx, record); err != nil {
		return PairingSecret{}, err
	}
	return PairingSecret{Code: code, ExpiresAt: expires}, nil
}

// PendingPairing 返回该账号最新的一条未消费配对记录，供状态接口在刷新后还原
// "待配对"状态。过期记录由存储层过滤并清理，返回值只含指纹，不含明文配对码。
func (s *Service) PendingPairing(ctx context.Context, owner, platform string) (*PendingPairing, error) {
	owner = strings.TrimSpace(owner)
	platform = strings.ToLower(strings.TrimSpace(platform))
	if owner == "" {
		return nil, ErrInvalidInput
	}
	lister, ok := s.store.(PendingPairingStore)
	if !ok {
		return nil, nil
	}
	records, err := lister.ListPendingPairings(ctx, owner, platform, s.now().UTC())
	if err != nil || len(records) == 0 {
		return nil, err
	}
	newest := records[0]
	return &PendingPairing{
		Platform:    newest.Platform,
		Fingerprint: pairingFingerprint(newest.CodeHash),
		CreatedAt:   newest.CreatedAt,
		ExpiresAt:   newest.ExpiresAt,
		Expired:     !newest.ExpiresAt.After(s.now().UTC()),
	}, nil
}

// pairingFingerprint 只暴露哈希前缀，便于前后端对齐"同一个配对码"而不泄露明文。
func pairingFingerprint(hash SecretHash) string {
	return fmt.Sprintf("%x", hash[:4])
}

func (s *Service) Pair(ctx context.Context, input PairInput) (PairResult, error) {
	platform := strings.ToLower(strings.TrimSpace(input.Platform))
	if platform != PlatformDoubao {
		return PairResult{}, ErrInvalidPlatform
	}
	if normalizePairingCode(input.Code) == "" {
		return PairResult{}, ErrPairingInvalid
	}
	if err := validateDeviceFields(input.DeviceName, input.OS, input.Version); err != nil {
		return PairResult{}, err
	}
	token, err := randomToken()
	if err != nil {
		return PairResult{}, err
	}
	id, err := randomID("lex_", 12)
	if err != nil {
		return PairResult{}, err
	}
	now := s.now().UTC()
	executor := ExecutorRecord{
		ID: id, Platform: platform, TokenHash: hashSecret(token),
		DeviceName: strings.TrimSpace(input.DeviceName), OS: strings.TrimSpace(input.OS), Version: strings.TrimSpace(input.Version),
		CreatedAt: now, UpdatedAt: now,
	}
	created, err := s.store.PairExecutor(ctx, hashPairingCode(input.Code), platform, executor, now)
	if err != nil {
		return PairResult{}, err
	}
	return PairResult{ExecutorID: created.ID, Token: token, HeartbeatIntervalSeconds: HeartbeatIntervalSeconds}, nil
}

func (s *Service) Heartbeat(ctx context.Context, token string, input HeartbeatInput) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrExecutorUnauthorized
	}
	if err := validateAccountStats(input.Accounts); err != nil {
		return err
	}
	if err := validateDeviceFields(input.DeviceName, input.OS, input.Version); err != nil {
		return err
	}
	record, err := s.store.ExecutorByTokenHash(ctx, hashSecret(token))
	if err != nil {
		return ErrExecutorUnauthorized
	}
	input.DeviceName = strings.TrimSpace(input.DeviceName)
	input.OS = strings.TrimSpace(input.OS)
	input.Version = strings.TrimSpace(input.Version)
	if err := s.store.UpdateHeartbeat(ctx, record.ID, input, s.now().UTC()); err != nil {
		return err
	}
	return nil
}

func (s *Service) List(ctx context.Context, owner string) ([]ExecutorView, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return nil, ErrInvalidInput
	}
	records, err := s.store.ListExecutors(ctx, owner)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	views := make([]ExecutorView, 0, len(records))
	seen := make(map[string]bool)
	for _, item := range records {
		key := item.Platform + "\x00" + item.DeviceName + "\x00" + item.OS
		if seen[key] {
			continue
		}
		seen[key] = true
		online := item.LastSeenAt != nil && !item.LastSeenAt.Before(now.Add(-OnlineThreshold))
		views = append(views, ExecutorView{
			ID: item.ID, Name: item.DeviceName, Platform: item.Platform, OS: item.OS, Version: item.Version,
			Online: online, LastSeenAt: item.LastSeenAt, Accounts: item.Accounts,
		})
	}
	return views, nil
}

func validateAccountStats(v AccountStats) error {
	values := []int{v.Total, v.Available, v.Busy, v.QuotaExhausted, v.LoginError, v.HumanVerification}
	for _, n := range values {
		if n < 0 || n > v.Total {
			return ErrInvalidAccountStats
		}
	}
	if v.Available+v.Busy+v.QuotaExhausted+v.LoginError+v.HumanVerification > v.Total {
		return ErrInvalidAccountStats
	}
	return nil
}

func validateDeviceFields(name, osName, version string) error {
	if len(strings.TrimSpace(name)) > 191 || len(strings.TrimSpace(osName)) > 32 || len(strings.TrimSpace(version)) > 64 {
		return ErrInvalidInput
	}
	return nil
}

func normalizePairingCode(code string) string {
	replacer := strings.NewReplacer("-", "", " ", "", "_", "")
	return strings.ToUpper(replacer.Replace(strings.TrimSpace(code)))
}

func hashPairingCode(code string) SecretHash {
	return sha256.Sum256([]byte(normalizePairingCode(code)))
}
func hashSecret(secret string) SecretHash { return sha256.Sum256([]byte(secret)) }

const pairingAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomPairingCode() (string, error) {
	raw := make([]byte, 10)
	entropy := make([]byte, 10)
	if _, err := rand.Read(entropy); err != nil {
		return "", err
	}
	for i, b := range entropy {
		raw[i] = pairingAlphabet[int(b)%len(pairingAlphabet)]
	}
	return string(raw[:5]) + "-" + string(raw[5:]), nil
}

func randomToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func randomID(prefix string, bytes int) (string, error) {
	raw := make([]byte, bytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + base64.RawURLEncoding.EncodeToString(raw), nil
}

func secretHashKey(hash SecretHash) string { return fmt.Sprintf("%x", hash[:]) }
