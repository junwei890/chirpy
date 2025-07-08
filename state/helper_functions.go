package state

import (
	"net/http"
	"encoding/json"
	"log"
)

type Error int
const (
	DatabaseError Error = iota
	BadRequest
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
	case DatabaseError:
		errorMessage = "Internal database error, try again"
		statusCode = http.StatusInternalServerError
	case BadRequest:
		errorMessage = "Bad request, try again"
		statusCode = http.StatusBadRequest
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
