package tests

import (
	"testing"
	"time"
	"net/http"
	"github.com/google/uuid"
	"github.com/junwei890/chirpy/internal/auth"
)

func TestHashPassword(t *testing.T) {
	password1 := "helloworld"
	password2 := "blazinglyfast"
	hashedPassword1, _ := auth.HashPassword(password1)
	hashedPassword2, _ := auth.HashPassword(password2)

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
			err := auth.CheckPasswordHash(testCase.hash, testCase.password)
			if (err != nil) != testCase.errorPresent {
				t.Errorf("test case: %s, failed.", testCase.name)
			}
		})
	}
}

func TestMakingAndValidatingJWT(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()
	secretKey1 := "helloworld"
	secretKey2 := "blazinglyfast"
	duration1 := time.Duration(60) * time.Second
	duration2 := time.Duration(1)
	jwtTokenUserID1, _ := auth.MakeJWT(userID1, secretKey1, duration1)
	jwtTokenUserID2, _ := auth.MakeJWT(userID2, secretKey2, duration2)

	testCases := []struct {
		name string
		userID uuid.UUID
		jwtToken string
		secretKey string
		expected uuid.UUID
		errorPresent bool
	}{
		{
			name: "Token is valid and returned UUID is the same",
			userID: userID1,
			jwtToken: jwtTokenUserID1,
			secretKey: secretKey1,
			expected: userID1,
			errorPresent: false,
		},
		{
			name: "Token is not valid",
			userID: userID1,
			jwtToken: jwtTokenUserID1,
			secretKey: secretKey2,
			expected: uuid.UUID{},
			errorPresent: true,
		},
		{
			name: "Token has expired",
			userID: userID2,
			jwtToken: jwtTokenUserID2,
			secretKey: secretKey2,
			expected: uuid.UUID{},
			errorPresent: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			returnedUserID, err := auth.ValidateJWT(testCase.jwtToken, testCase.secretKey)
			if (err != nil) != testCase.errorPresent && testCase.expected != returnedUserID {
				t.Errorf("Test case: %s, failed.", testCase.name)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	dummyHeader1 := http.Header{}
	dummyHeader2 := http.Header{}
	dummyHeader3 := http.Header{}
	dummyHeader1.Set("Authorization", "Bearer tokenString")
	dummyHeader2.Set("Authorization", "")
	
	testCases := []struct {
		name string
		header http.Header
		expected string
		errorPresent bool
	}{
		{
			name: "Token string present and correct",
			header: dummyHeader1,
			expected: "tokenString",
			errorPresent: false,
		},
		{
			name: "Token string not present",
			header: dummyHeader2,
			expected: "",
			errorPresent: true,
		},
		{
			name: "Authorization header not present",
			header: dummyHeader3,
			expected: "",
			errorPresent: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tokenString, err := auth.GetBearerToken(testCase.header)
			if (err != nil) != testCase.errorPresent && tokenString != testCase.expected {
				t.Errorf("test case: %s, failed.", testCase.name)
			}
		})
	}
}
