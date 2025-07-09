package main

import (
	"net/http"
	"log"
	"os"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/junwei890/chirpy/state"
	"github.com/junwei890/chirpy/internal/database"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	platform := os.Getenv("PLATFORM")

	ptrToAppState := &state.APIConfig{
		PtrToQueries: dbQueries,
		Platform: platform,
		Profanities: map[string]struct{}{
			"kerfuffle": {},
			"sharbert": {},
			"fornax": {},
		},
	}

	const root = "."
	const port = ":8080"

	const appPath = "/app/"
	const prefixToStrip = "/app"
	const getReadiness = "GET /api/healthz" 
	const getMetrics = "GET /admin/metrics"
	const postMetrics = "POST /admin/reset"
	const postUsers = "POST /api/users"
	const postLogin = "POST /api/login"
	const postChirps = "POST /api/chirps"
	const getChirps = "GET /api/chirps"
	const getChirp = "GET /api/chirps/{chirpID}"

	requestMultiplexer := http.NewServeMux()
	fileSystem := http.Dir(root)
	fileSystemHandler := http.FileServer(fileSystem)

	// Server readiness
	requestMultiplexer.HandleFunc(getReadiness, state.GetReadiness)

	// Metrics
	requestMultiplexer.Handle(appPath, http.StripPrefix(prefixToStrip, ptrToAppState.MiddlewareMetricsInc(fileSystemHandler)))
	requestMultiplexer.HandleFunc(getMetrics, ptrToAppState.GetMetrics)
	requestMultiplexer.HandleFunc(postMetrics, ptrToAppState.PostMetrics)

	// Chirp related
	requestMultiplexer.HandleFunc(postChirps, ptrToAppState.PostChirps)
	requestMultiplexer.HandleFunc(getChirps, ptrToAppState.GetChirps)
	requestMultiplexer.HandleFunc(getChirp, ptrToAppState.GetChirp)

	// User related
	requestMultiplexer.HandleFunc(postUsers, ptrToAppState.PostUsers)
	requestMultiplexer.HandleFunc(postLogin, ptrToAppState.PostLogin)

	server := &http.Server{
		Addr: port,
		Handler: requestMultiplexer,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	
}
