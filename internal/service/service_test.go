package service

import (
	"os"
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

func TestCreateFolder(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	parent := t.TempDir()
	if err := svc.CreateFolder(parent, "newdir"); err != nil {
		t.Fatalf("create folder: %v", err)
	}
	info, err := os.Stat(filepath.Join(parent, "newdir"))
	if err != nil {
		t.Fatalf("stat folder: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("created path is not a directory")
	}
}

func TestRename(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	parent := t.TempDir()
	oldPath := filepath.Join(parent, "old.txt")
	if err := os.WriteFile(oldPath, []byte("x"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := svc.Rename(oldPath, "new.txt"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("old path still exists")
	}
	if _, err := os.Stat(filepath.Join(parent, "new.txt")); err != nil {
		t.Errorf("new path does not exist: %v", err)
	}
}

func TestPasteRoutesCutToMove(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(src, []byte("x"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	dest := filepath.Join(tmp, "dest")
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	svc.SetClipboard("cut", []string{src})
	id, jobType, err := svc.Paste(dest)
	if err != nil {
		t.Fatalf("paste: %v", err)
	}
	if jobType != "move" {
		t.Errorf("jobType = %q, want move", jobType)
	}
	snap, err := svc.JobSnapshot(id)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snap.Type != "move" {
		t.Errorf("snap.Type = %q, want move", snap.Type)
	}
}

func TestPasteRoutesCopyToCopy(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	tmp := t.TempDir()
	src := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(src, []byte("x"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	dest := filepath.Join(tmp, "dest")
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	svc.SetClipboard("copy", []string{src})
	id, jobType, err := svc.Paste(dest)
	if err != nil {
		t.Fatalf("paste: %v", err)
	}
	if jobType != "copy" {
		t.Errorf("jobType = %q, want copy", jobType)
	}
	snap, err := svc.JobSnapshot(id)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snap.Type != "copy" {
		t.Errorf("snap.Type = %q, want copy", snap.Type)
	}
}

func TestPasteEmptyClipboard(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	svc.SetClipboard("", []string{})
	_, _, err := svc.Paste(t.TempDir())
	if err == nil {
		t.Error("expected error for empty clipboard, got nil")
	}
}
