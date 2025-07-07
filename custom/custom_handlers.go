package custom

import (
	"log"
	"net/http"
	"encoding/json"
)

// Custom handler to check if server is ready for requests, writing how we will respond to requests
func Readiness(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Println(err)
	}
}

func ValidateChirp(writer http.ResponseWriter, req *http.Request) {
	type validResponse struct {
		Valid bool `json:"valid"`
	}
	type responseError struct {
		Error string `json:"error"`
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
		badRequestResponseJSON, err := json.Marshal(badRequestResponse)
		if err != nil {
			log.Println(err)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		if _, err := writer.Write(badRequestResponseJSON); err != nil {
			log.Println(err)
		}
		return
	}

	if len(dataReceived.Body) > 140 {
		badRequestResponse := responseError{
			Error: "Chirp is too long",
		}
		badRequestResponseJSON, err := json.Marshal(badRequestResponse)
		if err != nil {
			log.Println(err)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		if _, err := writer.Write(badRequestResponseJSON); err != nil {
			log.Println(err)
		}
		return
	}

	statusOKResponse := validResponse{
		Valid: true,
	}
	statusOKResponseJSON, err := json.Marshal(statusOKResponse)
	if err != nil {
		log.Println(err)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(statusOKResponseJSON); err != nil {
		log.Println(err)
	}
}
