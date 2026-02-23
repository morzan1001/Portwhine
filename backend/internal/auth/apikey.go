package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	// apiKeyPrefix is prepended to every generated API key.
	apiKeyPrefix = "pw_"
	// apiKeyRandomBytes is the number of random bytes used for key generation.
	apiKeyRandomBytes = 32
	// apiKeyPrefixLen is the length of the extractable prefix (first 8 chars of raw key).
	apiKeyPrefixLen = 8
)

// GenerateAPIKey creates a new API key. It returns the raw key (to be shown
// to the user once), the SHA-256 hash of the raw key (to be stored), the
// short prefix for lookup, and any error.
func GenerateAPIKey() (rawKey string, keyHash string, prefix string, err error) {
	b := make([]byte, apiKeyRandomBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("generating random bytes: %w", err)
	}

	rawKey = apiKeyPrefix + hex.EncodeToString(b)
	keyHash = HashAPIKey(rawKey)
	prefix = rawKey[:apiKeyPrefixLen]

	return rawKey, keyHash, prefix, nil
}

// HashAPIKey returns the hex-encoded SHA-256 hash of the given raw API key.
func HashAPIKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
