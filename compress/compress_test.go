package compress

import (
	"os"
	"path/filepath"
	"testing"
)

func findTest7z(t *testing.T) string {
	t.Helper()
	if custom := os.Getenv("GARQ_7Z_PATH"); custom != "" {
		if _, err := os.Stat(custom); err == nil {
			return custom
		}
	}
	candidates := []string{
		"../7z/7z.exe",
		"../7z/7zz",
		"../../7z/7z.exe",
		"../../7z/7zz",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			t.Setenv("GARQ_7Z_PATH", abs)
			return abs
		}
	}
	t.Skip("7z not found, skipping compress tests")
	return ""
}

func TestCompressAndExtract(t *testing.T) {
	_ = findTest7z(t)
	srcDir := t.TempDir()
	outDir := t.TempDir()
	extractDir := t.TempDir()

	files := map[string]string{
		"a.txt": "content A",
		"b.txt": "content B",
	}
	for name, content := range files {
		os.WriteFile(filepath.Join(srcDir, name), []byte(content), 0644)
	}

	archive := filepath.Join(outDir, "test.7z")
	sources := []string{
		filepath.Join(srcDir, "a.txt"),
		filepath.Join(srcDir, "b.txt"),
	}

	if err := CompressMany(sources, archive); err != nil {
		t.Fatalf("CompressMany failed: %v", err)
	}

	if _, err := os.Stat(archive); os.IsNotExist(err) {
		t.Fatal("archive was not created")
	}

	if err := Extract(archive, extractDir); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	for name, expected := range files {
		got, err := os.ReadFile(filepath.Join(extractDir, name))
		if err != nil {
			t.Fatalf("read extracted file %s failed: %v", name, err)
		}
		if string(got) != expected {
			t.Errorf("file %s: got %q, want %q", name, got, expected)
		}
	}
}

func TestCompressEmpty(t *testing.T) {
	outDir := t.TempDir()
	archive := filepath.Join(outDir, "empty.7z")

	err := CompressMany([]string{}, archive)
	if err == nil {
		t.Fatal("expected error for empty sources")
	}
}

func TestCompressSingle(t *testing.T) {
	_ = findTest7z(t)
	srcDir := t.TempDir()
	outDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "single.txt")
	os.WriteFile(srcPath, []byte("single file"), 0644)

	archive := filepath.Join(outDir, "single.7z")
	if err := Compress(srcPath, archive); err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if _, err := os.Stat(archive); os.IsNotExist(err) {
		t.Fatal("archive was not created")
	}
}

func TestExtractToNewDir(t *testing.T) {
	_ = findTest7z(t)
	srcDir := t.TempDir()
	outDir := t.TempDir()
	extractDir := t.TempDir() + "/nested"

	srcPath := filepath.Join(srcDir, "file.txt")
	os.WriteFile(srcPath, []byte("test"), 0644)

	archive := filepath.Join(outDir, "test.7z")
	CompressMany([]string{srcPath}, archive)

	if err := Extract(archive, extractDir); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(extractDir, "file.txt"))
	if err != nil {
		t.Fatalf("read extracted file failed: %v", err)
	}
	if string(got) != "test" {
		t.Errorf("got %q, want %q", got, "test")
	}
}
