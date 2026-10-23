// Package authtoken creates and verifies compact, signed session tokens
// used for both friend and admin login cookies. A token proves that its
// subject (a friend ID or admin username) was issued by this server and
// has not expired - nothing more.
package authtoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalid = errors.New("authtoken: invalid token")
	ErrExpired = errors.New("authtoken: expired")
)

type payload struct {
	Subject string `json:"sub"`
	Expires int64  `json:"exp"`
}

// New returns a signed token for subject, valid for ttl.
func New(secret []byte, subject string, ttl time.Duration) (string, error) {
	p := payload{Subject: subject, Expires: time.Now().Add(ttl).Unix()}
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(data)
	sig := sign(secret, encoded)
	return encoded + "." + sig, nil
}

// Verify checks the token's signature and expiry and returns its subject.
func Verify(secret []byte, token string) (string, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return "", ErrInvalid
	}
	encoded, sig := parts[0], parts[1]
	want := sign(secret, encoded)
	if subtle.ConstantTimeCompare([]byte(sig), []byte(want)) != 1 {
		return "", ErrInvalid
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrInvalid
	}
	var p payload
	if err := json.Unmarshal(data, &p); err != nil {
		return "", ErrInvalid
	}
	if time.Now().Unix() > p.Expires {
		return "", ErrExpired
	}
	return p.Subject, nil
}

func sign(secret []byte, encoded string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(encoded))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
