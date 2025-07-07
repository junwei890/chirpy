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
	const readinessPath = "GET /healthz"
	const metricsPath = "GET /metrics"
	const resetPath = "POST /reset"

	// Creating app state
	ptrToAppState := &state.APIConfig{}

	// Creating a request multiplexer
	requestMultiplexer := http.NewServeMux()

	// Registering handlers to the multiplexer

	// File system
	// Defining my file system
	fileSystem := http.Dir(root)
	// Returns a handler that serves files from my filesystem
	fileSystemHandler := http.FileServer(fileSystem)

	// Readiness
	requestMultiplexer.HandleFunc(readinessPath, custom.Readiness)

	// File server hits then serve file system
	requestMultiplexer.Handle(appPath, http.StripPrefix(prefixToStrip, ptrToAppState.MiddlewareMetricsInc(fileSystemHandler)))

	// Metrics
	requestMultiplexer.HandleFunc(metricsPath, ptrToAppState.Metrics)
	requestMultiplexer.HandleFunc(resetPath, ptrToAppState.Reset)

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
