package main

import (
	"net/http"
	"log"
	"github.com/junwei890/chirpy/custom"
)

func main() {
	const root = "."
	const port = ":8080"
	const appPath = "/app/"
	const prefixToStrip = "/app"
	const readinessPath = "/healthz"

	// Creating a request multiplexer
	requestMultiplexer := http.NewServeMux()

	// Registering handlers to the multiplexer

	// File system
	fileSystem := http.Dir(root)
	fileSystemHandler := http.FileServer(fileSystem)
	requestMultiplexer.Handle(appPath, http.StripPrefix(prefixToStrip, fileSystemHandler))

	// Readiness
	requestMultiplexer.HandleFunc(readinessPath, custom.Readiness)

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
