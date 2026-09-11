package store

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// newToken returns a URL-safe random token to hand to the client, and the
// hex-encoded SHA-256 hash of it to store. Only the hash ever touches the
// database, so a stolen database doesn't hand out usable login/session
// tokens.
func newToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// newSlug returns a URL-safe random identifier suitable for share links:
// unguessable, and distinct from login/session tokens (a leaked share link
// must never double as a credential).
func newSlug() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
