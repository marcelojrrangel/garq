package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"garq/compress"
	copyimpl "garq/internal/copy"
	"garq/internal/db"
)

var (
	mu       sync.Mutex
	jobs     = make(map[int64]context.CancelFunc)
	pauseChs = make(map[int64]chan struct{})
)

func StartWorkerPool(n int, conn *sql.DB) {
	for i := 0; i < n; i++ {
		go workerLoop(i, conn)
	}
}

func CancelJob(jobID int64) {
	mu.Lock()
	defer mu.Unlock()
	if cancel, ok := jobs[jobID]; ok {
		cancel()
		delete(jobs, jobID)
	}
	// Desbloqueia se estiver pausado
	if ch, ok := pauseChs[jobID]; ok {
		select {
		case ch <- struct{}{}:
		default:
		}
		delete(pauseChs, jobID)
	}
}

// PauseJob suspende a execução de um job. Bloqueia o worker até ResumeJob ser chamado.
func PauseJob(jobID int64) {
	mu.Lock()
	ch, ok := pauseChs[jobID]
	mu.Unlock()
	if !ok {
		return
	}
	// Sinaliza pausa — o worker lê do canal e fica bloqueado aguardando retomada
	ch <- struct{}{}
}

// ResumeJob retoma um job pausado.
func ResumeJob(jobID int64) {
	mu.Lock()
	ch, ok := pauseChs[jobID]
	mu.Unlock()
	if !ok {
		return
	}
	ch <- struct{}{}
}

// checkPause verifica se o job deve pausar. Usa dois sinais no mesmo canal:
// primeiro sinal = pausar (bloqueia), segundo sinal = retomar (desbloqueia).
func checkPause(jobID int64, ctx context.Context) bool {
	mu.Lock()
	ch, ok := pauseChs[jobID]
	mu.Unlock()
	if !ok {
		return false
	}
	select {
	case <-ch:
		// Pausado — aguarda retomada ou cancelamento
		select {
		case <-ch:
			return false // retomado
		case <-ctx.Done():
			return true // cancelado enquanto pausado
		}
	default:
		return false
	}
}

func registerJob(ctx context.Context, jobID int64) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan struct{}, 2)
	mu.Lock()
	jobs[jobID] = cancel
	pauseChs[jobID] = ch
	mu.Unlock()
	return ctx
}

func unregisterJob(jobID int64) {
	mu.Lock()
	delete(jobs, jobID)
	delete(pauseChs, jobID)
	mu.Unlock()
}

func workerLoop(id int, conn *sql.DB) {
	log.Printf("worker %d started", id)
	for {
		jobID, typ, _, err := db.FetchPendingJob(conn)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		log.Printf("worker %d got job %d type=%s", id, jobID, typ)
		payload, err := db.GetJobPayload(conn, jobID)
		if err != nil {
			db.UpdateJobStatus(conn, jobID, "failed", 0, err.Error())
			continue
		}

		ctx := registerJob(context.Background(), jobID)
		var jobErr error

		switch typ {
		case "copy":
			jobErr = runCopyJob(ctx, conn, jobID, payload)
		case "compress":
			jobErr = runCompressJob(ctx, conn, jobID, payload)
		case "extract":
			jobErr = runExtractJob(ctx, conn, jobID, payload)
		case "move":
			jobErr = runMoveJob(ctx, conn, jobID, payload)
		case "delete":
			jobErr = runDeleteJob(ctx, conn, jobID, payload)
		default:
			jobErr = errors.New("unsupported job type")
		}

		unregisterJob(jobID)

		if ctx.Err() != nil {
			_ = db.UpdateJobStatus(conn, jobID, "failed", 0, "cancelled")
			continue
		}
		if jobErr != nil {
			_ = db.UpdateJobStatus(conn, jobID, "failed", 0, jobErr.Error())
			continue
		}
		_ = db.UpdateJobStatus(conn, jobID, "done", 1.0, "")
	}
}

func runCopyJob(ctx context.Context, conn *sql.DB, jobID int64, payload map[string]any) error {
	sources, err := getStringSlice(payload, "sources")
	if err != nil {
		return err
	}
	dest, err := getString(payload, "dest")
	if err != nil {
		return err
	}
	conflict, _ := getString(payload, "conflict")
	if conflict == "" {
		conflict = "replace"
	}
	if len(sources) == 0 {
		return errors.New("copy job requires at least one source")
	}

	for idx, src := range sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if checkPause(jobID, ctx) {
			return ctx.Err()
		}
		fileName := filepath.Base(src)
		dstPath := filepath.Join(dest, fileName)

		if _, err := os.Stat(dstPath); err == nil {
			switch conflict {
			case "skip":
				_ = db.UpdateJobStatus(conn, jobID, "running", float64(idx+1)/float64(len(sources)), "")
				continue
			case "rename":
				dstPath = renamePath(dstPath)
			}
		}

		if err := copyimpl.CopyFile(src, dest); err != nil {
			return err
		}
		_ = db.UpdateJobStatus(conn, jobID, "running", float64(idx+1)/float64(len(sources)), "")
	}
	return nil
}

func runCompressJob(ctx context.Context, conn *sql.DB, jobID int64, payload map[string]any) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	sources, err := getStringSlice(payload, "sources")
	if err != nil {
		return err
	}
	dest, err := getString(payload, "dest")
	if err != nil {
		return err
	}
	conflict, _ := getString(payload, "conflict")
	if conflict == "" {
		conflict = "replace"
	}
	if len(sources) == 0 {
		return errors.New("compress job requires at least one source")
	}

	if _, statErr := os.Stat(dest); statErr == nil {
		switch conflict {
		case "skip":
			_ = db.UpdateJobStatus(conn, jobID, "running", 1.0, "")
			return nil
		case "rename":
			dest = renamePath(dest)
		}
	}

	_ = db.UpdateJobStatus(conn, jobID, "running", 0.5, "")
	if err := compress.CompressMany(sources, dest); err != nil {
		return err
	}
	return nil
}

func runExtractJob(ctx context.Context, conn *sql.DB, jobID int64, payload map[string]any) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	archive, err := getString(payload, "archive")
	if err != nil {
		return err
	}
	dest, err := getString(payload, "dest")
	if err != nil {
		return err
	}
	conflict, _ := getString(payload, "conflict")
	if conflict == "" {
		conflict = "replace"
	}

	if _, statErr := os.Stat(dest); statErr == nil {
		switch conflict {
		case "skip":
			_ = db.UpdateJobStatus(conn, jobID, "running", 1.0, "")
			return nil
		case "rename":
			dest = renamePath(dest)
		}
	}

	_ = db.UpdateJobStatus(conn, jobID, "running", 0.5, "")
	if err := compress.Extract(archive, dest); err != nil {
		return err
	}
	return nil
}

func getString(m map[string]any, key string) (string, error) {
	raw, ok := m[key]
	if !ok {
		return "", fmt.Errorf("missing %s", key)
	}
	val, ok := raw.(string)
	if !ok || val == "" {
		return "", fmt.Errorf("invalid %s", key)
	}
	return val, nil
}

func getStringSlice(m map[string]any, key string) ([]string, error) {
	raw, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("missing %s", key)
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid %s", key)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		s, ok := item.(string)
		if !ok || s == "" {
			return nil, fmt.Errorf("invalid %s item", key)
		}
		out = append(out, s)
	}
	return out, nil
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

func runMoveJob(ctx context.Context, conn *sql.DB, jobID int64, payload map[string]any) error {
	sources, err := getStringSlice(payload, "sources")
	if err != nil {
		return err
	}
	dest, err := getString(payload, "dest")
	if err != nil {
		return err
	}
	conflict, _ := getString(payload, "conflict")
	if conflict == "" {
		conflict = "replace"
	}
	if len(sources) == 0 {
		return errors.New("move job requires at least one source")
	}

	for idx, src := range sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if checkPause(jobID, ctx) {
			return ctx.Err()
		}
		fileName := filepath.Base(src)
		dstPath := filepath.Join(dest, fileName)

		if _, err := os.Stat(dstPath); err == nil {
			switch conflict {
			case "skip":
				_ = db.UpdateJobStatus(conn, jobID, "running", float64(idx+1)/float64(len(sources)), "")
				continue
			case "rename":
				dstPath = renamePath(dstPath)
			}
		}

		if err := os.Rename(src, dstPath); err != nil {
			if err := copyFileSimple(src, dstPath); err != nil {
				return err
			}
			os.Remove(src)
		}
		_ = db.UpdateJobStatus(conn, jobID, "running", float64(idx+1)/float64(len(sources)), "")
	}
	return nil
}

func runDeleteJob(ctx context.Context, conn *sql.DB, jobID int64, payload map[string]any) error {
	sources, err := getStringSlice(payload, "sources")
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return errors.New("delete job requires at least one source")
	}

	for idx, src := range sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		info, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if info.IsDir() {
			err = os.RemoveAll(src)
		} else {
			err = os.Remove(src)
		}
		if err != nil {
			return err
		}
		_ = db.UpdateJobStatus(conn, jobID, "running", float64(idx+1)/float64(len(sources)), "")
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
