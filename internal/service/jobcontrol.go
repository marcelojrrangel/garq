package service

import "garq/internal/worker"

// JobSnapshot returns the current state of a single job.
func (s *Service) JobSnapshot(id int64) (JobSnapshot, error) {
	row := s.api.DB.QueryRow(
		`SELECT id, type, status, progress, COALESCE(error,''), COALESCE(payload,'') FROM jobs WHERE id=?`,
		id,
	)
	var j JobSnapshot
	err := row.Scan(&j.ID, &j.Type, &j.Status, &j.Progress, &j.ErrMsg, &j.Payload)
	return j, err
}

// ActiveJobCount returns the number of pending or running jobs.
func (s *Service) ActiveJobCount() (int, error) {
	row := s.api.DB.QueryRow(`SELECT COUNT(*) FROM jobs WHERE status IN ('pending','running')`)
	var count int
	err := row.Scan(&count)
	return count, err
}

// PauseJob delegates to the worker pool.
func (s *Service) PauseJob(id int64) { worker.PauseJob(id) }

// ResumeJob delegates to the worker pool.
func (s *Service) ResumeJob(id int64) { worker.ResumeJob(id) }

// CancelJob delegates to the worker pool.
func (s *Service) CancelJob(id int64) { worker.CancelJob(id) }
