package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CreateFolder creates a new folder under parent.
func (s *Service) CreateFolder(parent, name string) error {
	return os.MkdirAll(filepath.Join(parent, name), 0755)
}

// Rename renames oldPath to newName in the same directory.
func (s *Service) Rename(oldPath, newName string) error {
	return os.Rename(oldPath, filepath.Join(filepath.Dir(oldPath), newName))
}

// OpenFile opens the given path with the default shell association.
func (s *Service) OpenFile(path string) error {
	return exec.Command("cmd", "/c", "start", "", path).Start()
}

// ListDirectory returns the entries in path as service Entries.
func (s *Service) ListDirectory(path string) ([]Entry, error) {
	dtos, err := s.api.ListDirectory(path)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(dtos))
	for _, d := range dtos {
		mt, _ := time.Parse("2006-01-02 15:04:05", d.ModTime)
		entries = append(entries, Entry{
			Name:    d.Name,
			Path:    d.Path,
			IsDir:   d.IsDir,
			Size:    d.Size,
			ModTime: mt,
			Mode:    d.Mode,
		})
	}
	return entries, nil
}

// ListRoots returns the filesystem roots.
func (s *Service) ListRoots() ([]string, error) {
	return s.api.ListRoots()
}
