package localexecutor

import (
	"errors"
	"time"
)

var (
	ErrInvalidPlatform      = errors.New("invalid executor platform")
	ErrInvalidInput         = errors.New("invalid executor input")
	ErrPairingInvalid       = errors.New("pairing code is invalid or expired")
	ErrExecutorUnauthorized = errors.New("executor is unauthorized")
	ErrInvalidAccountStats  = errors.New("invalid executor account stats")
)

const (
	PlatformDoubao           = "doubao"
	PairingTTL               = 10 * time.Minute
	HeartbeatIntervalSeconds = 15
	OnlineThreshold          = 45 * time.Second
	// pairingCleanupRetention 是过期配对记录的保留时长。过期配对码本身已不可用，
	// 超过保留期后从现有 local_executor_pairings 表清理，不新增表。
	pairingCleanupRetention = time.Hour
)

type SecretHash [32]byte

type PairingRecord struct {
	ID            string
	OwnerUsername string
	Platform      string
	CodeHash      SecretHash
	ExpiresAt     time.Time
	ConsumedAt    *time.Time
	CreatedAt     time.Time
}

type PairingSecret struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// PendingPairing 是"已生成配对码但还没有设备消费"的只读视图，
// 只暴露指纹与时间，不回传明文配对码或哈希。
type PendingPairing struct {
	Platform    string    `json:"platform"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Expired     bool      `json:"expired"`
}

type PairInput struct {
	Code       string `json:"code"`
	DeviceName string `json:"deviceName"`
	Platform   string `json:"platform"`
	OS         string `json:"os"`
	Version    string `json:"version"`
}

type PairResult struct {
	ExecutorID               string `json:"executorId"`
	Token                    string `json:"token"`
	HeartbeatIntervalSeconds int    `json:"heartbeatIntervalSeconds"`
}

type AccountStats struct {
	Total             int `json:"total"`
	Available         int `json:"available"`
	Busy              int `json:"busy"`
	QuotaExhausted    int `json:"quotaExhausted"`
	LoginError        int `json:"loginError"`
	HumanVerification int `json:"humanVerification"`
}

type HeartbeatInput struct {
	DeviceName string       `json:"deviceName"`
	OS         string       `json:"os"`
	Version    string       `json:"version"`
	Accounts   AccountStats `json:"accounts"`
}

type ExecutorRecord struct {
	ID            string
	OwnerUsername string
	Platform      string
	TokenHash     SecretHash
	DeviceName    string
	OS            string
	Version       string
	Accounts      AccountStats
	LastSeenAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ExecutorView struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Platform   string       `json:"platform"`
	OS         string       `json:"os"`
	Version    string       `json:"version"`
	Online     bool         `json:"online"`
	LastSeenAt *time.Time   `json:"lastSeenAt,omitempty"`
	Accounts   AccountStats `json:"accounts"`
}
