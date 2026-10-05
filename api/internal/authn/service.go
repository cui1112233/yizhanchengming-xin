package authn

import (
	"context"
	"strings"
)

type LoginCredential struct {
	User         User
	PasswordHash string
	Active       bool
}

type IdentityStore interface {
	FindLoginCredential(ctx context.Context, username string) (LoginCredential, error)
}

type Service struct {
	identities IdentityStore
	sessions   *Manager
}

func NewService(identities IdentityStore, sessions *Manager) *Service {
	return &Service{identities: identities, sessions: sessions}
}

func (s *Service) Login(ctx context.Context, username, password string) (Credentials, User, error) {
	username = strings.TrimSpace(username)
	if s == nil || s.identities == nil || s.sessions == nil || username == "" || password == "" {
		return Credentials{}, User{}, ErrUnauthenticated
	}
	credential, err := s.identities.FindLoginCredential(ctx, username)
	if err != nil || !credential.Active || credential.User.ID <= 0 {
		return Credentials{}, User{}, ErrUnauthenticated
	}
	if err := VerifyPassword(credential.PasswordHash, password); err != nil {
		return Credentials{}, User{}, ErrUnauthenticated
	}
	issued, err := s.sessions.Issue(ctx, credential.User.ID)
	if err != nil {
		return Credentials{}, User{}, err
	}
	return issued, credential.User, nil
}

func (s *Service) AuthenticateAccess(ctx context.Context, token string) (User, error) {
	if s == nil || s.sessions == nil {
		return User{}, ErrUnauthenticated
	}
	return s.sessions.AuthenticateAccess(ctx, token)
}

func (s *Service) Refresh(ctx context.Context, token string) (Credentials, User, error) {
	if s == nil || s.sessions == nil {
		return Credentials{}, User{}, ErrUnauthenticated
	}
	return s.sessions.Refresh(ctx, token)
}

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	if s == nil || s.sessions == nil {
		return nil
	}
	return s.sessions.Logout(ctx, accessToken, refreshToken)
}
