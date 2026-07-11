package db

import (
	"database/sql"
	"encoding/json"

	_ "modernc.org/sqlite"
)

func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(string(mustReadSchema())); err != nil {
		return nil, err
	}
	// Mark jobs that were running as failed on crash recovery
	_, _ = db.Exec("UPDATE jobs SET status='failed', error='service restarted' WHERE status='running'")
	return db, nil
}

func mustReadSchema() []byte {
	// Schema embedded here for simplicity
	s := `
CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    payload TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    progress REAL DEFAULT 0.0,
    error TEXT
);
`
	return []byte(s)
}

// EnqueueJob inserts a job and returns inserted id
func EnqueueJob(db *sql.DB, typ string, payload any) (int64, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	res, err := db.Exec("INSERT INTO jobs(type,payload) VALUES(?,?)", typ, string(b))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func FetchPendingJob(db *sql.DB) (int64, string, string, error) {
	// simple fetch and mark running in a transaction
	tx, err := db.Begin()
	if err != nil {
		return 0, "", "", err
	}
	row := tx.QueryRow("SELECT id, type, payload FROM jobs WHERE status='pending' ORDER BY id LIMIT 1")
	var id int64
	var typ string
	var payload string
	if err := row.Scan(&id, &typ, &payload); err != nil {
		tx.Rollback()
		return 0, "", "", err
	}
	if _, err := tx.Exec("UPDATE jobs SET status='running', updated_at=CURRENT_TIMESTAMP WHERE id=?", id); err != nil {
		tx.Rollback()
		return 0, "", "", err
	}
	if err := tx.Commit(); err != nil {
		return 0, "", "", err
	}
	return id, typ, payload, nil
}

func UpdateJobStatus(db *sql.DB, id int64, status string, progress float64, errMsg string) error {
	_, err := db.Exec("UPDATE jobs SET status=?, progress=?, error=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", status, progress, errMsg, id)
	return err
}

func GetJobPayload(db *sql.DB, id int64) (map[string]any, error) {
	row := db.QueryRow("SELECT payload FROM jobs WHERE id=?", id)
	var payload string
	if err := row.Scan(&payload); err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		return nil, err
	}
	return out, nil
}
