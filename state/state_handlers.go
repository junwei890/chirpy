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
	SecretKey string
	WebhookKey string
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
		ChirpyRed bool `json:"is_chirpy_red"`
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
		ChirpyRed: userCreationDetails.IsChirpyRed,
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
	}
	type validResponse struct {
		ID uuid.UUID `json:"id"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
		ChirpyRed bool `json:"is_chirpy_red"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	profanities := map[string]struct{}{
		"kerfuffle": {},
		"sharbert": {},
		"fornax": {},
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

	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadJWT)
		return
	}
	userID, err := auth.ValidateJWT(bearerToken, a.SecretKey)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadJWT)
		return
	}

	if len(dataReceived.Body) > 140 {
		ErrorResponseWriter(writer, LongChirp)
		return
	}
	chirpInSlice := strings.Split(dataReceived.Body, " ")
	for index, word := range chirpInSlice {
		if _, ok := profanities[word]; ok {
			chirpInSlice[index] = "****"
		}
	}
	chirp := strings.Join(chirpInSlice, " ")

	createChirpParams := database.CreateChirpParams{
		Body: chirp,
		UserID: userID,
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
		ChirpyRed: createdChirp.IsChirpyRed,
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
		ChirpyRed bool `json:"is_chirpy_red"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	authorID := req.URL.Query().Get("author_id")
	parsedAuthorID, err := uuid.Parse(authorID)
	if err != nil {
		parsedAuthorID = uuid.Nil
	}

	sliceOfAllChirps, err := a.PtrToQueries.GetAllChirps(req.Context())
	if err != nil {
		ErrorResponseWriter(writer, DatabaseError)
		return
	}
	returnChirps := []oneChirp{}
	if parsedAuthorID != uuid.Nil {
		for _, chirp := range sliceOfAllChirps {
			if parsedAuthorID == chirp.UserID {
				formattedChirp := oneChirp{
					ID: chirp.ID,
					Body: chirp.Body,
					UserID: chirp.UserID,
					ChirpyRed: chirp.IsChirpyRed,
					CreatedAt: chirp.CreatedAt,
					UpdatedAt: chirp.UpdatedAt,
				}
				returnChirps = append(returnChirps, formattedChirp)
			}
		}
	} else {	
		for _, chirp := range sliceOfAllChirps {
			formattedChirp := oneChirp{
				ID: chirp.ID,
				Body: chirp.Body,
				UserID: chirp.UserID,
				ChirpyRed: chirp.IsChirpyRed,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
			}
			returnChirps = append(returnChirps, formattedChirp)
		}
	}
	chirpsInBytes, err := json.Marshal(returnChirps)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(chirpsInBytes); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}

func (a *APIConfig) GetChirpsByID(writer http.ResponseWriter, req *http.Request) {
	type validResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
		ChirpyRed bool `json:"is_chirpy_red"`
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
		ChirpyRed: chirpToGet.IsChirpyRed,
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
		ChirpyRed bool `json:"is_chirpy_red"`
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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
		ErrorResponseWriter(writer, UnauthorizedLogin)
		return
	}
	if err := auth.CheckPasswordHash(userDetails.HashedPassword, dataReceived.Password); err != nil {
		ErrorResponseWriter(writer, UnauthorizedLogin)
		return
	}

	jwtToken, err := auth.MakeJWT(userDetails.ID, a.SecretKey, time.Duration(3600) * time.Second)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()
	createRefreshTokenParams := database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: userDetails.ID,
	}
	createdRefreshToken, err := a.PtrToQueries.CreateRefreshToken(req.Context(), createRefreshTokenParams)
	if err != nil {
		ErrorResponseWriter(writer, DatabaseError)
		return
	}

	formattedUserDetails := validResponse{
		ID: userDetails.ID,
		CreatedAt: userDetails.CreatedAt,
		UpdatedAt: userDetails.UpdatedAt,
		Email: userDetails.Email,
		ChirpyRed: userDetails.IsChirpyRed,
		Token: jwtToken,
		RefreshToken: createdRefreshToken.Token,
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

func (a *APIConfig) PostRefresh(writer http.ResponseWriter, req *http.Request) {
	type validResponse struct {
		JWTToken string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadRT)
		return
	}
	refreshTokenDetails, err := a.PtrToQueries.GetToken(req.Context(), refreshToken)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadRT)
		return
	}
	if time.Now().After(refreshTokenDetails.ExpiresAt) {
		ErrorResponseWriter(writer, UnauthorizedBadRT)
		return
	}

	createdJWTToken, err := auth.MakeJWT(refreshTokenDetails.UserID, a.SecretKey, time.Duration(3600) * time.Second)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}

	tokenResponse := validResponse{
		JWTToken: createdJWTToken,
	}
	tokenResponseInBytes, err := json.Marshal(tokenResponse)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(tokenResponseInBytes); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}

func (a *APIConfig) PostRevoke(writer http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadRT)
		return
	}
	if err := a.PtrToQueries.RevokeToken(req.Context(), refreshToken); err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadRT)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (a *APIConfig) PutUsers(writer http.ResponseWriter, req *http.Request) {
	type requestBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type validResponse struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
		ChirpyRed bool `json:"is_chirpy_red"`
	}

	jwtToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadJWT)
		return
	}
	userID, err := auth.ValidateJWT(jwtToken, a.SecretKey)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadJWT)
		return
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

	updateUserParams := database.UpdateUserDetailsParams{
		Email: dataReceived.Email,
		HashedPassword: hashedPassword,
		ID: userID,
	}
	updatedUserDetails, err := a.PtrToQueries.UpdateUserDetails(req.Context(), updateUserParams)
	if err != nil {
		ErrorResponseWriter(writer, DatabaseError)
		return
	}

	formattedValidResponse := validResponse{
		ID: updatedUserDetails.ID,
		CreatedAt: updatedUserDetails.CreatedAt,
		UpdatedAt: updatedUserDetails.UpdatedAt,
		Email: updatedUserDetails.Email,
		ChirpyRed: updatedUserDetails.IsChirpyRed,
	}
	validResponseInBytes, err := json.Marshal(formattedValidResponse)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(validResponseInBytes); err != nil {
		ErrorResponseWriter(writer, ServiceError)
	}
}

func (a *APIConfig) DeleteChirps(writer http.ResponseWriter, req *http.Request) {
	jwtToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadJWT)
		return
	}
	userID, err := auth.ValidateJWT(jwtToken, a.SecretKey)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadJWT)
		return
	}

	chirpID := req.PathValue("chirpID")
	if chirpID == "" {
		ErrorResponseWriter(writer, NotFound)
		return
	}
	parsedChirpID, err := uuid.Parse(chirpID)
	if err != nil {
		ErrorResponseWriter(writer, NotFound)
		return
	}

	returnedChirp, err := a.PtrToQueries.GetOneChirp(req.Context(), parsedChirpID)
	if err != nil {
		ErrorResponseWriter(writer, NotFound)
		return
	}
	if returnedChirp.UserID != userID {
		ErrorResponseWriter(writer, Forbidden)
		return
	}

	if err := a.PtrToQueries.DeleteChirp(req.Context(), returnedChirp.ID); err != nil {
		ErrorResponseWriter(writer, DatabaseError)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (a *APIConfig) PostRed(writer http.ResponseWriter, req *http.Request) {
	type data struct {
		UserID string `json:"user_id"`
	}

	type requestBody struct {
		Event string `json:"event"`
		Data data `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		ErrorResponseWriter(writer, UnauthorizedBadAPIKey)
		return
	}
	if apiKey != a.WebhookKey {
		ErrorResponseWriter(writer, UnauthorizedBadAPIKey)
		return
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
	if dataReceived.Event != "user.upgraded" {
		ErrorResponseWriter(writer, NoContent)
		return
	}

	parsedUserID, err := uuid.Parse(dataReceived.Data.UserID)
	if err != nil {
		ErrorResponseWriter(writer, ServiceError)
		return
	}
	if err := a.PtrToQueries.UpdateRedUser(req.Context(), parsedUserID); err != nil {
		ErrorResponseWriter(writer, NotFound)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
