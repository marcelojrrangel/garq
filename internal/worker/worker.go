package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Compressor compresses and extracts archives with progress reporting.
type Compressor interface {
	CompressManyCtx(ctx context.Context, sources []string, dest string, progressCb func(float64)) error
	ExtractCtx(ctx context.Context, archive, dest string, progressCb func(float64)) error
}

// FileCopier copies a single file from src to destDir.
type FileCopier interface {
	CopyFile(src, destDir string) error
}

// JobStore persists and retrieves job records.
type JobStore interface {
	EnqueueJob(typ string, payload any) (int64, error)
	FetchPendingJob() (int64, string, string, error)
	UpdateJobStatus(id int64, status string, progress float64, errMsg string) error
	GetJobPayload(id int64) (map[string]any, error)
}

type jobState struct {
	cancel context.CancelFunc
	mu     sync.Mutex
	cond   *sync.Cond
	paused bool
}

var (
	mu     sync.Mutex
	states = make(map[int64]*jobState)
)

// StartWorkerPool launches n worker goroutines that process jobs via the given dependencies.
func StartWorkerPool(n int, store JobStore, compressor Compressor, copier FileCopier) {
	for i := 0; i < n; i++ {
		go workerLoop(i, store, compressor, copier)
	}
}

func CancelJob(jobID int64) {
	mu.Lock()
	js, ok := states[jobID]
	mu.Unlock()
	if ok {
		js.cancel()
		js.mu.Lock()
		js.paused = false
		js.cond.Broadcast()
		js.mu.Unlock()
	}
}

// PauseJob suspende a execução de um job. O worker bloqueia na próxima checkPause.
func PauseJob(jobID int64) {
	mu.Lock()
	js, ok := states[jobID]
	mu.Unlock()
	if !ok {
		return
	}
	js.mu.Lock()
	js.paused = true
	js.mu.Unlock()
}

// ResumeJob retoma um job pausado.
func ResumeJob(jobID int64) {
	mu.Lock()
	js, ok := states[jobID]
	mu.Unlock()
	if !ok {
		return
	}
	js.mu.Lock()
	js.paused = false
	js.cond.Broadcast()
	js.mu.Unlock()
}

// checkPause verifica se o job deve pausar. Bloqueia via sync.Cond até ResumeJob.
func checkPause(jobID int64, ctx context.Context) bool {
	mu.Lock()
	js, ok := states[jobID]
	mu.Unlock()
	if !ok {
		return false
	}
	js.mu.Lock()
	defer js.mu.Unlock()
	for js.paused {
		js.cond.Wait()
		if ctx.Err() != nil {
			return true // cancelado enquanto pausado
		}
	}
	return false
}

func registerJob(ctx context.Context, jobID int64) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	js := &jobState{cancel: cancel}
	js.cond = sync.NewCond(&js.mu)
	mu.Lock()
	states[jobID] = js
	mu.Unlock()
	return ctx
}

func unregisterJob(jobID int64) {
	mu.Lock()
	delete(states, jobID)
	mu.Unlock()
}

func workerLoop(id int, store JobStore, compressor Compressor, copier FileCopier) {
	log.Printf("worker %d started", id)
	for {
		jobID, typ, rawPayload, err := store.FetchPendingJob()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		log.Printf("worker %d got job %d type=%s", id, jobID, typ)

		ctx := registerJob(context.Background(), jobID)
		var jobErr error

		switch typ {
		case "copy":
			jobErr = runCopyJob(ctx, store, copier, jobID, rawPayload)
		case "compress":
			jobErr = runCompressJob(ctx, store, compressor, jobID, rawPayload)
		case "extract":
			jobErr = runExtractJob(ctx, store, compressor, jobID, rawPayload)
		case "move":
			jobErr = runMoveJob(ctx, store, copier, jobID, rawPayload)
		case "delete":
			jobErr = runDeleteJob(ctx, store, jobID, rawPayload)
		default:
			jobErr = errors.New("unsupported job type")
		}

		unregisterJob(jobID)

		if ctx.Err() != nil {
			_ = store.UpdateJobStatus(jobID, "failed", 0, "cancelled")
			continue
		}
		if jobErr != nil {
			_ = store.UpdateJobStatus(jobID, "failed", 0, jobErr.Error())
			continue
		}
		_ = store.UpdateJobStatus(jobID, "done", 1.0, "")
	}
}

func sanitizePath(p string) (string, error) {
	clean := filepath.Clean(p)
	if strings.Contains(clean, "..") {
		return "", errors.New("path traversal detected")
	}
	if !filepath.IsAbs(clean) {
		return "", errors.New("path must be absolute")
	}
	return clean, nil
}

func runCopyJob(ctx context.Context, store JobStore, copier FileCopier, jobID int64, rawPayload string) error {
	var p CopyPayload
	if err := json.Unmarshal([]byte(rawPayload), &p); err != nil {
		return fmt.Errorf("invalid copy payload: %w", err)
	}
	if p.Conflict == "" {
		p.Conflict = "replace"
	}
	if len(p.Sources) == 0 {
		return errors.New("copy job requires at least one source")
	}

	cleanDest, err := sanitizePath(p.Dest)
	if err != nil {
		return fmt.Errorf("invalid dest: %w", err)
	}

	for idx, src := range p.Sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if checkPause(jobID, ctx) {
			return ctx.Err()
		}
		cleanSrc, err := sanitizePath(src)
		if err != nil {
			return fmt.Errorf("invalid source: %w", err)
		}
		fileName := filepath.Base(cleanSrc)
		dstPath := filepath.Join(cleanDest, fileName)

		if _, err := os.Stat(dstPath); err == nil {
			switch p.Conflict {
			case "skip":
				_ = store.UpdateJobStatus(jobID, "running", float64(idx+1)/float64(len(p.Sources)), "")
				continue
			case "rename":
				dstPath = renamePath(dstPath)
			}
		}

		if err := copier.CopyFile(cleanSrc, cleanDest); err != nil {
			return err
		}
		_ = store.UpdateJobStatus(jobID, "running", float64(idx+1)/float64(len(p.Sources)), "")
	}
	return nil
}

func runCompressJob(ctx context.Context, store JobStore, compressor Compressor, jobID int64, rawPayload string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	var p CompressPayload
	if err := json.Unmarshal([]byte(rawPayload), &p); err != nil {
		return fmt.Errorf("invalid compress payload: %w", err)
	}
	if p.Conflict == "" {
		p.Conflict = "replace"
	}
	if len(p.Sources) == 0 {
		return errors.New("compress job requires at least one source")
	}

	cleanSources := make([]string, 0, len(p.Sources))
	for _, s := range p.Sources {
		cs, err := sanitizePath(s)
		if err != nil {
			return fmt.Errorf("invalid source: %w", err)
		}
		cleanSources = append(cleanSources, cs)
	}
	cleanDest, err := sanitizePath(p.Dest)
	if err != nil {
		return fmt.Errorf("invalid dest: %w", err)
	}

	if _, statErr := os.Stat(cleanDest); statErr == nil {
		switch p.Conflict {
		case "skip":
			_ = store.UpdateJobStatus(jobID, "running", 1.0, "")
			return nil
		case "rename":
			cleanDest = renamePath(cleanDest)
		}
	}

	progressCb := func(pct float64) {
		if checkPause(jobID, ctx) {
			return
		}
		_ = store.UpdateJobStatus(jobID, "running", pct, "")
	}
	if err := compressor.CompressManyCtx(ctx, cleanSources, cleanDest, progressCb); err != nil {
		return err
	}
	return nil
}

func runExtractJob(ctx context.Context, store JobStore, compressor Compressor, jobID int64, rawPayload string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	var p ExtractPayload
	if err := json.Unmarshal([]byte(rawPayload), &p); err != nil {
		return fmt.Errorf("invalid extract payload: %w", err)
	}
	if p.Conflict == "" {
		p.Conflict = "replace"
	}

	cleanArchive, err := sanitizePath(p.Archive)
	if err != nil {
		return fmt.Errorf("invalid archive: %w", err)
	}
	cleanDest, err := sanitizePath(p.Dest)
	if err != nil {
		return fmt.Errorf("invalid dest: %w", err)
	}

	if _, statErr := os.Stat(cleanDest); statErr == nil {
		switch p.Conflict {
		case "skip":
			_ = store.UpdateJobStatus(jobID, "running", 1.0, "")
			return nil
		case "rename":
			cleanDest = renamePath(cleanDest)
		}
	}

	progressCb := func(pct float64) {
		if checkPause(jobID, ctx) {
			return
		}
		_ = store.UpdateJobStatus(jobID, "running", pct, "")
	}
	if err := compressor.ExtractCtx(ctx, cleanArchive, cleanDest, progressCb); err != nil {
		return err
	}
	return nil
}

func renamePath(p string) string {
	ext := filepath.Ext(p)
	base := p[:len(p)-len(ext)]
	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return p
}

func runMoveJob(ctx context.Context, store JobStore, copier FileCopier, jobID int64, rawPayload string) error {
	var p MovePayload
	if err := json.Unmarshal([]byte(rawPayload), &p); err != nil {
		return fmt.Errorf("invalid move payload: %w", err)
	}
	if p.Conflict == "" {
		p.Conflict = "replace"
	}
	if len(p.Sources) == 0 {
		return errors.New("move job requires at least one source")
	}

	cleanDest, err := sanitizePath(p.Dest)
	if err != nil {
		return fmt.Errorf("invalid dest: %w", err)
	}

	for idx, src := range p.Sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if checkPause(jobID, ctx) {
			return ctx.Err()
		}
		cleanSrc, err := sanitizePath(src)
		if err != nil {
			return fmt.Errorf("invalid source: %w", err)
		}
		fileName := filepath.Base(cleanSrc)
		dstPath := filepath.Join(cleanDest, fileName)

		if _, err := os.Stat(dstPath); err == nil {
			switch p.Conflict {
			case "skip":
				_ = store.UpdateJobStatus(jobID, "running", float64(idx+1)/float64(len(p.Sources)), "")
				continue
			case "rename":
				dstPath = renamePath(dstPath)
			}
		}

		if err := os.Rename(cleanSrc, dstPath); err != nil {
			if err := copyFileSimple(cleanSrc, dstPath); err != nil {
				return err
			}
			os.Remove(cleanSrc)
		}
		_ = store.UpdateJobStatus(jobID, "running", float64(idx+1)/float64(len(p.Sources)), "")
	}
	return nil
}

func runDeleteJob(ctx context.Context, store JobStore, jobID int64, rawPayload string) error {
	var p DeletePayload
	if err := json.Unmarshal([]byte(rawPayload), &p); err != nil {
		return fmt.Errorf("invalid delete payload: %w", err)
	}
	if len(p.Sources) == 0 {
		return errors.New("delete job requires at least one source")
	}

	for idx, src := range p.Sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		cleanSrc, err := sanitizePath(src)
		if err != nil {
			return fmt.Errorf("invalid source: %w", err)
		}
		info, err := os.Stat(cleanSrc)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if info.IsDir() {
			err = os.RemoveAll(cleanSrc)
		} else {
			err = os.Remove(cleanSrc)
		}
		if err != nil {
			return err
		}
		_ = store.UpdateJobStatus(jobID, "running", float64(idx+1)/float64(len(p.Sources)), "")
	}
	return nil
}

func copyFileSimple(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	info, err := in.Stat()
	if err == nil {
		os.Chtimes(dst, info.ModTime(), info.ModTime())
	}
	return out.Close()
}
