package db

import (
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	path := t.TempDir() + "/test.db"
	conn, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer conn.Close()

	var count int
	err = conn.QueryRow("SELECT COUNT(*) FROM jobs").Scan(&count)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 jobs after init, got %d", count)
	}
}

func TestEnqueueAndFetch(t *testing.T) {
	path := t.TempDir() + "/test.db"
	conn, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer conn.Close()

	id, err := EnqueueJob(conn, "copy", map[string]any{"sources": []string{"/a"}, "dest": "/b"})
	if err != nil {
		t.Fatalf("EnqueueJob failed: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}

	jobID, typ, payload, err := FetchPendingJob(conn)
	if err != nil {
		t.Fatalf("FetchPendingJob failed: %v", err)
	}
	if jobID != id {
		t.Errorf("expected jobID %d, got %d", id, jobID)
	}
	if typ != "copy" {
		t.Errorf("expected type copy, got %s", typ)
	}
	if payload == "" {
		t.Error("expected non-empty payload")
	}
}

func TestFetchPendingJobEmpty(t *testing.T) {
	path := t.TempDir() + "/test.db"
	conn, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer conn.Close()

	_, _, _, err = FetchPendingJob(conn)
	if err == nil {
		t.Fatal("expected error for empty queue")
	}
}

func TestUpdateJobStatus(t *testing.T) {
	path := t.TempDir() + "/test.db"
	conn, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer conn.Close()

	id, _ := EnqueueJob(conn, "copy", map[string]any{"sources": []string{"/a"}, "dest": "/b"})
	err = UpdateJobStatus(conn, id, "running", 0.5, "")
	if err != nil {
		t.Fatalf("UpdateJobStatus failed: %v", err)
	}

	var status string
	var progress float64
	err = conn.QueryRow("SELECT status, progress FROM jobs WHERE id=?", id).Scan(&status, &progress)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if status != "running" {
		t.Errorf("expected status running, got %s", status)
	}
	if progress != 0.5 {
		t.Errorf("expected progress 0.5, got %f", progress)
	}
}

func TestGetJobPayload(t *testing.T) {
	path := t.TempDir() + "/test.db"
	conn, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer conn.Close()

	payload := map[string]any{"sources": []string{"/a", "/b"}, "dest": "/c"}
	id, _ := EnqueueJob(conn, "copy", payload)

	got, err := GetJobPayload(conn, id)
	if err != nil {
		t.Fatalf("GetJobPayload failed: %v", err)
	}
	if got["dest"] != "/c" {
		t.Errorf("expected dest /c, got %v", got["dest"])
	}
}

func TestDeleteJob(t *testing.T) {
	path := t.TempDir() + "/test.db"
	conn, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer conn.Close()

	id, _ := EnqueueJob(conn, "delete", map[string]any{"sources": []string{"/a"}})
	err = UpdateJobStatus(conn, id, "done", 1.0, "")
	if err != nil {
		t.Fatalf("UpdateJobStatus failed: %v", err)
	}

	var status string
	err = conn.QueryRow("SELECT status FROM jobs WHERE id=?", id).Scan(&status)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if status != "done" {
		t.Errorf("expected status done, got %s", status)
	}
}

func TestInitDBCreatesFile(t *testing.T) {
	path := t.TempDir() + "/test.db"
	conn, err := InitDB(path)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	conn.Close()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}
