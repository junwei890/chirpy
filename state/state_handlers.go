package state

import (
	"sync/atomic"
	"net/http"
	"fmt"
	"log"
	"io"
	"encoding/json"
	"time"
	"strings"
	"github.com/junwei890/chirpy/internal/database"
	"github.com/google/uuid"
)

type APIConfig struct {
	FileServerHits atomic.Int32
	PtrToQueries *database.Queries
	Platform string
	Profanities map[string]struct{}
}


func Readiness(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Println(err)
	}
}

func (a *APIConfig) MiddlewareMetricsInc(toHandle http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		a.FileServerHits.Add(1)
		toHandle.ServeHTTP(writer, req)
	})
}

func (a *APIConfig) Metrics(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	serverHits := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, a.FileServerHits.Load())
	if _, err := writer.Write([]byte(serverHits)); err != nil {
		log.Println(err)
	}
}

func (a *APIConfig) Reset(writer http.ResponseWriter, req *http.Request) {
	a.FileServerHits.Store(0)
	if a.Platform != "dev" {
		ErrorResponseWriter(writer, Forbidden)
	}
	if err := a.PtrToQueries.DeleteUsers(req.Context()); err != nil {
		ErrorResponseWriter(writer, DatabaseError)
	}
}

func (a *APIConfig) NewUser(writer http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Email string `json:"email"`
	}
	type validResponse struct {
		ID uuid.UUID `json:"id"`
		Email string `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	
	dataReceivedInBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
	}
	dataReceived := &requestBody{}
	if err := json.Unmarshal(dataReceivedInBytes, dataReceived); err != nil {
		ErrorResponseWriter(writer, BadRequest)
		return
	}

	userCreationDetails, err := a.PtrToQueries.CreateUser(req.Context(), dataReceived.Email)
	if err != nil {
		ErrorResponseWriter(writer, DatabaseError)
	}
	formattedUserCreationDetails := validResponse{
		ID: userCreationDetails.ID,
		Email: dataReceived.Email,
		CreatedAt: userCreationDetails.CreatedAt,
		UpdatedAt: userCreationDetails.UpdatedAt,
	}
	userCreationDetailsInBytes, err := json.Marshal(formattedUserCreationDetails)
	if err != nil {
		log.Println(err)
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	if _, err := writer.Write(userCreationDetailsInBytes); err != nil {
		log.Println(err)
	}
}

func (a *APIConfig) NewChirp(writer http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	type validResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	dataReceivedInBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
	}
	dataReceived := &requestBody{}
	if err := json.Unmarshal(dataReceivedInBytes, dataReceived); err != nil {
		ErrorResponseWriter(writer, BadRequest)
		return
	}

	if len(dataReceived.Body) > 140 {
		ErrorResponseWriter(writer, LongChirp)
		return
	}


	chirpInSlice := strings.Split(dataReceived.Body, " ")
	for index, word := range chirpInSlice {
		if _, ok := a.Profanities[word]; ok {
			chirpInSlice[index] = "****"
		}
	}
	chirp := strings.Join(chirpInSlice, " ")

	createChirpParams := database.CreateChirpParams{
		Body: chirp,
		UserID: dataReceived.UserID,
	}
	createdChirp, err := a.PtrToQueries.CreateChirp(req.Context(), createChirpParams)
	if err != nil {
		ErrorResponseWriter(writer, DatabaseError)
		return
	}

	formattedChirpCreationDetails := validResponse{
		ID: createdChirp.ID,
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
		Body: createdChirp.Body,
		UserID: createdChirp.UserID,
	}
	chirpCreationDetailsInBytes, err := json.Marshal(formattedChirpCreationDetails)
	if err != nil {
		log.Println(err)
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	writer.Write(chirpCreationDetailsInBytes)
}
