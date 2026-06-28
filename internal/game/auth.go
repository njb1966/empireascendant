package game

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func CheckPassword(hash, password string) bool {
	return hash == HashPassword(password)
}
