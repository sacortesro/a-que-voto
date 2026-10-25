package auth_test

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/sacortesro/a-que-voto/internal/auth"
)

func TestGenerateTokenHashMatchesRaw(t *testing.T) {
	raw, hash, err := auth.GenerateToken(32)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	sum := sha256.Sum256([]byte(raw))
	if !bytes.Equal(hash, sum[:]) {
		t.Fatal("hash is not sha256 of the raw token")
	}
}

func TestGenerateTokenIsUnique(t *testing.T) {
	a, _, err := auth.GenerateToken(32)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	b, _, err := auth.GenerateToken(32)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if a == "" || a == b {
		t.Fatalf("tokens not unique: %q %q", a, b)
	}
}
