//go:build !sevenzip_sdk

package compress

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type ProgressFunc func(progress float64)

func Compress(src, dest string) error {
	return CompressMany([]string{src}, dest)
}

func CompressMany(sources []string, dest string) error {
	return CompressManyCtx(context.Background(), sources, dest, nil)
}

func CompressManyCtx(ctx context.Context, sources []string, dest string, progressCb ProgressFunc) error {
	if len(sources) == 0 {
		return errors.New("sources cannot be empty")
	}
	args := []string{"a", "-t7z", "-mx=5", "-bsp1", dest}
	args = append(args, sources...)
	return run7zCtx(ctx, progressCb, args...)
}

func Extract(archive, dest string) error {
	return ExtractCtx(context.Background(), archive, dest, nil)
}

func ExtractCtx(ctx context.Context, archive, dest string, progressCb ProgressFunc) error {
	return run7zCtx(ctx, progressCb, "x", "-bsp1", archive, "-o"+dest, "-y")
}

func parseProgress(line string) (float64, bool) {
	if !strings.Contains(line, "%") {
		return 0, false
	}
	re := regexp.MustCompile(`(\d{1,3})%`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		pct, err := strconv.Atoi(matches[1])
		if err == nil && pct >= 0 && pct <= 100 {
			return float64(pct) / 100.0, true
		}
	}
	return 0, false
}

func scanLinesWithCarriageReturn(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func streamProgress(rc io.ReadCloser, cb ProgressFunc, ctx context.Context) {
	scanner := bufio.NewScanner(rc)
	scanner.Split(scanLinesWithCarriageReturn)
	for scanner.Scan() {
		line := scanner.Text()
		if pct, ok := parseProgress(line); ok && cb != nil {
			cb(pct)
		}
	}
}

func run7zCtx(ctx context.Context, progressCb ProgressFunc, args ...string) error {
	bin, err := find7zExecutable()
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, bin, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("7z start: %w", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		streamProgress(stdout, progressCb, ctx)
	}()

	waitErr := cmd.Wait()
	<-done

	if waitErr != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if stderr.Len() > 0 {
			return fmt.Errorf("7z failed: %s", stderr.String())
		}
		return waitErr
	}

	if progressCb != nil {
		progressCb(1.0)
	}
	return nil
}

func find7zExecutable() (string, error) {
	if custom := os.Getenv("GARQ_7Z_PATH"); custom != "" {
		if _, err := os.Stat(custom); err == nil {
			return custom, nil
		}
	}

	// Prefer a bundled 7z binary near the app to avoid requiring user installation.
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		localCandidates := []string{
			filepath.Join(exeDir, "7z", "7z.exe"),
			filepath.Join(exeDir, "tools", "7z", "7z.exe"),
			filepath.Join(exeDir, "bin", "7z", "7z.exe"),
			filepath.Join(exeDir, "7z", "7zz"),
			filepath.Join(exeDir, "tools", "7z", "7zz"),
			filepath.Join(exeDir, "bin", "7z", "7zz"),
		}
		for _, p := range localCandidates {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}

	// Also check relative to current working directory (useful during development/go run/wails dev)
	if wd, err := os.Getwd(); err == nil {
		wdCandidates := []string{
			filepath.Join(wd, "7z", "7z.exe"),
			filepath.Join(wd, "tools", "7z", "7z.exe"),
			filepath.Join(wd, "bin", "7z", "7z.exe"),
			filepath.Join(wd, "7z", "7zz"),
			filepath.Join(wd, "tools", "7z", "7zz"),
			filepath.Join(wd, "bin", "7z", "7zz"),
		}
		for _, p := range wdCandidates {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}

	if bin, err := exec.LookPath("7z"); err == nil {
		return bin, nil
	}

	if runtime.GOOS == "windows" {
		candidates := []string{
			`C:\Program Files\7-Zip\7z.exe`,
			`C:\Program Files (x86)\7-Zip\7z.exe`,
		}

		if pf := os.Getenv("ProgramFiles"); pf != "" {
			candidates = append([]string{filepath.Join(pf, "7-Zip", "7z.exe")}, candidates...)
		}
		if pfx86 := os.Getenv("ProgramFiles(x86)"); pfx86 != "" {
			candidates = append([]string{filepath.Join(pfx86, "7-Zip", "7z.exe")}, candidates...)
		}

		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}

	return "", errors.New(`7z not found. Install 7-Zip or set GARQ_7Z_PATH with full executable path`)
}
