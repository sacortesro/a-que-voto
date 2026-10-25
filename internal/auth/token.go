package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateToken(length int) (string, []byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", nil, err
	}

	raw := base64.RawURLEncoding.EncodeToString(bytes)
	hash := sha256.Sum256([]byte(raw))

	return raw, hash[:], nil
}
