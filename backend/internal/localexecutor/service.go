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
