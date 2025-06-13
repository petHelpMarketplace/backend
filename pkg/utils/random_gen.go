package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// Generate random set length string
func GenerateRandomKey(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("random_key_gen: failed to generate random key with lengh %d: %w", length, err)
	}
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}
