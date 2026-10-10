package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("auth record not found")
	ErrInvalid  = errors.New("invalid auth input")
)

type Account struct {
	Username          string
	PasswordHash      string
	PasswordAlgorithm string
	IsOwner           bool
	Active            bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Session struct {
	Token     string
	Username  string
	ExpiresAt time.Time
	IssuedAt  time.Time
	Browser   string
	OS        string
	IPHint    string
}

type Store interface {
	FindAccount(context.Context, string) (Account, error)
	CreateAccount(context.Context, Account) error
	UpdatePassword(context.Context, string, string, string) error
	SetActive(context.Context, string, bool) error
	CreateSession(context.Context, Session) error
	FindSession(context.Context, string, time.Time) (Session, error)
	RevokeSession(context.Context, string) error
}
