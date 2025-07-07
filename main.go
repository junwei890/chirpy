package main

import (
	"net/http"
	"log"
	"github.com/junwei890/chirpy/custom"
	"github.com/junwei890/chirpy/state"
)

func main() {
	const root = "."
	const port = ":8080"
	// Endpoints
	const appPath = "/app/"
	const prefixToStrip = "/app"
	const readinessPath = "GET /api/healthz" 
	const metricsPath = "GET /admin/metrics"
	const resetPath = "POST /admin/reset"
	const validationPath = "POST /api/validate_chirp"

	// Creating app state
	ptrToAppState := &state.APIConfig{}

	// Creating a request multiplexer
	requestMultiplexer := http.NewServeMux()

	// Registering handlers to the multiplexer

	// File system
	fileSystem := http.Dir(root) // Defining my file system
	// Returns a handler that serves files from my filesystem
	fileSystemHandler := http.FileServer(fileSystem)

	// Readiness
	requestMultiplexer.HandleFunc(readinessPath, custom.Readiness)

	// File server hits then serve file system
	requestMultiplexer.Handle(appPath, http.StripPrefix(prefixToStrip, ptrToAppState.MiddlewareMetricsInc(fileSystemHandler)))

	// Metrics
	requestMultiplexer.HandleFunc(metricsPath, ptrToAppState.Metrics)
	requestMultiplexer.HandleFunc(resetPath, ptrToAppState.Reset)

	// Validate chirp length
	requestMultiplexer.HandleFunc(validationPath, custom.ValidateChirp)

	// Creating a server struct and running it
	server := &http.Server{
		Addr: port,
		Handler: requestMultiplexer,
	}
	// This blocks until the server is shutdown
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	
}
