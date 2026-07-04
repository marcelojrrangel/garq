//go:build !windows && !linux
package copy

import (
	"io"
	"os"
	"path/filepath"
)

// CopyFile uses buffered fallback on non-Windows/non-Linux platforms.
func CopyFile(src, destDir string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	dstPath := filepath.Join(destDir, srcInfo.Name())
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	buf := make([]byte, 1024*1024)
	if _, err := io.CopyBuffer(out, in, buf); err != nil {
		return err
	}
	_ = os.Chtimes(dstPath, srcInfo.ModTime(), srcInfo.ModTime())
	return out.Sync()
}
