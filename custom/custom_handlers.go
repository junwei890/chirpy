package custom

import (
	"log"
	"net/http"
	"encoding/json"
	"strings"
)

func Readiness(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Println(err)
	}
}

func ValidateChirp(writer http.ResponseWriter, req *http.Request) {
	unallowedWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert": {},
		"fornax": {},
	}
	type responseError struct {
		Error string `json:"error"`
	}
	type validResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}
	type requestBody struct {
		Body string `json:"body"`
	}

	dataReceived := &requestBody{}
	decoderRequestData := json.NewDecoder(req.Body)
	if err := decoderRequestData.Decode(dataReceived); err != nil {
		badRequestResponse := responseError{
			Error: "Something went wrong",
		}
		badRequestResponseInBytes, err := json.Marshal(badRequestResponse)
		if err != nil {
			log.Println(err)
		}
		
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		if _, err := writer.Write(badRequestResponseInBytes); err != nil {
			log.Println(err)
		}
		return
	}
	
	chirp := dataReceived.Body
	if len(chirp) > 140 {
		badRequestResponse := responseError{
			Error: "Chirp is too long",
		}
		badRequestResponseInBytes, err := json.Marshal(badRequestResponse)
		if err != nil {
			log.Println(err)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		if _, err := writer.Write(badRequestResponseInBytes); err != nil {
			log.Println(err)
		}
		return
	}

	chirpSlice := strings.Split(chirp, " ")
	for index, word := range chirpSlice {
		if _, ok := unallowedWords[strings.ToLower(word)]; ok {
			chirpSlice[index] = "****"
		}
	}
	cleanedChirp := strings.Join(chirpSlice, " ")

	responseBody := &validResponse{
		CleanedBody: cleanedChirp,
	}
	responseBodyInBytes, err := json.Marshal(responseBody)
	if err != nil {
		log.Println(err)
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(responseBodyInBytes); err != nil {
		log.Println(err)
	}
}
