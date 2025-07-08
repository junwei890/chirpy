package state

import (
	"sync/atomic"
	"net/http"
	"fmt"
	"log"
	"github.com/junwei890/chirpy/internal/database"
)

type APIConfig struct {
	FileServerHits atomic.Int32
	PtrToQueries *database.Queries
}

func (a *APIConfig) MiddlewareMetricsInc(toHandle http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		a.FileServerHits.Add(1)
		toHandle.ServeHTTP(writer, req)
	})
}

func (a *APIConfig) Metrics(writer http.ResponseWriter, req *http.Request) {
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

func (a *APIConfig) Reset(writer http.ResponseWriter, req *http.Request) {
	a.FileServerHits.Store(0)
}
