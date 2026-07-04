package copy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	content := []byte("hello world test file")
	srcPath := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if err := CopyFile(srcPath, dstDir); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	dstPath := filepath.Join(dstDir, "test.txt")
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read result failed: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestCopyFilePreservesModTime(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "file.txt")
	if err := os.WriteFile(srcPath, []byte("data"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	srcInfo, _ := os.Stat(srcPath)

	if err := CopyFile(srcPath, dstDir); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	dstPath := filepath.Join(dstDir, "file.txt")
	dstInfo, _ := os.Stat(dstPath)
	if !srcInfo.ModTime().Equal(dstInfo.ModTime()) {
		t.Errorf("mod time mismatch: src=%v dst=%v", srcInfo.ModTime(), dstInfo.ModTime())
	}
}

func TestCopyFileOverwrites(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	dstPath := filepath.Join(dstDir, "file.txt")
	os.WriteFile(dstPath, []byte("old"), 0644)

	srcPath := filepath.Join(srcDir, "file.txt")
	os.WriteFile(srcPath, []byte("new"), 0644)

	if err := CopyFile(srcPath, dstDir); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	got, _ := os.ReadFile(dstPath)
	if string(got) != "new" {
		t.Errorf("expected 'new', got %q", got)
	}
}

func TestCopyFileNotFound(t *testing.T) {
	dstDir := t.TempDir()
	err := CopyFile("/nonexistent/file.txt", dstDir)
	if err == nil {
		t.Fatal("expected error for nonexistent source")
	}
}
