package auth

import (
	"time"
	"net/http"
	"errors"
	"strings"
	"crypto/rand"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashedPasswordInBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPasswordInBytes), nil
}

func CheckPasswordHash(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, secretKey string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject: userID.String(),
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

	jwtUUID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	jwtIssuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.UUID{}, err
	}

	if jwtIssuer != "chirpy" {
		return uuid.UUID{}, jwt.ErrTokenInvalidIssuer
	}

	returnUUID, err := uuid.Parse(jwtUUID)
	if err != nil {
		return uuid.UUID{}, jwt.ErrTokenInvalidSubject
	}
	return returnUUID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authInfo := headers.Get("Authorization")
	if authInfo == "" {
		return "", errors.New("authorization header does not exist")
	}

	authInfoSlice := strings.Fields(authInfo)
	if len(authInfoSlice) < 2 {
		return "", errors.New("token not present in authorization header")
	}
	return authInfoSlice[1], nil
}

func MakeRefreshToken() (string, error) {
	sliceToRead := make([]byte, 32)
	rand.Read(sliceToRead)
	bitEncodedString := hex.EncodeToString(sliceToRead)
	return bitEncodedString, nil
}
