package service

import (
	"database/sql"

	"garq/compress"
	"garq/internal/api"
	"garq/internal/copy"
	"garq/internal/db"
	"garq/internal/worker"
)

type Service struct {
	api *api.API
}

func New(a *api.API) *Service {
	return &Service{api: a}
}

// NewFromDB creates a Service directly from a database connection.
func NewFromDB(dbConn *sql.DB) *Service {
	return New(api.New(dbConn))
}

// InitDB initializes the SQLite database used by the application.
func InitDB(path string) (*sql.DB, error) {
	return db.InitDB(path)
}

// StartWorkerPool starts the background worker pool with the default adapters.
func StartWorkerPool(dbConn *sql.DB) {
	store := &db.DBStore{DB: dbConn}
	worker.StartWorkerPool(4, store, compress.CLIAdapter{}, copy.CopierAdapter{})
}
