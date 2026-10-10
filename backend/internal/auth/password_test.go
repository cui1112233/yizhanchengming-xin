package auth

import (
	"strings"
	"testing"
)

func TestPasswordVerifiesLegacyNodeScryptVector(t *testing.T) {
	const password = "correct horse battery staple"
	const hash = "scrypt$6e6f64652d6c65676163792d73616c74$312324022dc0eb50575164446f6bf3fa80c6bb5f7e02f03492e34d0d93f45fcbc63c66cdb3148743285f6b1fdba2f81c6d3218829e9bbd080d7bec6eed0f2783"

	if !VerifyPassword(password, hash, "scrypt") {
		t.Fatal("legacy Node scrypt password should verify")
	}
	if VerifyPassword("wrong password", hash, "scrypt") {
		t.Fatal("wrong legacy password should fail")
	}
}

func TestNewPasswordIsExplicitAndDoesNotContainPlaintext(t *testing.T) {
	hash, algorithm, err := HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if algorithm != PasswordAlgorithmScryptV1 {
		t.Fatalf("algorithm=%q", algorithm)
	}
	if strings.Contains(hash, "correct horse") {
		t.Fatal("password leaked into hash")
	}
	if !VerifyPassword("correct horse", hash, algorithm) {
		t.Fatal("new password should verify")
	}
	if VerifyPassword("wrong", hash, algorithm) {
		t.Fatal("wrong password should fail")
	}
}
