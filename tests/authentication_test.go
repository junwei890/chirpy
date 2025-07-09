package tests

import (
	"testing"
	"time"
	"github.com/google/uuid"
	"github.com/junwei890/chirpy/internal/authentication"
)

func TestHashPassword(t *testing.T) {
	password1 := "helloworld"
	password2 := "blazinglyfast"
	hashedPassword1, _ := authentication.HashPassword(password1)
	hashedPassword2, _ := authentication.HashPassword(password2)

	testCases := []struct {
		name string
		password string
		hash string
		errorPresent bool
	}{
		{
			name: "Hash match",
			password: password1,
			hash: hashedPassword1,
			errorPresent: false,
		},
		{
			name: "Hash mismatch",
			password: password1,
			hash: hashedPassword2,
			errorPresent: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := authentication.CheckPasswordHash(testCase.password, testCase.hash)
			if (err != nil) != testCase.errorPresent {
				t.Errorf("Test case: %s, failed.", testCase.name)
			}
		})
	}
}

func TestMakingAndValidatingJWT(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	secretKey1 := "helloworld"
	secretKey2 := "blazinglyfast"
	duration1 := time.Duration(60000000000)
	duration2 := time.Duration(60)
	jwtTokenUserID1, _ := authentication.MakeJWT(userID1, secretKey1, duration1)
	jwtTokenUserID2, _ := authentication.MakeJWT(userID2, secretKey2, duration2)

	testCases := []struct {
		name string
		userID uuid.UUID
		jwtToken string
		secretKey string
		errorPresent bool
	}{
		{
			name: "Token is valid and returned UUID is the same",
			userID: userID1,
			jwtToken: jwtTokenUserID1,
			secretKey: secretKey1,
			errorPresent: false,
		},
		{
			name: "Token is not valid",
			userID: userID1,
			jwtToken: jwtTokenUserID1,
			secretKey: secretKey2,
			errorPresent: true,
		},
		{
			name: "Token has expired",
			userID: userID2,
			jwtToken: jwtTokenUserID2,
			secretKey: secretKey2,
			errorPresent: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			returnedUserID, err := authentication.ValidateJWT(testCase.jwtToken, testCase.secretKey)
			if (err != nil) != testCase.errorPresent {
				t.Errorf("Test case: %s, failed.", testCase.name)
			}
			if testCase.name == "Token is valid and returned UUID is the same" && testCase.userID != returnedUserID {
				t.Errorf("Test case: %s, failed. %v != %v", testCase.name, testCase.userID, returnedUserID)
			}
		})
	}
}
