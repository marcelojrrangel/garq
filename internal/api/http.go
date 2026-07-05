package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

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

		addr := fmt.Sprintf("127.0.0.1:%d", httpPort)
		fmt.Printf("HTTP API server starting on %s\n", addr)
		go http.ListenAndServe(addr, mux)
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
