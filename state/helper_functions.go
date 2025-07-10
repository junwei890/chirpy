package state

import (
	"net/http"
	"encoding/json"
	"log"
)

type Error int
const (
	DatabaseError Error = iota
	ServiceError
	BadRequest
	NotFound
	NoContent
	UnauthorizedLogin
	UnauthorizedBadJWT
	UnauthorizedBadRT
	UnauthorizedBadAPIKey
	Forbidden
	LongChirp
)

func ErrorResponseWriter(writer http.ResponseWriter, error Error) {
	type errorResponse struct {
		Error string `json:"error"`	
	}
	var errorMessage string
	var statusCode int

	switch error {
	case DatabaseError, ServiceError:
		errorMessage = "Internal server error, try again"
		statusCode = http.StatusInternalServerError
	case BadRequest:
		errorMessage = "Bad request, try again"
		statusCode = http.StatusBadRequest
	case NotFound:
		errorMessage = "Not Found, try again"
		statusCode = http.StatusNotFound
	case NoContent:
		errorMessage = "Event does not exist"
		statusCode = http.StatusNoContent
	case UnauthorizedLogin:
		errorMessage = "Incorrect email or password"
		statusCode = http.StatusUnauthorized
	case UnauthorizedBadJWT:
		errorMessage = "Invalid JWT"
		statusCode = http.StatusUnauthorized
	case UnauthorizedBadRT:
		errorMessage = "Invalid Refresh token"
		statusCode = http.StatusUnauthorized
	case UnauthorizedBadAPIKey:
		errorMessage = "Invalid API key"
		statusCode = http.StatusUnauthorized
	case Forbidden:
		errorMessage = "You're not allowed to use this endpoint"
		statusCode = http.StatusForbidden
	case LongChirp:
		errorMessage = "Chirp is too long"
		statusCode = http.StatusBadRequest
	}

	errorResponseStruct := &errorResponse{
		Error: errorMessage,
	}
	errorResponseInBytes, err := json.Marshal(errorResponseStruct)
	if err != nil {
		log.Println(err)
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	writer.Write(errorResponseInBytes)
}
