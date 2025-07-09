package state

import (
	"sync/atomic"
	"net/http"
	"fmt"
	"io"
	"time"
	"encoding/json"
	"strings"
	"github.com/junwei890/chirpy/internal/database"
	"github.com/junwei890/chirpy/internal/auth"
	"github.com/google/uuid"
)

type APIConfig struct {
	FileServerHits atomic.Int32
	PtrToQueries *database.Queries
	Platform string
	Profanities map[string]struct{}
}


func GetReadiness(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}

func (a *APIConfig) MiddlewareMetricsInc(toHandle http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		a.FileServerHits.Add(1)
		toHandle.ServeHTTP(writer, req)
	})
}

func (a *APIConfig) GetMetrics(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	serverHits := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, a.FileServerHits.Load())
	if _, err := writer.Write([]byte(serverHits)); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}

func (a *APIConfig) PostMetrics(writer http.ResponseWriter, req *http.Request) {
	a.FileServerHits.Store(0)
	if a.Platform != "dev" {
		ErrorResponseWriter(writer, Forbidden)
		return
	}
	if err := a.PtrToQueries.DeleteUsers(req.Context()); err != nil {
		ErrorResponseWriter(writer, DatabaseError)
	}
}

func (a *APIConfig) PostUsers(writer http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	type validResponse struct {
		ID uuid.UUID `json:"id"`
		Email string `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	
	dataReceivedInBytes, err := io.ReadAll(req.Body)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	dataReceived := &requestBody{}
	if err := json.Unmarshal(dataReceivedInBytes, dataReceived); err != nil {
		ErrorResponseWriter(writer, BadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(dataReceived.Password)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}

	createUserParams := database.CreateUserParams{
		Email: dataReceived.Email,
		HashedPassword: hashedPassword,
	}
	userCreationDetails, err := a.PtrToQueries.CreateUser(req.Context(), createUserParams)
	if err != nil {
		ErrorResponseWriter(writer, DatabaseError)
		return
	}
	formattedUserCreationDetails := validResponse{
		ID: userCreationDetails.ID,
		Email: dataReceived.Email,
		CreatedAt: userCreationDetails.CreatedAt,
		UpdatedAt: userCreationDetails.UpdatedAt,
	}

	userCreationDetailsInBytes, err := json.Marshal(formattedUserCreationDetails)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	if _, err := writer.Write(userCreationDetailsInBytes); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}

func (a *APIConfig) PostChirps(writer http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	type validResponse struct {
		ID uuid.UUID `json:"id"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	dataReceivedInBytes, err := io.ReadAll(req.Body)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
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
		Body: chirp,
		UserID: createdChirp.UserID,
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
	}

	chirpCreationDetailsInBytes, err := json.Marshal(formattedChirpCreationDetails)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	if _, err := writer.Write(chirpCreationDetailsInBytes); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}

func (a *APIConfig) GetChirps(writer http.ResponseWriter, req *http.Request) {
	type oneChirp struct {
		ID uuid.UUID `json:"id"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	sliceOfAllChirps, err := a.PtrToQueries.GetAllChirps(req.Context())
	if err != nil {
		ErrorResponseWriter(writer, DatabaseError)
		return
	}
	var sliceOfFormattedChirps []oneChirp
	for _, chirp := range sliceOfAllChirps {
		formattedChirp := oneChirp{
			ID: chirp.ID,
			Body: chirp.Body,
			UserID: chirp.UserID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		}
		sliceOfFormattedChirps = append(sliceOfFormattedChirps, formattedChirp)
	}
	
	allChirpsInBytes, err := json.Marshal(sliceOfFormattedChirps)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(allChirpsInBytes); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}

func (a *APIConfig) GetChirp(writer http.ResponseWriter, req *http.Request) {
	type validResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	chirpID := req.PathValue("chirpID")
	if chirpID == "" {
		ErrorResponseWriter(writer, NotFound)
		return
	}
	castedChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		ErrorResponseWriter(writer, NotFound)
		return
	}
	chirpToGet, err := a.PtrToQueries.GetOneChirp(req.Context(), castedChirpID)
	if err != nil {
		ErrorResponseWriter(writer, NotFound)
		return
	}

	formattedChirpToGet := validResponse{
		ID: chirpToGet.ID,
		CreatedAt: chirpToGet.CreatedAt,
		UpdatedAt: chirpToGet.UpdatedAt,
		Body: chirpToGet.Body,
		UserID: chirpToGet.UserID,
	}
	chirpToGetInBytes, err := json.Marshal(formattedChirpToGet)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(chirpToGetInBytes); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}

func (a *APIConfig) PostLogin(writer http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}
	type validResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
	}

	dataReceivedInBytes, err := io.ReadAll(req.Body)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	dataReceived := &requestBody{}
	if err := json.Unmarshal(dataReceivedInBytes, dataReceived); err != nil {
		ErrorResponseWriter(writer, BadRequest)
		return
	}

	userDetails, err := a.PtrToQueries.GetUserByEmail(req.Context(), dataReceived.Email)
	if err != nil {
		ErrorResponseWriter(writer, Unauthorized)
		return
	}
	if err := auth.CheckPasswordHash(userDetails.HashedPassword, dataReceived.Password); err != nil {
		ErrorResponseWriter(writer, Unauthorized)
		return
	}

	formattedUserDetails := validResponse{
		ID: userDetails.ID,
		CreatedAt: userDetails.CreatedAt,
		UpdatedAt: userDetails.UpdatedAt,
		Email: userDetails.Email,
	}
	userDetailsInBytes, err := json.Marshal(formattedUserDetails)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(userDetailsInBytes); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}
