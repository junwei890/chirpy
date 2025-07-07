package state

import (
	"sync/atomic"
	"net/http"
	"fmt"
	"log"
)

type APIConfig struct {
	FileServerHits atomic.Int32
}

// This is a middleware that takes a handler then returns a handler that first increases server hits then calls the handler again
func (a *APIConfig) MiddlewareMetricsInc(toHandle http.Handler) http.Handler {
	// HandlerFunc allows me to write a handler as a function with the appropriate signature
	// writer is just some thing I can customise but in this case I don't
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		a.FileServerHits.Add(1)
		toHandle.ServeHTTP(writer, req)
	})
}

// A custom handler that logs number of hits to the server
func (a *APIConfig) Metrics(writer http.ResponseWriter, req *http.Request) {
	// Writing the body of response
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	serverHits := fmt.Sprintf("Hits: %d", a.FileServerHits.Load())
	if _, err := writer.Write([]byte(serverHits)); err != nil {
		log.Println(err)
	}
}

// A custom handler that resets number of hits to the server
func (a *APIConfig) Reset(writer http.ResponseWriter, req *http.Request) {
	a.FileServerHits.Store(0)
}
