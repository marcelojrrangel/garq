//go:build !sevenzip_sdk

package compress

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func Compress(src, dest string) error {
	return CompressMany([]string{src}, dest)
}

func CompressMany(sources []string, dest string) error {
	return CompressManyCLI(sources, dest)
}

func CompressManyCLI(sources []string, dest string) error {
	if len(sources) == 0 {
		return errors.New("sources cannot be empty")
	}
	args := []string{"a", "-t7z", "-mx=5", dest}
	args = append(args, sources...)
	return run7z(args...)
}

func Extract(archive, dest string) error {
	return run7z("x", archive, "-o"+dest, "-y")
}

func run7z(args ...string) error {
	bin, err := find7zExecutable()
	if err != nil {
		return err
	}
	cmd := exec.Command(bin, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("7z failed: %s", stderr.String())
		}
		return err
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
