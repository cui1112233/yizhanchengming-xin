package authn

import (
	"errors"
	"testing"
)

func TestPasswordHashIsStrongAndOneWay(t *testing.T) {
	const password = "correct horse battery staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == password {
		t.Fatalf("password hash must be non-empty and must not equal plaintext")
	}
	if err := VerifyPassword(hash, password); err != nil {
		t.Fatalf("VerifyPassword(valid): %v", err)
	}
	if err := VerifyPassword(hash, "wrong-password"); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("VerifyPassword(invalid) = %v, want ErrUnauthenticated", err)
	}
}

func TestPasswordHashRejectsEmptyPassword(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("HashPassword(empty) expected error")
	}
}
