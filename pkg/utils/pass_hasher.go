package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashGen generates a bcrypt hash from a plain string.
func HashGen(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// HashCompare compares a bcrypt hash with a plain string.
func HashCompare(hash, plain string) error {

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
