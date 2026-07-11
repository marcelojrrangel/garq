package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// httpPort is used by the internal debug HTTP server.
// Note: cmd/app/ also starts an HTTP server on :8080 via http.Server directly.
// These two servers serve different purposes — the internal one serves /api/roots and
// /api/directory for development/debug, while cmd/app/ serves job-related endpoints.
// They coexist on different ports and do not conflict.
const httpPort = 19876

var (
	httpOnce sync.Once
	httpAPI  *API
)

// APIHandler returns an http.Handler that serves /api/* routes.
// Used for debug/development HTTP interface.
func APIHandler(a *API) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/roots") {
			handleRootsHandler(a)(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/api/directory") {
			handleDirectoryHandler(a)(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}

func StartHTTPServer(a *API) {
	httpOnce.Do(func() {
		httpAPI = a
		mux := http.NewServeMux()
		mux.HandleFunc("/api/roots", handleRootsHandler(a))
		mux.HandleFunc("/api/directory", handleDirectoryHandler(a))

		a.HTTPServer = &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d", httpPort),
			Handler: mux,
		}
		fmt.Printf("HTTP API server starting on %s\n", a.HTTPServer.Addr)
		go func() {
			if err := a.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("HTTP server error: %v\n", err)
			}
		}()
	})
}

func handleRootsHandler(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		roots, err := a.ListRoots()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"roots": roots})
	}
}

func handleDirectoryHandler(a *API) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Query().Get("path")
		if path == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "path parameter required"})
			return
		}
		entries, err := a.ListDirectory(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"entries": entries})
	}
}
