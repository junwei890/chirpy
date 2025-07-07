package custom

import (
	"log"
	"net/http"
)

// Custom handler to check if server is ready for requests, writing how we will respond to requests
func Readiness(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Println(err)
	}
}
