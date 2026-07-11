package api

// DirEntryDTO represents a filesystem entry returned by ListDirectory.
type DirEntryDTO struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
	Mode    string `json:"mode"`
}

// JobDTO represents a job returned by GetJobs.
type JobDTO struct {
	ID        int64   `json:"id"`
	Type      string  `json:"type"`
	Payload   any     `json:"payload"`
	Status    string  `json:"status"`
	Progress  float64 `json:"progress"`
	Error     string  `json:"error"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}
