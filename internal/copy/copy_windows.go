//go:build windows
package copy

import (
	"io"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

// CopyFile uses Windows API copy with buffered fallback.
func CopyFile(src, destDir string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	dstPath := filepath.Join(destDir, srcInfo.Name())

	srcPtr, err := windows.UTF16PtrFromString(src)
	if err != nil {
		return err
	}
	dstPtr, err := windows.UTF16PtrFromString(dstPath)
	if err != nil {
		return err
	}

	if err := copyFileW(srcPtr, dstPtr); err == nil {
		_ = os.Chtimes(dstPath, srcInfo.ModTime(), srcInfo.ModTime())
		return nil
	}

	// Fallback to buffered copy.
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

func copyFileW(src, dst *uint16) error {
	k32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := k32.NewProc("CopyFileW")
	r1, _, e1 := proc.Call(
		uintptr(unsafe.Pointer(src)),
		uintptr(unsafe.Pointer(dst)),
		uintptr(0), // failIfExists = FALSE
	)
	if r1 == 0 {
		if e1 != windows.ERROR_SUCCESS {
			return e1
		}
		return windows.GetLastError()
	}
	return nil
}
