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

	ptrToAppState := &state.APIConfig{}

	requestMultiplexer := http.NewServeMux()

	fileSystem := http.Dir(root)
	fileSystemHandler := http.FileServer(fileSystem)

	requestMultiplexer.HandleFunc(readinessPath, custom.Readiness)

	requestMultiplexer.Handle(appPath, http.StripPrefix(prefixToStrip, ptrToAppState.MiddlewareMetricsInc(fileSystemHandler)))

	requestMultiplexer.HandleFunc(metricsPath, ptrToAppState.Metrics)
	requestMultiplexer.HandleFunc(resetPath, ptrToAppState.Reset)

	requestMultiplexer.HandleFunc(validationPath, custom.ValidateChirp)

	server := &http.Server{
		Addr: port,
		Handler: requestMultiplexer,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	
}
