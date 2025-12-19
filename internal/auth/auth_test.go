package auth

import "testing"

func TestAuthenticate(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	a := New(hash)
	if !a.Authenticate("secret") {
		t.Fatalf("expected password to authenticate")
	}
	if a.Authenticate("wrong") {
		t.Fatalf("expected authentication to fail for wrong password")
	}
}

func TestAuthenticateEmptyHash(t *testing.T) {
	a := New("")
	if a.Authenticate("anything") {
		t.Fatalf("expected empty hash to fail authentication")
	}
}
