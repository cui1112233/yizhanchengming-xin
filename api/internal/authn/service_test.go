package authn

import (
	"context"
	"errors"
	"testing"
)

type fakeIdentityStore struct {
	credential LoginCredential
	err        error
}

func (f fakeIdentityStore) FindLoginCredential(context.Context, string) (LoginCredential, error) {
	return f.credential, f.err
}

func TestServiceLoginIssuesSessionOnlyForValidPassword(t *testing.T) {
	hash, err := HashPassword("s3cure-password")
	if err != nil {
		t.Fatal(err)
	}
	user := User{ID: 7, Username: "alice", DisplayName: "Alice", Role: "member"}
	sessions := newFakeSessionStore(user)
	manager := NewManager(sessions, Options{})
	service := NewService(fakeIdentityStore{credential: LoginCredential{User: user, PasswordHash: hash, Active: true}}, manager)

	issued, gotUser, err := service.Login(context.Background(), "alice", "s3cure-password")
	if err != nil {
		t.Fatalf("Login(valid): %v", err)
	}
	if issued.AccessToken == "" || issued.RefreshToken == "" || gotUser.ID != user.ID {
		t.Fatalf("unexpected login result: %#v %#v", issued, gotUser)
	}

	if _, _, err := service.Login(context.Background(), "alice", "wrong"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("Login(wrong) = %v, want ErrUnauthenticated", err)
	}
}

func TestServiceLoginDoesNotRevealUnknownOrDisabledAccount(t *testing.T) {
	manager := NewManager(newFakeSessionStore(User{ID: 1, Username: "unused"}), Options{})
	for name, store := range map[string]IdentityStore{
		"unknown": fakeIdentityStore{err: ErrUnauthenticated},
		"disabled": fakeIdentityStore{credential: LoginCredential{User: User{ID: 2, Username: "disabled"}, PasswordHash: "$2a$10$invalid", Active: false}},
	} {
		t.Run(name, func(t *testing.T) {
			service := NewService(store, manager)
			if _, _, err := service.Login(context.Background(), name, "password"); !errors.Is(err, ErrUnauthenticated) {
				t.Fatalf("Login = %v, want ErrUnauthenticated", err)
			}
		})
	}
}
