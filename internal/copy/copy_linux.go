//go:build linux
package copy

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

// CopyFile uses copy_file_range on Linux with buffered fallback.
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

	remaining := srcInfo.Size()
	for remaining > 0 {
		n, err := unix.CopyFileRange(int(in.Fd()), nil, int(out.Fd()), nil, int(remaining), 0)
		if err != nil {
			// Cross-device, unsupported kernel/fs or invalid case: fallback.
			if errors.Is(err, unix.EXDEV) || errors.Is(err, unix.ENOSYS) || errors.Is(err, unix.EINVAL) || errors.Is(err, unix.EOPNOTSUPP) {
				return bufferedFallback(in, out, dstPath, srcInfo.ModTime())
			}
			return err
		}
		if n == 0 {
			break
		}
		remaining -= int64(n)
	}

	_ = os.Chtimes(dstPath, srcInfo.ModTime(), srcInfo.ModTime())
	return out.Sync()
}

func bufferedFallback(in *os.File, out *os.File, dstPath string, modTime time.Time) error {
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := out.Truncate(0); err != nil {
		return err
	}
	if _, err := out.Seek(0, io.SeekStart); err != nil {
		return err
	}
	buf := make([]byte, 1024*1024)
	if _, err := io.CopyBuffer(out, in, buf); err != nil {
		return err
	}
	_ = os.Chtimes(dstPath, modTime, modTime)
	return out.Sync()
}
