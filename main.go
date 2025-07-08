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
	const getReadinessPath = "GET /api/healthz" 
	const getMetricsPath = "GET /admin/metrics"
	const metricsResetPath = "POST /admin/reset"
	const newUserPath = "POST /api/users"
	const newChirpPath = "POST /api/chirps"
	const allChirpsPath = "GET /api/chirps"

	requestMultiplexer := http.NewServeMux()
	fileSystem := http.Dir(root)
	fileSystemHandler := http.FileServer(fileSystem)

	// Server readiness
	requestMultiplexer.HandleFunc(getReadinessPath, state.Readiness)

	// Metrics
	requestMultiplexer.Handle(appPath, http.StripPrefix(prefixToStrip, ptrToAppState.MiddlewareMetricsInc(fileSystemHandler)))
	requestMultiplexer.HandleFunc(getMetricsPath, ptrToAppState.Metrics)
	requestMultiplexer.HandleFunc(metricsResetPath, ptrToAppState.Reset)

	// Chirp related
	requestMultiplexer.HandleFunc(newChirpPath, ptrToAppState.NewChirp)
	requestMultiplexer.HandleFunc(allChirpsPath, ptrToAppState.AllChirps)

	// User related
	requestMultiplexer.HandleFunc(newUserPath, ptrToAppState.NewUser)

	server := &http.Server{
		Addr: port,
		Handler: requestMultiplexer,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	
}
