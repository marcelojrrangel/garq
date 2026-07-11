package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lxn/walk"
)

// SetClipboard stores a typed clipboard operation in the walk clipboard.
func (s *Service) SetClipboard(op string, paths []string) {
	b, _ := json.Marshal(ClipboardOp{Op: op, Paths: paths})
	walk.Clipboard().SetText(string(b))
}

// GetClipboard reads and parses the typed clipboard operation.
func (s *Service) GetClipboard() (ClipboardOp, error) {
	text, err := walk.Clipboard().Text()
	if err != nil {
		return ClipboardOp{}, err
	}
	if text == "" {
		return ClipboardOp{}, fmt.Errorf("empty clipboard")
	}
	var op ClipboardOp
	if err := json.Unmarshal([]byte(text), &op); err != nil {
		return ClipboardOp{}, err
	}
	return op, nil
}

// Paste reads the clipboard and enqueues the matching job.
// It returns the job ID, job type ("copy" or "move"), and any error.
func (s *Service) Paste(dest string) (int64, string, error) {
	op, err := s.GetClipboard()
	if err != nil {
		return 0, "", err
	}
	if len(op.Paths) == 0 {
		return 0, "", fmt.Errorf("no paths in clipboard")
	}
	switch strings.ToLower(op.Op) {
	case "cut":
		id, err := s.api.AddMoveJob(op.Paths, dest, "replace")
		if err != nil {
			return 0, "", err
		}
		walk.Clipboard().Clear()
		return id, "move", nil
	case "copy":
		id, err := s.api.AddCopyJob(op.Paths, dest, "replace")
		if err != nil {
			return 0, "", err
		}
		return id, "copy", nil
	default:
		return 0, "", fmt.Errorf("unknown clipboard op: %s", op.Op)
	}
}
