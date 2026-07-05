package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"garq/internal/db"
	"garq/internal/worker"
)

// API holds resources for the file manager backend.
type API struct {
	DB  *sql.DB
	Ctx context.Context
}

// AddCopyJob enqueues a copy job. Called from frontend via Wails.
func (a *API) AddCopyJob(sources []string, dest string, conflict string) (int64, error) {
	if len(sources) == 0 {
		return 0, errors.New("sources cannot be empty")
	}
	if dest == "" {
		return 0, errors.New("destination cannot be empty")
	}
	if conflict == "" {
		conflict = "replace"
	}
	payload := map[string]any{"sources": sources, "dest": dest, "conflict": conflict}
	return db.EnqueueJob(a.DB, "copy", payload)
}

func (a *API) AddCompressJob(sources []string, dest string, conflict string) (int64, error) {
	if len(sources) == 0 {
		return 0, errors.New("sources cannot be empty")
	}
	if dest == "" {
		return 0, errors.New("destination is required")
	}
	if conflict == "" {
		conflict = "replace"
	}
	payload := map[string]any{"sources": sources, "dest": dest, "conflict": conflict}
	return db.EnqueueJob(a.DB, "compress", payload)
}

func (a *API) AddExtractJob(archive string, dest string, conflict string) (int64, error) {
	if archive == "" || dest == "" {
		return 0, errors.New("archive and destination are required")
	}
	if conflict == "" {
		conflict = "replace"
	}
	payload := map[string]any{"archive": archive, "dest": dest, "conflict": conflict}
	return db.EnqueueJob(a.DB, "extract", payload)
}

// AddMoveJob enqueues a move (cut+paste) job.
func (a *API) AddMoveJob(sources []string, dest string, conflict string) (int64, error) {
	if len(sources) == 0 {
		return 0, errors.New("sources cannot be empty")
	}
	if dest == "" {
		return 0, errors.New("destination cannot be empty")
	}
	if conflict == "" {
		conflict = "replace"
	}
	payload := map[string]any{"sources": sources, "dest": dest, "conflict": conflict}
	return db.EnqueueJob(a.DB, "move", payload)
}

// AddDeleteJob enqueues a delete job.
func (a *API) AddDeleteJob(sources []string) (int64, error) {
	if len(sources) == 0 {
		return 0, errors.New("sources cannot be empty")
	}
	payload := map[string]any{"sources": sources}
	return db.EnqueueJob(a.DB, "delete", payload)
}

// GetJobs returns recent jobs (simple representation). Frontend can parse payload JSON.
func (a *API) GetJobs() ([]map[string]any, error) {
	rows, err := a.DB.Query("SELECT id,type,payload,status,progress,error,created_at,updated_at FROM jobs ORDER BY id DESC LIMIT 200")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var typ, payload, status string
		var progress float64
		var errMsg sql.NullString
		var createdAt, updatedAt string
		if err := rows.Scan(&id, &typ, &payload, &status, &progress, &errMsg, &createdAt, &updatedAt); err != nil {
			continue
		}
		var parsed any
		_ = json.Unmarshal([]byte(payload), &parsed)
		errStr := ""
		if errMsg.Valid {
			errStr = errMsg.String
		}
		m := map[string]any{
			"id":         id,
			"type":       typ,
			"payload":    parsed,
			"status":     status,
			"progress":   progress,
			"error":      errStr,
			"created_at": createdAt,
			"updated_at": updatedAt,
		}
		out = append(out, m)
	}
	return out, nil
}

// CancelJob cancels a running or pending job.
func (a *API) CancelJob(jobID int64) error {
	worker.CancelJob(jobID)
	return nil
}
