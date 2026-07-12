package service

import "garq/internal/shell"

// Copy enqueues a copy job.
func (s *Service) Copy(sources []string, dest, conflict string) (int64, error) {
	return s.api.AddCopyJob(sources, dest, conflict)
}

// Move enqueues a move job.
func (s *Service) Move(sources []string, dest, conflict string) (int64, error) {
	return s.api.AddMoveJob(sources, dest, conflict)
}

// Compress enqueues a compress job.
func (s *Service) Compress(sources []string, dest, conflict string) (int64, error) {
	return s.api.AddCompressJob(sources, dest, conflict)
}

// Extract enqueues an extract job.
func (s *Service) Extract(archive, dest, conflict string) (int64, error) {
	return s.api.AddExtractJob(archive, dest, conflict)
}

// Delete enqueues a permanent-delete job.
func (s *Service) Delete(sources []string) (int64, error) {
	return s.api.AddDeleteJob(sources)
}

// DeleteToRecycle moves the given paths to the recycle bin.
func (s *Service) DeleteToRecycle(paths []string) error {
	return shell.RecycleItems(paths)
}
