package authentication

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashedPasswordInBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPasswordInBytes), nil
}

func CheckPasswordHash(password, hash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, secretKey string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer: "chirpy",
		Subject: userID.String(),
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
	}
	createdToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := createdToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, secretKey string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	stringifiedUUID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	returnUUID ,err := uuid.Parse(stringifiedUUID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return returnUUID, nil
}
