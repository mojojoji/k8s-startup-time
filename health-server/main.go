package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

func main() {
	// Set GOMAXPROCS to match container CPU limits
	runtime.GOMAXPROCS(1)

	// Create a custom server with minimal settings
	server := &http.Server{
		Addr: fmt.Sprintf(":%s", getPort()),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
				return
			}
			http.NotFound(w, r)
		}),
		// Minimal timeouts
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	log.Printf("Starting health server on port %s", getPort())
	log.Fatal(server.ListenAndServe())
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}
