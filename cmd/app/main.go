package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"garq/compress"
	"garq/internal/api"
	copyimpl "garq/internal/copy"
	"garq/internal/db"
	"garq/internal/worker"
)

func main() {
	// Initialize DB
	dbPath := "garq.db"
	dbConn, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	// Start worker pool
	store := &db.DBStore{DB: dbConn}
	worker.StartWorkerPool(4, store, compress.CLIAdapter{}, copyimpl.CopierAdapter{})
	bindAPI := api.New(dbConn)

	// Minimal HTTP API for demonstration (replace with Wails bindings)
	http.HandleFunc("/jobs/copy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Sources []string `json:"sources"`
			Dest    string   `json:"dest"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := bindAPI.AddCopyJob(payload.Sources, payload.Dest, "replace")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": id})
	})

	http.HandleFunc("/jobs/compress", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Sources []string `json:"sources"`
			Dest    string   `json:"dest"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := bindAPI.AddCompressJob(payload.Sources, payload.Dest, "replace")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": id})
	})

	http.HandleFunc("/jobs/extract", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Archive string `json:"archive"`
			Dest    string `json:"dest"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := bindAPI.AddExtractJob(payload.Archive, payload.Dest, "replace")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": id})
	})

	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		jobs, err := bindAPI.GetJobs()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(jobs)
	})

	// Simple health
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	srv := &http.Server{Addr: ":8080"}

	// Graceful shutdown
	go func() {
		log.Println("HTTP server listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down")
	srv.Close()
}
