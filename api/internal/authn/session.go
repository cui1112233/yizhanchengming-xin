package authn

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"time"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type User struct {
	ID           int64    `json:"id"`
	Username     string   `json:"username"`
	DisplayName  string   `json:"name"`
	Role         string   `json:"role"`
	TeamID       int64    `json:"teamId,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type SessionRecord struct {
	UserID           int64      `json:"userId,omitempty"`
	AccessTokenHash  string     `json:"-"`
	RefreshTokenHash string     `json:"-"`
	AccessExpiresAt  time.Time  `json:"accessExpiresAt,omitempty"`
	RefreshExpiresAt time.Time  `json:"refreshExpiresAt,omitempty"`
	RevokedAt        *time.Time `json:"revokedAt,omitempty"`
}

type SessionStore interface {
	CreateSession(ctx context.Context, record SessionRecord) error
	ResolveAccess(ctx context.Context, accessHash string, now time.Time) (User, error)
	RotateByRefresh(ctx context.Context, oldRefreshHash string, next SessionRecord, now time.Time) (User, error)
	RevokeByAccess(ctx context.Context, accessHash string, now time.Time) error
	RevokeByRefresh(ctx context.Context, refreshHash string, now time.Time) error
}

type Options struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Now        func() time.Time
	Random     io.Reader
}

type Credentials struct {
	AccessToken  string `json:"-"`
	RefreshToken string `json:"-"`
}

type Manager struct {
	store      SessionStore
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
	random     io.Reader
}

func NewManager(store SessionStore, options Options) *Manager {
	accessTTL := options.AccessTTL
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	refreshTTL := options.RefreshTTL
	if refreshTTL <= 0 {
		refreshTTL = 30 * 24 * time.Hour
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	random := options.Random
	if random == nil {
		random = rand.Reader
	}
	return &Manager{store: store, accessTTL: accessTTL, refreshTTL: refreshTTL, now: now, random: random}
}

func (m *Manager) Issue(ctx context.Context, userID int64) (Credentials, error) {
	if m == nil || m.store == nil || userID <= 0 {
		return Credentials{}, ErrUnauthenticated
	}
	credentials, record, err := m.newSessionRecord(userID)
	if err != nil {
		return Credentials{}, err
	}
	if err := m.store.CreateSession(ctx, record); err != nil {
		return Credentials{}, err
	}
	return credentials, nil
}

func (m *Manager) AuthenticateAccess(ctx context.Context, token string) (User, error) {
	if m == nil || m.store == nil || token == "" {
		return User{}, ErrUnauthenticated
	}
	return m.store.ResolveAccess(ctx, hashToken(token), m.now().UTC())
}

func (m *Manager) Refresh(ctx context.Context, refreshToken string) (Credentials, User, error) {
	if m == nil || m.store == nil || refreshToken == "" {
		return Credentials{}, User{}, ErrUnauthenticated
	}
	credentials, record, err := m.newSessionRecord(0)
	if err != nil {
		return Credentials{}, User{}, err
	}
	user, err := m.store.RotateByRefresh(ctx, hashToken(refreshToken), record, m.now().UTC())
	if err != nil {
		return Credentials{}, User{}, err
	}
	return credentials, user, nil
}

func (m *Manager) Logout(ctx context.Context, accessToken, refreshToken string) error {
	if m == nil || m.store == nil {
		return nil
	}
	now := m.now().UTC()
	if accessToken != "" {
		if err := m.store.RevokeByAccess(ctx, hashToken(accessToken), now); err != nil && !errors.Is(err, ErrUnauthenticated) {
			return err
		}
	}
	if refreshToken != "" {
		if err := m.store.RevokeByRefresh(ctx, hashToken(refreshToken), now); err != nil && !errors.Is(err, ErrUnauthenticated) {
			return err
		}
	}
	return nil
}

func (m *Manager) newSessionRecord(userID int64) (Credentials, SessionRecord, error) {
	accessToken, err := randomToken(m.random, 32)
	if err != nil {
		return Credentials{}, SessionRecord{}, err
	}
	refreshToken, err := randomToken(m.random, 48)
	if err != nil {
		return Credentials{}, SessionRecord{}, err
	}
	now := m.now().UTC()
	return Credentials{AccessToken: accessToken, RefreshToken: refreshToken}, SessionRecord{
		UserID:           userID,
		AccessTokenHash:  hashToken(accessToken),
		RefreshTokenHash: hashToken(refreshToken),
		AccessExpiresAt:  now.Add(m.accessTTL),
		RefreshExpiresAt: now.Add(m.refreshTTL),
	}, nil
}

func randomToken(reader io.Reader, size int) (string, error) {
	value := make([]byte, size)
	if _, err := io.ReadFull(reader, value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashToken(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
