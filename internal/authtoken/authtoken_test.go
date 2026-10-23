package authtoken

import (
	"testing"
	"time"
)

func TestNewAndVerify(t *testing.T) {
	secret := []byte("test-secret")
	token, err := New(secret, "friend-1", time.Hour)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	subject, err := Verify(secret, token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if subject != "friend-1" {
		t.Errorf("subject = %q, want friend-1", subject)
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	token, _ := New([]byte("secret-a"), "friend-1", time.Hour)
	if _, err := Verify([]byte("secret-b"), token); err != ErrInvalid {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
}

func TestVerify_Expired(t *testing.T) {
	secret := []byte("test-secret")
	token, _ := New(secret, "friend-1", -time.Hour)
	if _, err := Verify(secret, token); err != ErrExpired {
		t.Errorf("err = %v, want ErrExpired", err)
	}
}

func TestVerify_Malformed(t *testing.T) {
	secret := []byte("test-secret")
	for _, bad := range []string{"", "no-dot", "a.b.c", "!!!.!!!"} {
		if _, err := Verify(secret, bad); err == nil {
			t.Errorf("Verify(%q) = nil error, want error", bad)
		}
	}
}
