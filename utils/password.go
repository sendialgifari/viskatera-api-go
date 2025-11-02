package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MustHashPlaceholder returns a valid hash for an unusable random password.
func MustHashPlaceholder() string {
	h, err := HashPassword("_unusable_" + "placeholder_password")
	if err != nil {
		return ""
	}
	return h
}
