package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const passwordCost = 12

// PasswordGenerator hashes and verifies passwords using bcrypt.
type PasswordGenerator struct {
	cost int
}

func NewPasswordGenerator() PasswordGenerator {
	return PasswordGenerator{cost: passwordCost}
}

func (p PasswordGenerator) EncryptPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt password: %w", err)
	}
	return string(hashed), nil
}

func (p PasswordGenerator) ComparePassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}
	return nil
}
