package handlers

import (
	"context"
	"net/http"
)

// Serve starts the demo HTTP server.
func Serve() error {
	server := &http.Server{Addr: ":0"}
	go server.ListenAndServe() // nolint:errcheck - fixture only
	return server.Shutdown(context.Background())
}
