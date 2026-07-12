package service

import "time"

// Entry represents a filesystem item returned to the UI.
type Entry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
	Mode    string
}

// JobSnapshot is a lightweight view of a job record.
type JobSnapshot struct {
	ID       int64
	Type     string
	Status   string
	Progress float64
	ErrMsg   string
	Payload  string
}

// ClipboardOp is the typed clipboard payload exchanged by the UI.
type ClipboardOp struct {
	Op    string   `json:"op"`
	Paths []string `json:"paths"`
}
