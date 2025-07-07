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
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		a.FileServerHits.Add(1)
		toHandle.ServeHTTP(writer, req)
	})
}

// A custom handler that logs number of hits to the server
func (a *APIConfig) Metrics(writer http.ResponseWriter, req *http.Request) {
	// Writing the body of response
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

// A custom handler that resets number of hits to the server
func (a *APIConfig) Reset(writer http.ResponseWriter, req *http.Request) {
	a.FileServerHits.Store(0)
}
