package service

import (
	"path/filepath"
	"testing"

	"garq/internal/api"
	"garq/internal/db"
)

func setupTestService(t *testing.T) (*Service, func()) {
	t.Helper()
	dir := t.TempDir()
	dbConn, err := db.InitDB(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	a := api.New(dbConn)
	svc := New(a)
	return svc, func() { dbConn.Close() }
}

func TestJobSnapshot(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	id, err := db.EnqueueJob(svc.api.DBConn(), "copy", map[string]any{
		"sources": []string{"a.txt"},
		"dest":    "b",
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	snap, err := svc.JobSnapshot(id)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snap.ID != id {
		t.Errorf("ID = %d, want %d", snap.ID, id)
	}
	if snap.Type != "copy" {
		t.Errorf("Type = %q, want copy", snap.Type)
	}
	if snap.Status != "pending" {
		t.Errorf("Status = %q, want pending", snap.Status)
	}
	if snap.Progress != 0 {
		t.Errorf("Progress = %v, want 0", snap.Progress)
	}
	if snap.ErrMsg != "" {
		t.Errorf("ErrMsg = %q, want empty", snap.ErrMsg)
	}
	if snap.Payload == "" {
		t.Error("Payload should not be empty")
	}

	if err := db.UpdateJobStatus(svc.api.DBConn(), id, "running", 0.5, "oops"); err != nil {
		t.Fatalf("update status: %v", err)
	}
	snap, err = svc.JobSnapshot(id)
	if err != nil {
		t.Fatalf("snapshot after update: %v", err)
	}
	if snap.Status != "running" {
		t.Errorf("Status = %q, want running", snap.Status)
	}
	if snap.Progress != 0.5 {
		t.Errorf("Progress = %v, want 0.5", snap.Progress)
	}
	if snap.ErrMsg != "oops" {
		t.Errorf("ErrMsg = %q, want oops", snap.ErrMsg)
	}
}

func TestActiveJobCount(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	count, err := svc.ActiveJobCount()
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	id, err := db.EnqueueJob(svc.api.DBConn(), "move", map[string]any{
		"sources": []string{"a.txt"},
		"dest":    "b",
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	count, err = svc.ActiveJobCount()
	if err != nil {
		t.Fatalf("count after enqueue: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	if err := db.UpdateJobStatus(svc.api.DBConn(), id, "done", 1.0, ""); err != nil {
		t.Fatalf("update status: %v", err)
	}
	count, err = svc.ActiveJobCount()
	if err != nil {
		t.Fatalf("count after done: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}
