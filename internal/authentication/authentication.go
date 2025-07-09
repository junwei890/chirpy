package authentication

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPasswordInBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	hashedPassword := string(hashedPasswordInBytes)
	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) error {
	passwordInBytes := []byte(password)
	hashInBytes := []byte(hash)
	if err := bcrypt.CompareHashAndPassword(passwordInBytes, hashInBytes); err != nil {
		return err
	}
	return nil
}
