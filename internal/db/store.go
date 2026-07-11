package db

import (
	"database/sql"
	"encoding/json"
)

// DBStore implements worker.JobStore backed by a SQLite database.
type DBStore struct {
	DB *sql.DB
}

// EnqueueJob inserts a job and returns its ID.
func (s *DBStore) EnqueueJob(typ string, payload any) (int64, error) {
	return EnqueueJob(s.DB, typ, payload)
}

// FetchPendingJob fetches and marks a pending job as running.
func (s *DBStore) FetchPendingJob() (int64, string, string, error) {
	return FetchPendingJob(s.DB)
}

// UpdateJobStatus updates the status, progress, and error for a job.
func (s *DBStore) UpdateJobStatus(id int64, status string, progress float64, errMsg string) error {
	return UpdateJobStatus(s.DB, id, status, progress, errMsg)
}

// GetJobPayload returns the parsed payload for a job.
func (s *DBStore) GetJobPayload(id int64) (map[string]any, error) {
	return GetJobPayload(s.DB, id)
}

// UnmarshalPayload is a helper that unmarshals the raw payload JSON from FetchPendingJob
// into the provided target type.
func UnmarshalPayload(rawJSON string, target any) error {
	return json.Unmarshal([]byte(rawJSON), target)
}
